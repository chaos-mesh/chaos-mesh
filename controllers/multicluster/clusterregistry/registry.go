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
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"

	fxlogr "github.com/chaos-mesh/fx-logr"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	controllerconfig "sigs.k8s.io/controller-runtime/pkg/config"
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

type appLifecycle interface {
	Start(context.Context) error
	Stop(context.Context) error
}

type remoteCluster struct {
	app appLifecycle
	client.Client
	desired   *clientcmdapi.Config
	startDone chan struct{}
	startErr  error
	gate      chan struct{}
	state     clusterState
	stopped   chan struct{}
	stopErr   error
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

	lock  *sync.Mutex
	build func(string, *rest.Config) (appLifecycle, client.Client)
}

func New(logger logr.Logger, manageClient client.Client) *RemoteClusterRegistry {
	r := &RemoteClusterRegistry{
		clusters: make(map[string]*remoteCluster),
		logger:   logger.WithName("clusterregistry"),
		client:   manageClient,

		lock: &sync.Mutex{},
	}
	r.build = func(name string, cfg *rest.Config) (appLifecycle, client.Client) { return r.newApp(name, cfg) }
	return r
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
		Scheme:     scheme,
		Controller: controllerconfig.Controller{SkipNameValidation: ptr.To(true)},
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

func (r *RemoteClusterRegistry) reserveLocked(name string) *remoteCluster {
	cluster := &remoteCluster{gate: make(chan struct{}, 1), state: clusterStarting, startDone: make(chan struct{})}
	cluster.gate <- struct{}{}
	r.clusters[name] = cluster
	return cluster
}

func (r *RemoteClusterRegistry) startReserved(name string, cluster *remoteCluster, config *rest.Config, desired *clientcmdapi.Config) {
	cluster.app, cluster.Client = r.build(name, config)
	err := cluster.app.Start(context.Background())
	r.lock.Lock()
	if err != nil {
		r.stopLocked(name, cluster)
	} else if cluster.state == clusterStarting {
		cluster.desired = desired
		cluster.state = clusterReady
	}
	r.lock.Unlock()
	<-cluster.gate
	if err != nil {
		<-cluster.stopped
		if cluster.stopErr != nil {
			err = fmt.Errorf("%w; cleanup failed: %w", err, cluster.stopErr)
		}
		err = errors.Wrapf(err, "start controller-manager of remote cluster %s", name)
	}
	cluster.startErr = err
	close(cluster.startDone)
}

// Ensure starts or replaces the remote manager when its decoded kubeconfig changes.
// A canceled wait keeps the name reserved until actual lifecycle cleanup finishes.
func (r *RemoteClusterRegistry) Ensure(ctx context.Context, name string, desired *clientcmdapi.Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if desired == nil {
		return errors.New("kubeconfig is required")
	}
	snapshot := desired.DeepCopy()
	cfg, err := clientcmd.NewDefaultClientConfig(*snapshot.DeepCopy(), nil).ClientConfig()
	if err != nil {
		return errors.Wrap(err, "get rest config from kubeconfig")
	}
	r.lock.Lock()
	if err := ctx.Err(); err != nil {
		r.lock.Unlock()
		return err
	}
	if current := r.clusters[name]; current != nil {
		if current.state != clusterReady {
			err := current.stopErr
			if err == nil {
				err = errors.Wrapf(ErrBusy, "ensure cluster: %s", name)
			}
			r.lock.Unlock()
			return err
		}
		if reflect.DeepEqual(current.desired, snapshot) {
			r.lock.Unlock()
			return nil
		}
		r.stopLocked(name, current)
		r.lock.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-current.stopped:
			if current.stopErr != nil {
				return current.stopErr
			}
		}
		r.lock.Lock()
		if err := ctx.Err(); err != nil {
			r.lock.Unlock()
			return err
		}
		if current := r.clusters[name]; current != nil {
			matches := current.state == clusterReady && reflect.DeepEqual(current.desired, snapshot)
			r.lock.Unlock()
			if matches {
				return nil
			}
			return errors.Wrapf(ErrBusy, "ensure cluster: %s", name)
		}
	}
	cluster := r.reserveLocked(name)
	r.lock.Unlock()
	go r.startReserved(name, cluster, cfg, snapshot)
	select {
	case <-ctx.Done():
	case <-cluster.startDone:
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	if err := ctx.Err(); err != nil {
		if r.clusters[name] == cluster {
			r.stopLocked(name, cluster)
		}
		return err
	}
	if cluster.startErr != nil {
		return cluster.startErr
	}
	if r.clusters[name] != cluster || cluster.state != clusterReady {
		return errors.Wrapf(ErrBusy, "ensure cluster: %s", name)
	}
	return nil
}

// Spawn starts a remote manager, rejecting an existing registration.
// Reconciliation should use Ensure to apply desired configuration changes.
func (r *RemoteClusterRegistry) Spawn(name string, config *rest.Config) error {
	r.lock.Lock()
	if existing := r.clusters[name]; existing != nil {
		err := ErrBusy
		if existing.state == clusterReady {
			err = ErrAlreadyExist
		}
		r.lock.Unlock()
		return errors.Wrapf(err, "spawn cluster: %s", name)
	}
	cluster := r.reserveLocked(name)
	r.lock.Unlock()
	go r.startReserved(name, cluster, config, nil)
	<-cluster.startDone
	return cluster.startErr
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
