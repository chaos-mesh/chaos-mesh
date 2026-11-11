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

package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/metrics"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
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
		fx.Supply(controllermetrics.Registry),
		fx.Invoke(metrics.Register),
		fx.Invoke(experiment.DeleteIncompleteExperiments),
		fx.Invoke(schedule.DeleteIncompleteSchedules),
	)
	sqliteDriver = "sqlite3"
)

// NewDBStore returns a new gorm.DB
func NewDBStore(lc fx.Lifecycle, conf *config.ChaosDashboardConfig, logger logr.Logger) (*gorm.DB, error) {
	var dialector gorm.Dialector
	ds := conf.Database.Datasource

	switch conf.Database.Driver {
	case "mysql":
		dialector = mysql.Open(ds)
	case "postgres":
		dialector = postgres.Open(ds)
	case sqliteDriver:
		if !strings.Contains(ds, "?") {
			ds += "?cache=shared"
		} else if !strings.Contains(ds, "cache=") {
			ds += "&cache=shared"
		}
		if !strings.HasPrefix(ds, "file:") {
			ds = "file:" + ds
		}
		dialector = sqlite.Open(ds)
	case "mssql":
		dialector = sqlserver.Open(ds)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", conf.Database.Driver)
	}

	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		logger.Error(err, "Failed to open DB: ", "driver => ", conf.Database.Driver)
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	if conf.Database.Driver == sqliteDriver {
		sqlDB.SetMaxOpenConns(1)
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return sqlDB.Close()
		},
	})

	return gormDB, nil
}
