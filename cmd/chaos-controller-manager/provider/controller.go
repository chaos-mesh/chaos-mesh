// Copyright 2021 Chaos Mesh Authors.
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

package provider

import (
	"context"
	"math"
	"net"
	"strconv"

	"github.com/go-logr/logr"
	lru "github.com/hashicorp/golang-lru"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = v1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

// NewScheme returns the runtime.Scheme used by controller-runtime
func NewScheme() *runtime.Scheme {
	return scheme
}

// NewOption returns the manager.Options for build the controller-runtime Manager
func NewOption(logger logr.Logger, scheme *runtime.Scheme) *ctrl.Options {
	setupLog := logger.WithName("setup")

	leaderElectionNamespace := config.ControllerCfg.Namespace
	if len(leaderElectionNamespace) == 0 {
		leaderElectionNamespace = "default"
	}
	options := ctrl.Options{
		// TODO: accept the schema from parameter instead of using scheme directly
		Scheme:                     scheme,
		MetricsBindAddress:         net.JoinHostPort(config.ControllerCfg.MetricsHost, strconv.Itoa(config.ControllerCfg.MetricsPort)),
		LeaderElection:             config.ControllerCfg.EnableLeaderElection,
		LeaderElectionNamespace:    leaderElectionNamespace,
		LeaderElectionResourceLock: "configmapsleases",
		LeaderElectionID:           "chaos-mesh",
		LeaseDuration:              &config.ControllerCfg.LeaderElectLeaseDuration,
		RetryPeriod:                &config.ControllerCfg.LeaderElectRetryPeriod,
		RenewDeadline:              &config.ControllerCfg.LeaderElectRenewDeadline,
		Port:                       config.ControllerCfg.WebhookPort,
		Host:                       config.ControllerCfg.WebhookHost,
		// Don't aggregate events
		EventBroadcaster: record.NewBroadcasterWithCorrelatorOptions(record.CorrelatorOptions{
			MaxEvents:            math.MaxInt32,
			MaxIntervalInSeconds: 1,
		}),
	}

	if config.ControllerCfg.ClusterScoped {
		setupLog.Info("Chaos controller manager is running in cluster scoped mode.")
		// will not specific a certain namespace
	} else {
		setupLog.Info("Chaos controller manager is running in namespace scoped mode.", "targetNamespace", config.ControllerCfg.TargetNamespace)
		options.Namespace = config.ControllerCfg.TargetNamespace
	}

	return &options
}

// NewConfig would fetch the rest.Config from environment. When it failed to fetch config, it would exit the whole application.
func NewConfig() *rest.Config {
	return ctrl.GetConfigOrDie()
}

// NewManager would build the controller-runtime manager with the given parameters.
func NewManager(options *ctrl.Options, cfg *rest.Config) (ctrl.Manager, error) {
	if config.ControllerCfg.QPS > 0 {
		cfg.QPS = config.ControllerCfg.QPS
	}
	if config.ControllerCfg.Burst > 0 {
		cfg.Burst = config.ControllerCfg.Burst
	}

	return ctrl.NewManager(cfg, *options)
}

// NewAuthCli would build the authorizationv1.AuthorizationV1Client with given parameters.
func NewAuthCli(cfg *rest.Config) (*authorizationv1.AuthorizationV1Client, error) {

	if config.ControllerCfg.QPS > 0 {
		cfg.QPS = config.ControllerCfg.QPS
	}
	if config.ControllerCfg.Burst > 0 {
		cfg.Burst = config.ControllerCfg.Burst
	}

	return authorizationv1.NewForConfig(cfg)
}

// NewClient would build the controller-runtime client.Client with given parameters.
func NewClient(mgr ctrl.Manager, scheme *runtime.Scheme) (client.Client, error) {
	// TODO: make this size configurable
	cache, err := lru.New(100)
	if err != nil {
		return nil, err
	}
	return &UpdatedClient{
		client: mgr.GetClient(),
		scheme: scheme,
		cache:  cache,
	}, nil
}

type noCacheReader struct {
	fx.Out

	client.Reader `name:"no-cache"`
}

// NewNoCacheReader builds a client.Reader with no cache.
// TODO: we could return with fx.Annotate instead of struct noCacheReader and magic name "no-cache"
func NewNoCacheReader(mgr ctrl.Manager) noCacheReader {
	return noCacheReader{
		Reader: mgr.GetAPIReader(),
	}
}

type controlPlaneCacheReader struct {
	fx.Out

	client.Reader `name:"control-plane-cache"`
}

// NewControlPlaneCacheReader builds a client.Reader with cache for certain usage for control plane
func NewControlPlaneCacheReader(logger logr.Logger, cfg *rest.Config) (controlPlaneCacheReader, error) {
	mapper, err := apiutil.NewDynamicRESTMapper(cfg)
	if err != nil {
		return controlPlaneCacheReader{}, err
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	// Create the cache for the cached read client and registering informers
	cacheReader, err := cache.New(cfg, cache.Options{Scheme: scheme, Mapper: mapper, Resync: nil, Namespace: config.ControllerCfg.Namespace})
	if err != nil {
		return controlPlaneCacheReader{}, err
	}
	// TODO: store the channel and use it to stop
	// FIXME: goroutine leaks
	go func() {
		// FIXME: get context from parameter
		err := cacheReader.Start(context.TODO())
		if err != nil {
			logger.Error(err, "fail to start cached client")
		}
	}()

	c, err := client.New(cfg, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return controlPlaneCacheReader{}, err
	}

	cachedClient, err := client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader:       cacheReader,
		Client:            c,
		UncachedObjects:   nil,
		CacheUnstructured: false,
	})
	if err != nil {
		return controlPlaneCacheReader{}, err
	}

	return controlPlaneCacheReader{
		Reader: cachedClient,
	}, nil
}

func NewClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(config)
}

// Module would provide objects to fx for dependency injection.
var Module = fx.Provide(
	NewOption,
	NewClient,
	NewClientSet,
	NewManager,
	NewAuthCli,
	NewScheme,
	NewConfig,
	NewNoCacheReader,
	NewControlPlaneCacheReader,
)
