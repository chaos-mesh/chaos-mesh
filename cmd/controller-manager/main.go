// Copyright 2019 PingCAP, Inc.
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
	"os"
	"time"

	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
	apiWebhook "github.com/pingcap/chaos-mesh/api/webhook"
	"github.com/pingcap/chaos-mesh/controllers"
	"github.com/pingcap/chaos-mesh/pkg/flags"
	"github.com/pingcap/chaos-mesh/pkg/utils"
	"github.com/pingcap/chaos-mesh/pkg/version"
	"github.com/pingcap/chaos-mesh/pkg/webhook/config"
	"github.com/pingcap/chaos-mesh/pkg/webhook/config/watcher"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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
	metricsAddr          string
	enableLeaderElection bool
	certsDir             string
	configDir            string
	printVersion         bool

	cmWatcherLabels = flags.NewMapStringStringFlag()
	watcherConfig   = watcher.NewConfig()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = chaosmeshv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func parseFlags() {
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&certsDir, "certs", "/etc/webhook/certs",
		"The directory for storing certs key file and cert file")
	flag.StringVar(&configDir, "conf", "/etc/webhook/conf",
		"The directory for storing webhook config files")
	flag.StringVar(&watcherConfig.Namespace, "configmap-namespace", "",
		"Namespace to search for ConfigMaps to load Injection Configs from (default: current namespace)")
	flag.Var(&cmWatcherLabels, "configmap-labels",
		"Label pairs used to discover ConfigMaps in Kubernetes. These should be key1=value[,key2=val2,...]")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")

	flag.Parse()
}

func main() {
	parseFlags()

	version.PrintVersionInfo("Controller manager")
	if printVersion {
		os.Exit(0)
	}

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.PodChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("PodChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PodChaos")
		os.Exit(1)
	}

	if err = (&controllers.NetworkChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("NetworkChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NetworkChaos")
		os.Exit(1)
	}

	if err = (&controllers.IoChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("IoChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "IoChaos")
		os.Exit(1)
	}

	if err = (&controllers.TimeChaosReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("TimeChaos"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TimeChaos")
		os.Exit(1)
	}

	setupLog.Info("setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = certsDir
	webhookConfig, err := config.LoadConfigDirectory(configDir)
	if err != nil {
		setupLog.Error(err, "load webhook config error")
		os.Exit(1)
	}

	stopCh := ctrl.SetupSignalHandler()

	watchConfig(webhookConfig, stopCh)

	hookServer.Register("/inject-v1-pod", &webhook.Admission{Handler: &apiWebhook.PodInjector{
		Config: webhookConfig,
	}})

	// +kubebuilder:scaffold:builder

	setupLog.Info("Starting manager")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func watchConfig(cfg *config.Config, stopCh <-chan struct{}) {
	watcherConfig.ConfigMapLabels = cmWatcherLabels.ToMapStringString()
	// start up the watcher, and get the first batch of ConfigMaps
	// to set in the config.
	// make sure to union this with any file configs we loaded from disk
	configWatcher, err := watcher.New(*watcherConfig)
	if err != nil {
		setupLog.Error(err, "unable to create ConfigMap watchers")
		os.Exit(1)
	}
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
				err := configWatcher.Watch(sigChan, stopCh)
				if err != nil {
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
				updatedInjectionConfigs, err := configWatcher.Get()
				if err != nil {
					setupLog.Error(err, "unable to get ConfigMaps")
					continue
				}

				if len(updatedInjectionConfigs) == 0 {
					setupLog.Info("No updated injection configs")
					continue
				}

				setupLog.Info("Got updated InjectionConfigs from reconciliation",
					"updated config count", len(updatedInjectionConfigs))

				newInjectionConfigs := make([]*config.InjectionConfig, len(updatedInjectionConfigs)+len(cfg.Injections))
				{
					i := 0
					for k := range cfg.Injections {
						newInjectionConfigs[i] = cfg.Injections[k]
						i++
					}
					for i, watched := range updatedInjectionConfigs {
						newInjectionConfigs[i+len(cfg.Injections)] = watched
					}
				}

				setupLog.Info("Updating server with newly loaded configurations",
					"origin configs count", len(cfg.Injections), "updated configs count", len(updatedInjectionConfigs))
				cfg.ReplaceInjectionConfigs(newInjectionConfigs)
				setupLog.Info("Configuration replaced")
			case <-stopCh:
				break
			}
		}

	}()
}
