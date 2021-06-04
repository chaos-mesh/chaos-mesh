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

package ttlcontroller

import (
	"context"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/core"

	runtimeutil "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	log = ctrl.Log.WithName("database ttl controller")
)

// Controller defines the database ttl controller
type Controller struct {
	experiment core.ExperimentStore
	event      core.EventStore
	ttlconfig  *TTLconfig
}

// TTLconfig defines the ttl
type TTLconfig struct {
	// databaseTTLResyncPeriod defines the time interval to cleanup data in the database.
	DatabaseTTLResyncPeriod time.Duration
	// EventTTL defines the ttl of events
	EventTTL time.Duration
	// ArchiveExperimentTTL defines the ttl of archive experiments
	ArchiveExperimentTTL time.Duration
}

// NewController returns a new database ttl controller
func NewController(
	experiment core.ExperimentStore,
	event core.EventStore,
	ttlc *TTLconfig,
) *Controller {
	return &Controller{
		experiment: experiment,
		event:      event,
		ttlconfig:  ttlc,
	}
}

// Register periodically calls function runWorker to delete the data.
func Register(c *Controller, controllerRuntimeStopCh <-chan struct{}) {
	defer runtimeutil.HandleCrash()
	log.Info("starting database TTL controller")
	go wait.Until(c.runWorker, c.ttlconfig.DatabaseTTLResyncPeriod, controllerRuntimeStopCh)
}

// runWorker is a long-running function that will call the
// function in order to delete the events and archives.
func (c *Controller) runWorker() {
	log.Info("deleting expired data from the database")
	c.event.DeleteByCreateTime(context.Background(), c.ttlconfig.EventTTL)
	c.experiment.DeleteByFinishTime(context.Background(), c.ttlconfig.ArchiveExperimentTTL)
}
