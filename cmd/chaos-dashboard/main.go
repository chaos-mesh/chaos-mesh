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
	"context"
	"flag"
	stdlog "log"
	"os"

	fxlogr "github.com/chaos-mesh/fx-logr"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"go.uber.org/fx"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/collector"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/ttlcontroller"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
)

// @title Chaos Mesh Dashboard API
// @version 2.5
// @description Swagger for Chaos Mesh Dashboard. If you encounter any problems with API, please click on the issues link below to report.

// @contact.name GitHub Issues
// @contact.url https://github.com/chaos-mesh/chaos-mesh/issues

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api
func main() {
	var printVersion bool

	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.Parse()

	version.PrintVersionInfo("Chaos Dashboard")
	if printVersion {
		os.Exit(0)
	}

	rootLogger, err := log.NewDefaultZapLogger()
	if err != nil {
		stdlog.Fatal("failed to create root logger", err)
	}

	ctrl.SetLogger(rootLogger)
	mainLog := rootLogger.WithName("main")
	fxLogger := rootLogger.WithName("fx")

	dashboardConfig, err := config.GetChaosDashboardEnv()
	if err != nil {
		mainLog.Error(err, "invalid ChaosDashboardConfig")
		os.Exit(1)
	}
	dashboardConfig.Version = version.Get().GitVersion

	persistTTLConfigParsed, err := dashboardConfig.PersistTTL.Parse()
	if err != nil {
		mainLog.Error(err, "invalid PersistTTLConfig")
		os.Exit(1)
	}

	controllerRuntimeSignalHandlerContext := ctrl.SetupSignalHandler()
	app := fx.New(
		fx.WithLogger(fxlogr.WithLogr(&fxLogger)),
		fx.Supply(rootLogger),
		fx.Provide(
			func() (context.Context, *config.ChaosDashboardConfig, *config.TTLConfig) {
				return controllerRuntimeSignalHandlerContext, dashboardConfig, persistTTLConfigParsed
			},
			store.Bootstrap,
			collector.Bootstrap,
			ttlcontroller.Bootstrap,
		),
		store.Module,
		apiserver.Module,
		fx.Invoke(collector.Register),
		fx.Invoke(ttlcontroller.Register),
	)

	app.Run()
}
