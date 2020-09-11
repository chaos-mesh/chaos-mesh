// Copyright 2020 Chaos Mesh Authors.
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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/core"

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
	conf *config.ChaosDashboardConfig,
	archive core.ExperimentStore,
	event core.EventStore,
) (*Server, client.Client, client.Reader) {
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

	for kind, chaosKind := range v1alpha1.AllKinds() {
		if err = (&ChaosCollector{
			Client:  s.Mgr.GetClient(),
			Log:     ctrl.Log.WithName("collector").WithName(kind),
			archive: archive,
			event:   event,
		}).Setup(s.Mgr, chaosKind.Chaos); err != nil {
			log.Error(err, "unable to create collector", "collector", kind)
			os.Exit(1)
		}
	}

	return s, s.Mgr.GetClient(), s.Mgr.GetAPIReader()
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
