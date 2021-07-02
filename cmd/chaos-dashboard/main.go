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

package main

import (
	"flag"
	"os"

	"go.uber.org/fx"

	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/apiserver"
	"github.com/chaos-mesh/chaos-mesh/pkg/collector"
	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/store"
	"github.com/chaos-mesh/chaos-mesh/pkg/store/dbstore"
	"github.com/chaos-mesh/chaos-mesh/pkg/ttlcontroller"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
)

var (
	log = ctrl.Log.WithName("dashboard")
)

var (
	printVersion bool
)

// @title Chaos Mesh Dashboard API
// @version 2.0
// @description Swagger docs for Chaos Mesh Dashboard. If you encounter any problems with API, please click on the issues link below to report bugs or questions.

// @contact.name Issues
// @contact.url https://github.com/chaos-mesh/chaos-mesh/issues

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /api
func main() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.Parse()

	version.PrintVersionInfo("Chaos Dashboard")
	if printVersion {
		os.Exit(0)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mainLog := log.WithName("main")

	dashboardConfig, err := config.GetChaosDashboardEnv()
	if err != nil {
		mainLog.Error(err, "invalid ChaosDashboardConfig")
		os.Exit(1)
	}
	dashboardConfig.Version = version.Get().GitVersion

	persistTTLConfigParsed, err := config.ParsePersistTTLConfig(dashboardConfig.PersistTTL)
	if err != nil {
		mainLog.Error(err, "invalid PersistTTLConfig")
		os.Exit(1)
	}

	ctrlRuntimeStopCh := ctrl.SetupSignalHandler()
	app := fx.New(
		fx.Provide(
			func() (<-chan struct{}, *config.ChaosDashboardConfig, *ttlcontroller.TTLconfig) {
				return ctrlRuntimeStopCh, dashboardConfig, persistTTLConfigParsed
			},
			dbstore.NewDBStore,
			collector.NewServer,
			ttlcontroller.NewController,
		),
		store.Module,
		apiserver.Module,
		fx.Invoke(collector.Register),
		fx.Invoke(ttlcontroller.Register),
	)

	app.Run()
}
