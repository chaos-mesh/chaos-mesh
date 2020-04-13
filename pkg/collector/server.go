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
	"context"
	"github.com/pingcap/chaos-mesh/pkg/store"
	"os"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/pkg/config"

	"go.uber.org/fx"

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

func NewServer(lc fx.Lifecycle, conf *config.ChaosServerConfig, db *store.DB) client.Client {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
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
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("collector").WithName("PodChaos"),
		db:     db,
	}).Setup(mgr, &v1alpha1.PodChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "PodChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("collector").WithName("NetworkChaos"),
		db:     db,
	}).Setup(mgr, &v1alpha1.NetworkChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "NetworkChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("collector").WithName("IoChaos"),
		db:     db,
	}).Setup(mgr, &v1alpha1.IoChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "IoChaos")
		os.Exit(1)
	}

	if err = (&ChaosCollector{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("TimeChaos"),
		db:     db,
	}).Setup(mgr, &v1alpha1.TimeChaos{}); err != nil {
		log.Error(err, "unable to create collector", "collector", "TimeChaos")
		os.Exit(1)
	}

	lc.Append(fx.Hook{
		// To mitigate the impact of deadlocks in application startup and
		// shutdown, Fx imposes a time limit on OnStart and OnStop hooks. By
		// default, hooks have a total of 30 seconds to complete. Timeouts are
		// passed via Go's usual context.Context.
		OnStart: func(ctx context.Context) error {
			go func() {
				log.Info("Starting collector")
				if err := mgr.Start(ctx.Done()); err != nil {
					log.Error(err, "problem running collector")
					os.Exit(1)
				}
			}()
			return nil
		},
	})

	return mgr.GetClient()
}
