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

package ttlcontroller

import (
	"context"
	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/core"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	log "github.com/sirupsen/logrus"
	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// databaseTTLResyncPeriod defines the time interval to synchronize data in the database.
	databaseTTLResyncPeriod = 8 * time.Hour
	eventTTL = 7 * 24 * time.Hour
	archiveExperimentTTL = 14 * 24 * time.Hour
)

type Controller struct {
	archive      core.ExperimentStore
	event        core.EventStore
}

// NewController returns a new database ttl controller
func NewController(
	config *config.ChaosServerConfig,
	archive core.ExperimentStore,
	event core.EventStore,
) *Controller {
	controller := &Controller{
		archive:      archive,
		event:        event,
	}
	return controller
}

func Register (c *Controller, stopCh <-chan struct{}) error {
	defer runtimeutil.HandleCrash()
	log.Infof("starting database TTL controller")
	go wait.Until(c.runWorker, databaseTTLResyncPeriod, stopCh)
	log.Info("shutting database TTL controller")
	return nil
}

// runWorker is a long-running function that will continually call the
// function in order to delete the events and archives.
func (c *Controller) runWorker() {
	c.event.DeleteByFinishTime(context.Background(), eventTTL)
	c.archive.DeleteByFinishTime(context.Background(), archiveExperimentTTL)
}
