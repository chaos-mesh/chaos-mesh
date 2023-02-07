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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/remotechaosmonitor"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

type remoteCluster struct {
	app *fx.App

	client.Client
}

// RemoteClusterRegistry will manage all controllers running on a remote
// cluster. The construction of these controllers (managers) will be managed by
// `fx`. The main process of constructing a controller manage is nearly the same
// with the main one. The only difference is that we'll need to provide a new
// `RestConfig`, and `Populate` the client to allow others to use its client.
type RemoteClusterRegistry struct {
	clusters map[string]*remoteCluster
	logger   logr.Logger
	client   client.Client

	lock *sync.Mutex
}

func New(logger logr.Logger, client client.Client) *RemoteClusterRegistry {
	return &RemoteClusterRegistry{
		clusters: make(map[string]*remoteCluster),
		logger:   logger.WithName("clusterregistry"),
		client:   client,

		lock: &sync.Mutex{},
	}
}

// run will start the controller manager of the remote cluster
func run(lc fx.Lifecycle, mgr ctrl.Manager, logger logr.Logger) error {
	// TODO: use the global signal context with a cancel function
	executionCtx, cancel := context.WithCancel(context.TODO())
	stopChan := make(chan struct{})

	go func() {
		setupLog := logger.WithName("setup")
		setupLog.Info("Starting manager")

		if err := mgr.Start(executionCtx); err != nil {
			setupLog.Error(err, "unable to start manager")
			os.Exit(1)
		}

		stopChan <- struct{}{}
	}()

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cancel()
			<-stopChan
			return nil
		},
	})

	return nil
}

// TODO: unify this option with global provider.NewOption
func controllerManagerOption(scheme *runtime.Scheme) *ctrl.Options {
	options := ctrl.Options{
		// TODO: accept the schema from parameter instead of using scheme directly
		Scheme:             scheme,
		MetricsBindAddress: "0",
		// TODO: enable leader election
		LeaderElection: false,
		RetryPeriod:    &config.ControllerCfg.LeaderElectRetryPeriod,
		RenewDeadline:  &config.ControllerCfg.LeaderElectRenewDeadline,
	}

	// TODO: consider the cluster scope / namespace scope with multi-cluster

	return &options
}

// WithClient enables developer getting a client of remote cluster to operate
// inside the remote cluster.
//
// TODO: add more kinds of client, like `no-cache` into this registry, if they
// are needed
func (r *RemoteClusterRegistry) WithClient(name string, f func(c client.Client) error) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	cluster, ok := r.clusters[name]
	if !ok {
		return errors.Wrapf(ErrNotExist, "lookup cluster: %s", name)
	}

	return f(cluster.Client)
}

// Stop stops the running controller-manager which watches the remote cluster.
func (r *RemoteClusterRegistry) Stop(ctx context.Context, name string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	cluster, ok := r.clusters[name]
	if !ok {
		return errors.Wrapf(ErrNotExist, "lookup cluster: %s", name)
	}

	err := cluster.app.Stop(ctx)
	if err != nil {
		return errors.Wrapf(err, "stop fx app: %s", name)
	}
	delete(r.clusters, name)

	r.logger.Info("controller manager stopped", "name", name)

	return nil
}

// Spawn starts the controller-manager and watches the remote cluster
func (r *RemoteClusterRegistry) Spawn(name string, config *rest.Config) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.clusters[name]; ok {
		return errors.Wrapf(ErrAlreadyExist, "spawn cluster: %s", name)
	}

	localClient := r.client
	var remoteClient client.Client
	app := fx.New(
		fx.Logger(log.NewLogrPrinter(r.logger.WithName("remotecluster-fx-"+name))),
		fx.Supply(controllermetrics.Registry),
		fx.Supply(r.logger.WithName("remotecluster-"+name)),
		fx.Supply(config),
		fx.Supply(fx.Annotated{
			Name:   "cluster-name",
			Target: name,
		}),
		fx.Provide(
			fx.Annotate(func() client.Client {
				return localClient
			}, fx.ResultTags(`name:"manage-client"`)),
		),
		fx.Provide(
			controllerManagerOption,
			provider.NewClient,
			provider.NewManager,
			provider.NewScheme,
		),
		fx.Option(types.ChaosObjects),
		// more reconcilers can be listed here to add themselves to the
		// controller manager
		remotechaosmonitor.Module,
		fx.Populate(&remoteClient),
		fx.Invoke(run),
	)

	err := app.Start(context.TODO())
	if err != nil {
		return errors.Wrapf(err, "start controller-manager of remote cluster %s", name)
	}

	r.clusters[name] = &remoteCluster{
		app:    app,
		Client: remoteClient,
	}

	return nil
}
