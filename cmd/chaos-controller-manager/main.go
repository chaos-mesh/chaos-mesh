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

package main

import (
	"flag"
	stdlog "log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"golang.org/x/time/rate"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/controllers"
	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	ctrlserver "github.com/chaos-mesh/chaos-mesh/pkg/ctrl"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/metrics"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
	apiWebhook "github.com/chaos-mesh/chaos-mesh/pkg/webhook"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

var (
	printVersion bool
	setupLog     = ctrl.Log.WithName("setup")
)

func parseFlags() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.Parse()
}

func main() {
	parseFlags()
	version.PrintVersionInfo("Controller manager")
	if printVersion {
		os.Exit(0)
	}

	rootLogger, err := log.NewDefaultZapLogger()
	if err != nil {
		stdlog.Fatal("failed to create root logger", err)
	}
	log.ReplaceGlobals(rootLogger)
	ctrl.SetLogger(rootLogger)

	// set RPCTimeout config
	grpcUtils.RPCTimeout = ccfg.ControllerCfg.RPCTimeout
	app := fx.New(
		fx.Logger(log.NewLogrPrinter(rootLogger.WithName("fx"))),
		fx.Supply(controllermetrics.Registry),
		fx.Supply(rootLogger),
		fx.Provide(metrics.NewChaosControllerManagerMetricsCollector),
		fx.Provide(ctrlserver.New),
		fx.Options(
			provider.Module,
			controllers.Module,
			selector.Module,
			types.ChaosObjects,
			types.WebhookObjects,
		),
		fx.Invoke(Run),
	)

	app.Run()
}

// RunParams contains all the parameters needed to run the chaos-controller-manager
type RunParams struct {
	fx.In
	// Mgr is the controller-runtime Manager to register controllers and webhooks to.
	Mgr ctrl.Manager
	// Logger is the root logger used in the application.
	Logger logr.Logger
	// AuthCli is the typed kubernetes authorization client. Required for the authentication webhooks.
	AuthCli *authorizationv1.AuthorizationV1Client
	// DaemonClientBuilder is the builder/factory for creating chaos daemon clients.
	DaemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder
	// MetricsCollector collects metrics for observability.
	MetricsCollector *metrics.ChaosControllerManagerMetricsCollector
	// CtrlServer is the graphql server for chaosctl.
	CtrlServer *handler.Server

	// Objs collects all the kinds of chaos custom resource objects that would be handled by the controller/reconciler.
	Objs []types.Object `group:"objs"`
	// WebhookObjs collects all the kinds of chaos custom resource objects that would be handled by the validation and mutation webhooks.
	WebhookObjs []types.WebhookObject `group:"webhookObjs"`
}

// Run is the one of the entrypoints for fx application of chaos-controller-manager. It would bootstrap the
// controller-runtime manager and register all the controllers and webhooks.
// Please notice that Run is NOT the only one entrypoint, every other functions called by fx.Invoke are also entrypoint.
func Run(params RunParams) error {
	mgr := params.Mgr
	authCli := params.AuthCli
	metricsCollector := params.MetricsCollector

	var err error
	for _, obj := range params.Objs {
		if !ccfg.ShouldStartWebhook(obj.Name) {
			continue
		}

		err = ctrl.NewWebhookManagedBy(mgr).
			For(obj.Object).
			Complete()
		if err != nil {
			return err
		}
	}

	for _, obj := range params.WebhookObjs {
		if !ccfg.ShouldStartWebhook(obj.Name) {
			continue
		}

		err = ctrl.NewWebhookManagedBy(mgr).
			For(obj.Object).
			Complete()
		if err != nil {
			return err
		}
	}

	if ccfg.ShouldStartWebhook("schedule") {
		// setup schedule webhook
		err = ctrl.NewWebhookManagedBy(mgr).
			For(&v1alpha1.Schedule{}).
			Complete()
		if err != nil {
			return err
		}
	}

	if ccfg.ShouldStartWebhook("workflow") {
		err = ctrl.NewWebhookManagedBy(mgr).
			For(&v1alpha1.Workflow{}).
			Complete()
		if err != nil {
			return err
		}
	}

	setupLog.Info("Setting up webhook server")
	hookServer := mgr.GetWebhookServer()
	hookServer.CertDir = ccfg.ControllerCfg.CertsDir
	conf := config.NewConfigWatcherConf()

	controllerRuntimeSignalHandler := ctrl.SetupSignalHandler()

	if ccfg.ControllerCfg.PprofAddr != "0" {
		go func() {
			if err := http.ListenAndServe(ccfg.ControllerCfg.PprofAddr, nil); err != nil {
				setupLog.Error(err, "unable to start pprof server")
				os.Exit(1)
			}
		}()
	}

	if ccfg.ControllerCfg.CtrlAddr != "" {
		go func() {
			mutex := http.NewServeMux()
			mutex.Handle("/", playground.Handler("GraphQL playground", "/query"))
			mutex.Handle("/query", params.CtrlServer)
			setupLog.Info("setup ctrlserver", "addr", ccfg.ControllerCfg.CtrlAddr)
			setupLog.Error(http.ListenAndServe(ccfg.ControllerCfg.CtrlAddr, mutex), "unable to start ctrlserver")
		}()
	}

	if err = ccfg.ControllerCfg.WatcherConfig.Verify(); err != nil {
		setupLog.Error(err, "invalid environment configuration")
		os.Exit(1)
	}
	configWatcher, err := watcher.New(*ccfg.ControllerCfg.WatcherConfig, metricsCollector)
	if err != nil {
		setupLog.Error(err, "unable to create config watcher")
		os.Exit(1)
	}

	go watchConfig(configWatcher, conf, controllerRuntimeSignalHandler.Done())
	hookServer.Register("/inject-v1-pod", &webhook.Admission{
		Handler: &apiWebhook.PodInjector{
			Config:        conf,
			ControllerCfg: ccfg.ControllerCfg,
			Metrics:       metricsCollector,
			Logger:        params.Logger.WithName("pod-injector"),
		}},
	)
	hookServer.Register("/validate-auth", &webhook.Admission{
		Handler: apiWebhook.NewAuthValidator(ccfg.ControllerCfg.SecurityMode, authCli,
			ccfg.ControllerCfg.ClusterScoped, ccfg.ControllerCfg.TargetNamespace, ccfg.ControllerCfg.EnableFilterNamespace,
			params.Logger.WithName("validate-auth")),
	},
	)

	setupLog.Info("Starting manager")
	if err := mgr.Start(controllerRuntimeSignalHandler); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	return nil
}

func setupWatchQueue(stopCh <-chan struct{}, configWatcher *watcher.K8sConfigMapWatcher) workqueue.Interface {
	// watch for reconciliation signals, and grab configmaps, then update the running configuration
	// for the server
	sigChan := make(chan interface{}, 10)

	queue := workqueue.NewRateLimitingQueue(&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(0.5), 1)})

	go func() {
		for {
			select {
			case <-stopCh:
				queue.ShutDown()
				return
			case <-sigChan:
				queue.AddRateLimited(struct{}{})
			}
		}
	}()

	go func() {
		for {
			setupLog.Info("Launching watcher for ConfigMaps")
			if err := configWatcher.Watch(sigChan, stopCh); err != nil {
				switch err {
				case watcher.ErrWatchChannelClosed:
					// known issue: https://github.com/kubernetes/client-go/issues/334
					setupLog.Info("watcher channel has closed, restart watcher")
				default:
					setupLog.Error(err, "unable to watch new ConfigMaps")
					os.Exit(1)
				}
			}

			select {
			case <-stopCh:
				close(sigChan)
				return
			default:
				// sleep 2 seconds to prevent excessive log due to infinite restart
				time.Sleep(2 * time.Second)
			}
		}
	}()

	return queue
}

func watchConfig(configWatcher *watcher.K8sConfigMapWatcher, cfg *config.Config, stopCh <-chan struct{}) {
	queue := setupWatchQueue(stopCh, configWatcher)

	for {
		item, shutdown := queue.Get()
		if shutdown {
			break
		}
		func() {
			defer queue.Done(item)

			setupLog.Info("Triggering ConfigMap reconciliation")
			updatedInjectionConfigs, err := configWatcher.GetInjectionConfigs()
			if err != nil {
				setupLog.Error(err, "unable to get ConfigMaps")
				return
			}

			setupLog.Info("Updating server with newly loaded configurations",
				"original configs count", len(cfg.Injections), "updated configs count", len(updatedInjectionConfigs))
			cfg.ReplaceInjectionConfigs(updatedInjectionConfigs)
			setupLog.Info("Configuration replaced")
		}()
	}
}
