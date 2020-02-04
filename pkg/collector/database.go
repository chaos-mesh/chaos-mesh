// Copyright 2019 PingCAP, Inc.
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

package collector

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	databaseLog = ctrl.Log.WithName("database")
)

// DatabaseClient is a client for querying mysql
type DatabaseClient struct {
	db *sqlx.DB
}

type Event struct {
	Name              string          `json:"name"`
	Namespace         string          `json:"namespace"`
	Type              string          `json:"type"`
	AffectedNamespace map[string]bool `json:"affected_namespace"`
	StartTime         *time.Time      `json:"start_time"`
	EndTime           *time.Time      `json:"end_time"`
}

func NewDatabaseClient(dataSource string) (*DatabaseClient, error) {
	databaseLog.Info("connecting to database", "dataSource", dataSource)
	db, err := sqlx.Open("mysql", dataSource)
	if err != nil {
		return nil, err
	}
	databaseLog.Info("database connected")

	return &DatabaseClient{
		db,
	}, nil
}

func (client *DatabaseClient) WriteAffectedNamespace(e Event, id int64, tx *sql.Tx) error {
	for namespace := range e.AffectedNamespace {
		_, err := tx.Exec("INSERT INTO chaos_operator.affected_namespaces (event_id, namespace) VALUES (?, ?)", id, namespace)
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				databaseLog.Error(rb, "rollback error")
			}
			return err
		}
	}

	return nil
}

func (client *DatabaseClient) WriteEvent(e Event) error {
	tx, err := client.db.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec("INSERT INTO chaos_operator.events (name, namespace, type, start_time, end_time) VALUES (?, ?, ?, ?, NULL)", e.Name, e.Namespace, e.Type, e.StartTime) // Weired bug! I have to specify the database name
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			databaseLog.Error(rb, "rollback error")
		}
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			databaseLog.Error(rb, "rollback error")
		}
		return err
	}
	err = client.WriteAffectedNamespace(e, id, tx)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (client *DatabaseClient) UpdateEvent(e Event) error {
	tx, err := client.db.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec("UPDATE chaos_operator.events SET end_time=? WHERE name=? AND namespace=? AND isNULL(end_time) order by start_time desc limit 1;", e.EndTime, e.Name, e.Namespace)
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			databaseLog.Error(rb, "rollback error")
		}
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			databaseLog.Error(rb, "rollback error")
		}
		return err
	}

	if rows == 0 {
		result, err := tx.Exec("INSERT INTO chaos_operator.events (name, namespace, type, start_time, end_time) VALUES (?, ?, ?, ?, ?)", e.Name, e.Namespace, e.Type, e.StartTime, e.EndTime)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			if rb := tx.Rollback(); rb != nil {
				databaseLog.Error(rb, "rollback error")
			}
			return err
		}
		err = client.WriteAffectedNamespace(e, id, tx)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
