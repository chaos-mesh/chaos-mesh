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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	chaosmeshv1alpha1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	apiWebhook "github.com/chaos-mesh/chaos-mesh/api/webhook"
	"github.com/chaos-mesh/chaos-mesh/controllers"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/metrics"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")

	// EventCoalesceWindow is the window for coalescing events from ConfigMapWatcher
	EventCoalesceWindow = time.Second * 3
)

var (
	printVersion bool
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = chaosmeshv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func parseFlags() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.Parse()
}

func main() {
	parseFlags()
	version.PrintVersionInfo("Controller manager")
	if printVersion {
		os.Exit(0)
	}

	// set RPCTimeout config
	utils.RPCTimeout = common.ControllerCfg.RPCTimeout

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: common.ControllerCfg.MetricsAddr,
		LeaderElection:     common.ControllerCfg.EnableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.PodChaosReconciler{
		Client:        mgr.GetClient(),
		Reader:        mgr.GetAPIReader(),
		EventRecorder: mgr.GetEventRecorderFor("podchaos-controller"),
		Log:           ctrl.Log.WithName("controllers").WithName("PodChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PodChaos")
		os.Exit(1)
	}
	if err = (&chaosmeshv1alpha1.PodChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "PodChaos")
		os.Exit(1)
	}

	if err = (&controllers.NetworkChaosReconciler{
		Client:        mgr.GetClient(),
		Reader:        mgr.GetAPIReader(),
		EventRecorder: mgr.GetEventRecorderFor("networkchaos-controller"),
		Log:           ctrl.Log.WithName("controllers").WithName("NetworkChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkChaos")
		os.Exit(1)
	}
	if err = (&chaosmeshv1alpha1.NetworkChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "NetworkChaos")
		os.Exit(1)
	}

	if err = (&controllers.IoChaosReconciler{
		Client:        mgr.GetClient(),
		Reader:        mgr.GetAPIReader(),
		EventRecorder: mgr.GetEventRecorderFor("iochaos-controller"),
		Log:           ctrl.Log.WithName("controllers").WithName("IoChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IoChaos")
		os.Exit(1)
	}
	if err = (&chaosmeshv1alpha1.IoChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "IoChaos")
		os.Exit(1)
	}

	if err = (&controllers.TimeChaosReconciler{
		Client:        mgr.GetClient(),
		Reader:        mgr.GetAPIReader(),
		EventRecorder: mgr.GetEventRecorderFor("timechaos-controller"),
		Log:           ctrl.Log.WithName("controllers").WithName("TimeChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TimeChaos")
		os.Exit(1)
	}
	if err = (&chaosmeshv1alpha1.TimeChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "TimeChaos")
		os.Exit(1)
	}

	if err = (&controllers.KernelChaosReconciler{
		Client:        mgr.GetClient(),
		Reader:        mgr.GetAPIReader(),
		EventRecorder: mgr.GetEventRecorderFor("kernelchaos-controller"),
		Log:           ctrl.Log.WithName("controllers").WithName("KernelChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KernelChaos")
		os.Exit(1)
	}
	if err = (&chaosmeshv1alpha1.KernelChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KernelChaos")
		os.Exit(1)
	}

	if err = (&controllers.StressChaosReconciler{
		Client:        mgr.GetClient(),
		Reader:        mgr.GetAPIReader(),
		EventRecorder: mgr.GetEventRecorderFor("stresschaos-controller"),
		Log:           ctrl.Log.WithName("controllers").WithName("StressChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "StressChaos")
		os.Exit(1)
	}
	if err = (&chaosmeshv1alpha1.StressChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "StressChaos")
		os.Exit(1)
	}

	// We only setup webhook for podnetworkchaos, and the logic of applying chaos are in the validation
	// webhook, because we need to get the running result synchronously in network chaos reconciler
	v1alpha1.RegisterRawPodNetworkHandler(&podnetworkchaos.Handler{
		Client: mgr.GetClient(),
		Reader: mgr.GetAPIReader(),
		Log:    ctrl.Log.WithName("handler").WithName("PodNetworkChaos"),
	})
	if err = (&chaosmeshv1alpha1.PodNetworkChaos{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "PodNetworkChaos")
		os.Exit(1)
	}

	// Init metrics collector
	metricsCollector := metrics.NewChaosCollector(mgr.GetCache(), controllermetrics.Registry)

	setupLog.Info("Setting up webhook server")

	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = common.ControllerCfg.CertsDir
	conf := config.NewConfigWatcherConf()
	stopCh := ctrl.SetupSignalHandler()

	if common.ControllerCfg.PprofAddr != "0" {
		go func() {
			if err := http.ListenAndServe(common.ControllerCfg.PprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof server")
				os.Exit(1)
			}
		}()
	}

	if err = common.ControllerCfg.WatcherConfig.Verify(); err != nil {
		setupLog.Error(err, "invalid environment configuration")
		os.Exit(1)
	}
	configWatcher, err := watcher.New(*common.ControllerCfg.WatcherConfig, metricsCollector)
	if err != nil {
		setupLog.Error(err, "unable to create config watcher")
		os.Exit(1)
	}

	watchConfig(configWatcher, conf, stopCh)
	hookServer.Register("/inject-v1-pod", &webhook.Admission{
		Handler: &apiWebhook.PodInjector{
			Config:  conf,
			Metrics: metricsCollector,
		}},
	)

	// +kubebuilder:scaffold:builder

	setupLog.Info("Starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

}

func watchConfig(configWatcher *watcher.K8sConfigMapWatcher, cfg *config.Config, stopCh <-chan struct{}) {
	go func() {
		// watch for reconciliation signals, and grab configmaps, then update the running configuration
		// for the server
		sigChan := make(chan interface{}, 10)
		//debouncedChan := make(chan interface{}, 10)

		// debounce events from sigChan, so we dont hammer apiserver on reconciliation
		eventsCh := utils.Coalescer(EventCoalesceWindow, sigChan, stopCh)

		go func() {
			for {
				setupLog.Info("Launching watcher for ConfigMaps")
				if err := configWatcher.Watch(sigChan, stopCh); err != nil {
					switch err {
					case watcher.ErrWatchChannelClosed:
						setupLog.Error(err, "watcher got error, try to restart watcher")
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

		for {
			select {
			case <-eventsCh:
				setupLog.Info("Triggering ConfigMap reconciliation")
				updatedInjectionConfigs, err := configWatcher.GetInjectionConfigs()
				if err != nil {
					setupLog.Error(err, "unable to get ConfigMaps")
					continue
				}

				setupLog.Info("Updating server with newly loaded configurations",
					"original configs count", len(cfg.Injections), "updated configs count", len(updatedInjectionConfigs))
				cfg.ReplaceInjectionConfigs(updatedInjectionConfigs)
				setupLog.Info("Configuration replaced")
			case <-stopCh:
				break
			}
		}

	}()
}
