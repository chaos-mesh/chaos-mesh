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

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	fxlogr "github.com/chaos-mesh/fx-logr"
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
)

var (
	printVersion bool

	// TODO: create the logger through dependency injection
	setupLog = ctrl.Log.WithName("setup")
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
	fxLogger := rootLogger.WithName("fx")

	// set RPCTimeout config
	grpcUtils.RPCTimeout = ccfg.ControllerCfg.RPCTimeout
	app := fx.New(
		fx.WithLogger(fxlogr.WithLogr(&fxLogger)),
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

	hookServer.Register("/validate-auth", &webhook.Admission{
		Handler: apiWebhook.NewAuthValidator(ccfg.ControllerCfg.SecurityMode, authCli, mgr.GetScheme(),
			ccfg.ControllerCfg.ClusterScoped, ccfg.ControllerCfg.TargetNamespace, ccfg.ControllerCfg.EnableFilterNamespace,
			params.Logger.WithName("validate-auth"),
		),
	},
	)

	setupLog.Info("Starting manager")
	if err := mgr.Start(controllerRuntimeSignalHandler); err != nil {
		setupLog.Error(err, "unable to start manager")
		// TODO: return the error instead of exit
		os.Exit(1)
	}

	return nil
}
