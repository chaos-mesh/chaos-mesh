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
	"time"

	fxlogr "github.com/chaos-mesh/fx-logr"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/multicluster/remotechaosmonitor"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
)

const remoteRequestTimeout = 10 * time.Second

type clusterState int

const (
	clusterStarting clusterState = iota
	clusterReady
	clusterStopping
	clusterStopped
)

type remoteCluster struct {
	app *fx.App
	client.Client
	gate    chan struct{}
	state   clusterState
	stopped chan struct{}
	stopErr error
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

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				defer close(stopChan)
				setupLog := logger.WithName("setup")
				setupLog.Info("Starting manager")
				if err := mgr.Start(executionCtx); err != nil {
					setupLog.Error(err, "unable to start manager")
					os.Exit(1)
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
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
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
		// TODO: enable leader election
		LeaderElection: false,
		RetryPeriod:    &config.ControllerCfg.LeaderElectRetryPeriod,
		RenewDeadline:  &config.ControllerCfg.LeaderElectRenewDeadline,
	}

	// TODO: consider the cluster scope / namespace scope with multi-cluster

	return &options
}

// WithClient runs a callback with a live remote client. Calls for the same
// cluster are serialized; busy or stopping entries return ErrBusy immediately.
// The callback must not stop or recursively borrow its own cluster.
func (r *RemoteClusterRegistry) WithClient(name string, f func(c client.Client) error) error {
	r.lock.Lock()
	cluster, ok := r.clusters[name]
	if !ok {
		r.lock.Unlock()
		return errors.Wrapf(ErrNotExist, "lookup cluster: %s", name)
	}
	if cluster.state != clusterReady {
		r.lock.Unlock()
		return errors.Wrapf(ErrBusy, "lookup cluster: %s", name)
	}
	select {
	case cluster.gate <- struct{}{}:
		r.lock.Unlock()
		defer func() { <-cluster.gate }()
		return f(cluster.Client)
	default:
		r.lock.Unlock()
		return errors.Wrapf(ErrBusy, "lookup cluster: %s", name)
	}
}

// Stop closes admission immediately and waits for active use and actual manager
// shutdown. A caller timeout does not release the name for replacement.
func (r *RemoteClusterRegistry) Stop(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.lock.Lock()
	cluster, ok := r.clusters[name]
	if !ok {
		r.lock.Unlock()
		return errors.Wrapf(ErrNotExist, "lookup cluster: %s", name)
	}
	r.stopLocked(name, cluster)
	r.lock.Unlock()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-cluster.stopped:
		return cluster.stopErr
	}
}

func (r *RemoteClusterRegistry) stopLocked(name string, cluster *remoteCluster) {
	if cluster.stopped != nil {
		return
	}
	cluster.state = clusterStopping
	cluster.stopped = make(chan struct{})
	go func() {
		cluster.gate <- struct{}{}
		defer func() { <-cluster.gate }()
		err := cluster.app.Stop(context.Background())
		r.lock.Lock()
		defer r.lock.Unlock()
		cluster.stopErr = errors.Wrapf(err, "stop fx app: %s", name)
		if err == nil {
			cluster.state = clusterStopped
			if r.clusters[name] == cluster {
				delete(r.clusters, name)
			}
			r.logger.Info("controller manager stopped", "name", name)
		}
		close(cluster.stopped)
	}()
}

func (r *RemoteClusterRegistry) spawn(name string, build func() (*fx.App, client.Client)) error {
	r.lock.Lock()
	if existing, ok := r.clusters[name]; ok {
		err := ErrBusy
		if existing.state == clusterReady {
			err = ErrAlreadyExist
		}
		r.lock.Unlock()
		return errors.Wrapf(err, "spawn cluster: %s", name)
	}
	cluster := &remoteCluster{gate: make(chan struct{}, 1), state: clusterStarting}
	cluster.gate <- struct{}{}
	r.clusters[name] = cluster
	r.lock.Unlock()

	cluster.app, cluster.Client = build()
	err := cluster.app.Start(context.Background())
	r.lock.Lock()
	if err != nil {
		r.stopLocked(name, cluster)
	} else if cluster.state == clusterStarting {
		cluster.state = clusterReady
	}
	r.lock.Unlock()
	<-cluster.gate
	if err != nil {
		<-cluster.stopped
		return errors.Wrapf(err, "start controller-manager of remote cluster %s", name)
	}
	return nil
}

// Spawn starts the controller-manager and watches the remote cluster.
func (r *RemoteClusterRegistry) Spawn(name string, config *rest.Config) error {
	return r.spawn(name, func() (*fx.App, client.Client) { return r.newApp(name, config) })
}

func forwardingConfig(config *rest.Config) *rest.Config {
	copied := rest.CopyConfig(config)
	if copied.Timeout <= 0 || copied.Timeout > remoteRequestTimeout {
		copied.Timeout = remoteRequestTimeout
	}
	return copied
}

func forwardingClient(mgr ctrl.Manager) (client.Client, error) {
	config := forwardingConfig(mgr.GetConfig())
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}
	mapper, err := apiutil.NewDynamicRESTMapper(config, httpClient)
	if err != nil {
		return nil, err
	}
	return client.New(config, client.Options{
		Scheme: mgr.GetScheme(), Mapper: mapper, HTTPClient: httpClient,
	})
}

func (r *RemoteClusterRegistry) newApp(name string, config *rest.Config) (*fx.App, client.Client) {
	remoteFxLogger := r.logger.WithName("remotecluster-fx-" + name)

	localClient := r.client
	var remoteClient struct {
		fx.In
		Client client.Client `name:"forwarding-client"`
	}
	app := fx.New(
		fx.WithLogger(fxlogr.WithLogr(&remoteFxLogger)),
		fx.Supply(controllermetrics.Registry),
		fx.Supply(r.logger.WithName("remotecluster-"+name)),
		fx.Supply(rest.CopyConfig(config)),
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
			fx.Annotate(forwardingClient, fx.ResultTags(`name:"forwarding-client"`)),
		),
		fx.Option(types.ChaosObjects),
		// more reconcilers can be listed here to add themselves to the
		// controller manager
		remotechaosmonitor.Module,
		fx.Populate(&remoteClient),
		fx.Invoke(run),
	)

	return app, remoteClient.Client
}
