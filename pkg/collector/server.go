// Copyright 2020 PingCAP, Inc.
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

package collector

import (
	"os"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
	log    = ctrl.Log.WithName("collector")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v1alpha1.AddToScheme(scheme)
}

// Server defines a server to manage collectors.
type Server struct {
	Mgr ctrl.Manager
}

// NewServer returns a CollectorServer and Client.
func NewServer(
	conf *config.ChaosServerConfig,
	archive core.ArchiveStore,
	event core.EventStore,
) (*Server, client.Client) {
	var err error
	s := &Server{}
	s.Mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: conf.MetricAddress,
		LeaderElection:     conf.EnableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		log.Error(err, "unable to start collector")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client:  s.Mgr.GetClient(),
		Log:     ctrl.Log.WithName("collector").WithName("PodChaos"),
		archive: archive,
		event:   event,
	}).Setup(s.Mgr, &v1alpha1.PodChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "PodChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client:  s.Mgr.GetClient(),
		Log:     ctrl.Log.WithName("collector").WithName("NetworkChaos"),
		archive: archive,
		event:   event,
	}).Setup(s.Mgr, &v1alpha1.NetworkChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "NetworkChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client:  s.Mgr.GetClient(),
		Log:     ctrl.Log.WithName("collector").WithName("IoChaos"),
		archive: archive,
		event:   event,
	}).Setup(s.Mgr, &v1alpha1.IoChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "IoChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client:  s.Mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("TimeChaos"),
		archive: archive,
		event:   event,
	}).Setup(s.Mgr, &v1alpha1.TimeChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "TimeChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client:  s.Mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("KernelChaos"),
		archive: archive,
		event:   event,
	}).Setup(s.Mgr, &v1alpha1.KernelChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "KernelChaos")
		os.Exit(1)
	}
	return s, s.Mgr.GetClient()
}

// Register starts collectors manager.
func Register(s *Server, stopCh <-chan struct{}) {
	go func() {
		log.Info("Starting collector")
		if err := s.Mgr.Start(stopCh); err != nil {
			log.Error(err, "problem running collector")
			os.Exit(1)
		}
	}()
}
