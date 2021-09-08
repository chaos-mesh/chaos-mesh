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

package store

import (
	"context"

	"go.uber.org/fx"

	"github.com/jinzhu/gorm"
	ctrl "sigs.k8s.io/controller-runtime"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
)

var (
	// Module includes the providers provided by store.
	Module = fx.Options(
		fx.Provide(
			experiment.NewStore,
			event.NewStore,
			schedule.NewStore,
			workflow.NewStore,
		),
		fx.Invoke(experiment.DeleteIncompleteExperiments),
		fx.Invoke(schedule.DeleteIncompleteSchedules),
	)
	sqliteDriver = "sqlite3"
	log          = ctrl.Log.WithName("store").WithName("dbstore")
)

// NewDBStore returns a new gorm.DB
func NewDBStore(lc fx.Lifecycle, conf *config.ChaosDashboardConfig) (*gorm.DB, error) {
	ds := conf.Database.Datasource

	// fix error `database is locked`, refer to https://github.com/mattn/go-sqlite3/blob/master/README.md#faq
	if conf.Database.Driver == sqliteDriver {
		ds += "?cache=shared"
	}

	gormDB, err := gorm.Open(conf.Database.Driver, ds)
	if err != nil {
		log.Error(err, "Failed to open DB: ", "driver => ", conf.Database.Driver, " datasource => ", conf.Database.Datasource)

		return nil, err
	}

	// fix error `database is locked`, refer to https://github.com/mattn/go-sqlite3/blob/master/README.md#faq
	if conf.Database.Driver == sqliteDriver {
		gormDB.DB().SetMaxOpenConns(1)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return gormDB.Close()
		},
	})

	return gormDB, nil
}
