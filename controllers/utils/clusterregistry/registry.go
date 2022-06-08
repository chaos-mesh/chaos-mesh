// Copyright 2022 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package clusterregistry

import (
	"context"
	"os"
	"sync"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type RemoteCluster struct {
	app *fx.App

	client.Client
}

type RemoteClusterRegistry struct {
	clusters map[string]RemoteCluster
	logger   logr.Logger
	client   client.Client
	// remote controller context is the context of all remote cluster controllers
	remoteControllerContext context.Context

	*sync.Mutex
}

func New(logger logr.Logger, client client.Client, remoteControllerContext context.Context) *RemoteClusterRegistry {
	return &RemoteClusterRegistry{
		clusters:                make(map[string]RemoteCluster),
		logger:                  logger.WithName("clusterregistry"),
		client:                  client,
		remoteControllerContext: remoteControllerContext,

		Mutex: &sync.Mutex{},
	}
}

// run will start the controller manager of the remote cluster
func run(mgr ctrl.Manager, logger logr.Logger) error {
	setupLog := logger.WithName("setup")
	setupLog.Info("Starting manager")

	controllerRuntimeSignalHandler := ctrl.SetupSignalHandler()

	if err := mgr.Start(controllerRuntimeSignalHandler); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	return nil
}

func (r *RemoteClusterRegistry) Spawn(name string, config *rest.Config) error {
	r.Lock()
	defer r.Unlock()

	app := fx.New(
		fx.Logger(log.NewLogrPrinter(r.logger.WithName("fx-"+name))),
		fx.Supply(controllermetrics.Registry),
		fx.Supply(r.logger.WithName("remotecluster-"+name)),
		fx.Provide(
			provider.NewOption,
			provider.NewClient,
			provider.NewClientSet,
			provider.NewManager,
			provider.NewScheme,
		),
		fx.Invoke(run),
	)

	err := app.Start(r.remoteControllerContext)
	if err != nil {
		return errors.Wrapf(err, "start controller-manager of remote cluster %s", name)
	}

	r.clusters[name] = RemoteCluster{
		app: app,
	}

	return nil
}
