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

	chaosoperatorv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/collector"
	"github.com/pingcap/chaos-mesh/pkg/server"
	"github.com/pingcap/chaos-mesh/pkg/version"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	_ "github.com/go-sql-driver/mysql"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var (
	metricsAddr          string
	enableLeaderElection bool
	printVersion         bool
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = chaosoperatorv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func parseFlags() {
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for chaos collector. Enabling this will ensure there is only one active chaos collector.")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")

	flag.Parse()
}

func main() {
	parseFlags()

	version.PrintVersionInfo("Chaos collector")
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
		setupLog.Error(err, "unable to start collector")
		os.Exit(1)
	}

	if err = (&collector.ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("collector").WithName("PodChaos"),
	}).Setup(mgr, &chaosoperatorv1alpha1.PodChaos{}); err != nil {
		setupLog.Error(err, "unable to create collector", "collector", "PodChaos")
		os.Exit(1)
	}

	if err = (&collector.ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("collector").WithName("NetworkChaos"),
	}).Setup(mgr, &chaosoperatorv1alpha1.NetworkChaos{}); err != nil {
		setupLog.Error(err, "unable to create collector", "collector", "NetworkChaos")
		os.Exit(1)
	}

	if err = (&collector.ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("collector").WithName("IoChaos"),
	}).Setup(mgr, &chaosoperatorv1alpha1.IoChaos{}); err != nil {
		setupLog.Error(err, "unable to create collector", "collector", "IoChaos")
		os.Exit(1)
	}

	if err = (&collector.ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("TimeChaos"),
	}).Setup(mgr, &chaosoperatorv1alpha1.TimeChaos{}); err != nil {
		setupLog.Error(err, "unable to create collector", "collector", "TimeChaos")
		os.Exit(1)
	}

	stopCh := ctrl.SetupSignalHandler()

	// +kubebuilder:scaffold:builder

	go func() {
		setupLog.Info("Starting server")
		s := server.SetupServer(mgr.GetClient())
		s.Run()
	}()

	setupLog.Info("Starting collector")
	if err := mgr.Start(stopCh); err != nil {
		setupLog.Error(err, "problem running collector")
		os.Exit(1)
	}
}
