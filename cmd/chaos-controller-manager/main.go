// Copyright 2019 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	apiWebhook "github.com/chaos-mesh/chaos-mesh/api/webhook"
	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/metrics"
	"github.com/chaos-mesh/chaos-mesh/controllers/podiochaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"

	_ "github.com/chaos-mesh/chaos-mesh/controllers/awschaos/detachvolume"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/awschaos/ec2restart"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/awschaos/ec2stop"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/dnschaos"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/httpchaos"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/iochaos"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/jvmchaos"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/kernelchaos"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/partition"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/trafficcontrol"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/podchaos/containerkill"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/podchaos/podfailure"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/podchaos/podkill"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/stresschaos"
	_ "github.com/chaos-mesh/chaos-mesh/controllers/timechaos"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var (
	printVersion                   bool
	restConfigQPS, restConfigBurst int
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func parseFlags() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.IntVar(&restConfigQPS, "rest-config-qps", 30, "QPS of rest config.")
	flag.IntVar(&restConfigBurst, "rest-config-burst", 50, "burst of rest config.")
	flag.Parse()
}

func main() {
	parseFlags()
	version.PrintVersionInfo("Controller manager")
	if printVersion {
		os.Exit(0)
	}

	// set RPCTimeout config
	grpcUtils.RPCTimeout = ccfg.ControllerCfg.RPCTimeout

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: ccfg.ControllerCfg.MetricsAddr,
		LeaderElection:     ccfg.ControllerCfg.EnableLeaderElection,
		Port:               9443,
	}

	if ccfg.ControllerCfg.ClusterScoped {
		setupLog.Info("Chaos controller manager is running in cluster scoped mode.")
		// will not specific a certain namespace
	} else {
		setupLog.Info("Chaos controller manager is running in namespace scoped mode.", "targetNamespace", ccfg.ControllerCfg.TargetNamespace)
		options.Namespace = ccfg.ControllerCfg.TargetNamespace
	}

	cfg := ctrl.GetConfigOrDie()
	setRestConfig(cfg)
	mgr, err := ctrl.NewManager(cfg, options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	err = router.SetupWithManagerAndConfigs(mgr, ccfg.ControllerCfg)
	if err != nil {
		setupLog.Error(err, "fail to setup with manager")
		os.Exit(1)
	}

	// We only setup webhook for podiochaos, and the logic of applying chaos are in the mutation
	// webhook, because we need to get the running result synchronously in io chaos reconciler
	v1alpha1.RegisterPodIoHandler(&podiochaos.Handler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("handler").WithName("PodIOChaos"),
	})
	if err = (&v1alpha1.PodIoChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "PodIOChaos")
		os.Exit(1)
	}

	// We only setup webhook for podnetworkchaos, and the logic of applying chaos are in the validation
	// webhook, because we need to get the running result synchronously in network chaos reconciler
	v1alpha1.RegisterRawPodNetworkHandler(&podnetworkchaos.Handler{
		Client:                  mgr.GetClient(),
		Reader:                  mgr.GetAPIReader(),
		Log:                     ctrl.Log.WithName("handler").WithName("PodNetworkChaos"),
		AllowHostNetworkTesting: ccfg.ControllerCfg.AllowHostNetworkTesting,
	})
	if err = (&v1alpha1.PodNetworkChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "PodNetworkChaos")
		os.Exit(1)
	}

	// Init metrics collector
	metricsCollector := metrics.NewChaosCollector(mgr.GetCache(), controllermetrics.Registry)

	setupLog.Info("Setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = ccfg.ControllerCfg.CertsDir
	conf := config.NewConfigWatcherConf()

	stopCh := ctrl.SetupSignalHandler()

	if ccfg.ControllerCfg.PprofAddr != "0" {
		go func() {
			if err := http.ListenAndServe(ccfg.ControllerCfg.PprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof server")
				os.Exit(1)
			}
		}()
	}

	if err = ccfg.ControllerCfg.WatcherConfig.Verify(); err != nil {
		setupLog.Error(err, "invalid environment configuration")
		os.Exit(1)
	}
	configWatcher, err := watcher.New(*ccfg.ControllerCfg.WatcherConfig, metricsCollector)
	if err != nil {
		setupLog.Error(err, "unable to create config watcher")
		os.Exit(1)
	}

	go watchConfig(configWatcher, conf, stopCh)
	hookServer.Register("/inject-v1-pod", &webhook.Admission{
		Handler: &apiWebhook.PodInjector{
			Config:        conf,
			ControllerCfg: ccfg.ControllerCfg,
			Metrics:       metricsCollector,
		}},
	)

	// +kubebuilder:scaffold:builder

	setupLog.Info("Starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
}

func setRestConfig(c *rest.Config) {
	if restConfigQPS > 0 {
		c.QPS = float32(restConfigQPS)
	}
	if restConfigBurst > 0 {
		c.Burst = restConfigBurst
	}
}

func setupWatchQueue(stopCh <-chan struct{}, configWatcher *watcher.K8sConfigMapWatcher) workqueue.Interface {
	// watch for reconciliation signals, and grab configmaps, then update the running configuration
	// for the server
	sigChan := make(chan interface{}, 10)

	queue := workqueue.NewRateLimitingQueue(&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(0.5), 1)})

	go func() {
		for {
			select {
			case <-stopCh:
				queue.ShutDown()
				return
			case <-sigChan:
				queue.AddRateLimited(struct{}{})
			}
		}
	}()

	go func() {
		for {
			setupLog.Info("Launching watcher for ConfigMaps")
			if err := configWatcher.Watch(sigChan, stopCh); err != nil {
				switch err {
				case watcher.ErrWatchChannelClosed:
					// known issue: https://github.com/kubernetes/client-go/issues/334
					setupLog.Info("watcher channel has closed, restart watcher")
				default:
					setupLog.Error(err, "unable to watch new ConfigMaps")
					os.Exit(1)
				}
			}

			select {
			case <-stopCh:
				close(sigChan)
				return
			default:
				// sleep 2 seconds to prevent excessive log due to infinite restart
				time.Sleep(2 * time.Second)
			}
		}
	}()

	return queue
}

func watchConfig(configWatcher *watcher.K8sConfigMapWatcher, cfg *config.Config, stopCh <-chan struct{}) {
	queue := setupWatchQueue(stopCh, configWatcher)

	for {
		item, shutdown := queue.Get()
		if shutdown {
			break
		}
		func() {
			defer queue.Done(item)

			setupLog.Info("Triggering ConfigMap reconciliation")
			updatedInjectionConfigs, err := configWatcher.GetInjectionConfigs()
			if err != nil {
				setupLog.Error(err, "unable to get ConfigMaps")
				return
			}

			setupLog.Info("Updating server with newly loaded configurations",
				"original configs count", len(cfg.Injections), "updated configs count", len(updatedInjectionConfigs))
			cfg.ReplaceInjectionConfigs(updatedInjectionConfigs)
			setupLog.Info("Configuration replaced")
		}()
	}
}
