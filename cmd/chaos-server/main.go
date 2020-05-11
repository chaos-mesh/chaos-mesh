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

package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/pingcap/chaos-mesh/pkg/apiserver"
	"github.com/pingcap/chaos-mesh/pkg/collector"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/store"
	"github.com/pingcap/chaos-mesh/pkg/store/dbstore"
	"github.com/pingcap/chaos-mesh/pkg/version"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	log = ctrl.Log.WithName("setup")
)

var (
	printVersion bool
)

func main() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.Parse()

	conf, err := config.EnvironChaosServer()
	if err != nil {
		log.Error(err, "main: invalid configuration")
		os.Exit(1)
	}

	version.PrintVersionInfo("Chaos Server")
	if printVersion {
		os.Exit(0)
	}

	ctrl.SetLogger(zap.Logger(true))

	stopCh := ctrl.SetupSignalHandler()

	app := fx.New(
		fx.Provide(
			func() (<-chan struct{}, *config.ChaosServerConfig) {
				return stopCh, &conf
			},
			dbstore.NewDBStore,
			collector.NewServer,
		),
		store.Module,
		apiserver.Module,
		fx.Invoke(collector.Register))

	startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		log.Error(err, "failed to start app")
		os.Exit(1)
	}

	<-stopCh
	stopCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Error(err, "failed to stop app")
		os.Exit(1)
	}
}
