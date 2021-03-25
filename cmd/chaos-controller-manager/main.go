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
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/iochaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/metrics"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/partition"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/trafficcontrol"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/containerkill"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/podfailure"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/podkill"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/timechaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/delete"
	"github.com/chaos-mesh/chaos-mesh/controllers/desiredphase"
	"github.com/chaos-mesh/chaos-mesh/controllers/podiochaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	apiWebhook "github.com/chaos-mesh/chaos-mesh/api/webhook"
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

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PodChaos{}).
		Named("podchaos-records").
		Complete(&common.Reconciler{
			Impl: podchaos.NewImpl(
				podkill.NewImpl(mgr.GetClient()),
				podfailure.NewImpl(mgr.GetClient()),
				containerkill.NewImpl(mgr.GetClient(), ctrl.Log),
			),
			Object: &v1alpha1.PodChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup PodChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PodChaos{}).
		Named("podchaos-desiredphase").
		Complete(&desiredphase.Reconciler{
			Object: &v1alpha1.PodChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup PodChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.PodChaos{}).
		Named("podchaos-delete").
		Complete(&delete.Reconciler{
			Object: &v1alpha1.PodChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup PodChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NetworkChaos{}).
		Named("networkchaos-records").
		Complete(&common.Reconciler{
			Impl: networkchaos.NewImpl(
				trafficcontrol.NewImpl(mgr.GetClient(), ctrl.Log),
				partition.NewImpl(mgr.GetClient(), ctrl.Log),
			),
			Object: &v1alpha1.NetworkChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup NetworkChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NetworkChaos{}).
		Named("networkchaos-desiredphase").
		Complete(&desiredphase.Reconciler{
			Object: &v1alpha1.NetworkChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup NetworkChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NetworkChaos{}).
		Named("networkchaos-delete").
		Complete(&delete.Reconciler{
			Object: &v1alpha1.NetworkChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup NetworkChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Named("timechaos-records").
		Complete(&common.Reconciler{
			Impl: timechaos.NewImpl(
				mgr.GetClient(), ctrl.Log,
			),
			Object: &v1alpha1.TimeChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup TimeChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Named("timechaos-desiredphase").
		Complete(&desiredphase.Reconciler{
			Object: &v1alpha1.TimeChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup TimeChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Named("timechaos-delete").
		Complete(&delete.Reconciler{
			Object: &v1alpha1.TimeChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup TimeChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.IoChaos{}).
		Named("iochaos-records").
		Complete(&common.Reconciler{
			Impl: iochaos.NewImpl(
				mgr.GetClient(), ctrl.Log,
			),
			Object: &v1alpha1.IoChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup IoChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.IoChaos{}).
		Named("iochaos-desiredphase").
		Complete(&desiredphase.Reconciler{
			Object: &v1alpha1.IoChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup IoChaos reconciler")
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.IoChaos{}).
		Named("iochaos-delete").
		Complete(&delete.Reconciler{
			Object: &v1alpha1.IoChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log:    ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup IoChaos reconciler")
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.PodChaos{}).
		Complete(); err != nil {
		setupLog.Error(err, "fail to setup PodChaos webhook")
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.NetworkChaos{}).
		Complete(); err != nil {
		setupLog.Error(err, "fail to setup NetworkChaos webhook")
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.TimeChaos{}).
		Complete(); err != nil {
		setupLog.Error(err, "fail to setup TimeChaos webhook")
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.IoChaos{}).
		Complete(); err != nil {
		setupLog.Error(err, "fail to setup TimeChaos webhook")
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
