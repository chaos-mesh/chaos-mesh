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
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/cwen0/chaos-operator/pkg/client/clientset/versioned"
	informers "github.com/cwen0/chaos-operator/pkg/client/informers/externalversions"
	"github.com/cwen0/chaos-operator/pkg/controller"
	"github.com/cwen0/chaos-operator/pkg/controller/podchaos"
	"github.com/cwen0/chaos-operator/pkg/signals"
	"github.com/cwen0/chaos-operator/pkg/version"
	"github.com/golang/glog"
	flag "github.com/spf13/pflag"

	"k8s.io/apiserver/pkg/util/logs"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	printVersion bool
	pprofPort    string
)

func init() {
	flag.BoolVarP(&printVersion, "version", "V", false, "print version information and exit")
	flag.StringVar(&pprofPort, "pprof", "", "controller manager pprof port")
	flag.DurationVar(&controller.ResyncDuration, "resync-duration", time.Duration(30*time.Second), "resync time of informer")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo()

	if printVersion {
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := rest.InClusterConfig()
	if err != nil {
		glog.Fatalf("failed to get config: %v", err)
	}

	kubeCli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("failed to get kubernetes Clientset: %v", err)
	}

	cli, err := versioned.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("failed to create Clientset: %v", err)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, controller.ResyncDuration)
	informerFactory := informers.NewSharedInformerFactory(cli, controller.ResyncDuration)

	podChaosController := podchaos.NewController(
		kubeCli, cli,
		kubeInformerFactory,
		informerFactory)

	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(stopCh)
	informerFactory.Start(stopCh)

	go podChaosController.Run(stopCh)

	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", pprofPort), nil))
}
