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

package dbstore

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.uber.org/fx"
	ctrl "sigs.k8s.io/controller-runtime"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config/dashboard"
)

var (
	sqliteDriver string = "sqlite3"
	log                 = ctrl.Log.WithName("store/dbstore")
)

// DB defines a db storage.
type DB struct {
	*gorm.DB
}

// NewDBStore returns a new DB
func NewDBStore(lc fx.Lifecycle, conf *config.ChaosDashboardConfig) (*DB, error) {
	dsn := conf.Database.Datasource

	// fix error `database is locked`, refer to https://github.com/mattn/go-sqlite3/blob/master/README.md#faq
	if conf.Database.Driver == sqliteDriver {
		dsn += "?cache=shared"
	}

	gormDB, err := gorm.Open(conf.Database.Driver, dsn)
	if err != nil {
		log.Error(err, "failed to open DB", "driver", conf.Database.Driver, "datasource", conf.Database.Datasource)
		return nil, err
	}

	// fix error `database is locked`, refer to https://github.com/mattn/go-sqlite3/blob/master/README.md#faq
	if conf.Database.Driver == sqliteDriver {
		gormDB.DB().SetMaxOpenConns(1)
	}

	db := &DB{
		gormDB,
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return db.Close()
		},
	})

	return db, nil
}
