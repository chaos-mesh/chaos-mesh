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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/podchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/podchaos/containerkill"
	"github.com/chaos-mesh/chaos-mesh/controllers/podchaos/podfailure"
	"github.com/chaos-mesh/chaos-mesh/controllers/podchaos/podkill"
	"github.com/chaos-mesh/chaos-mesh/controllers/podiochaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"net/http"
	_ "net/http/pprof"
	"os"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

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
		Complete(&common.Reconciler{
			Impl: podchaos.NewImpl(
				podkill.NewImpl(mgr.GetClient()),
				podfailure.NewImpl(mgr.GetClient()),
				containerkill.NewImpl(mgr.GetClient(), ctrl.Log),
				),
			Object: &v1alpha1.PodChaos{},
			Client: mgr.GetClient(),
			Reader: mgr.GetAPIReader(),
			Log: ctrl.Log,
		}); err != nil {
		setupLog.Error(err, "fail to setup PodChaos reconciler")
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.PodChaos{}).
		Complete(); err != nil {
		setupLog.Error(err, "fail to setup PodChaos webhook")
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

	setupLog.Info("Setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = ccfg.ControllerCfg.CertsDir

	stopCh := ctrl.SetupSignalHandler()

	if ccfg.ControllerCfg.PprofAddr != "0" {
		go func() {
			if err := http.ListenAndServe(ccfg.ControllerCfg.PprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof server")
				os.Exit(1)
			}
		}()
	}

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

