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

// Package ttlcontroller provides a TTL (time to live) mechanism to clear old objects
// in the database.
package ttlcontroller

import (
	"context"

	"github.com/go-logr/logr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type Controller struct {
	logger          logr.Logger
	eventStore      core.EventStore
	experimentStore core.ExperimentStore
	scheduleStore   core.ScheduleStore
	workflowStore   core.WorkflowStore
	ttlconfig       *config.TTLConfig
}

// NewController returns a new database ttl controller
func NewController(
	eventStore core.EventStore,
	experimentStore core.ExperimentStore,
	scheduleStore core.ScheduleStore,
	workflowStore core.WorkflowStore,
	ttlconfig *config.TTLConfig,
	logger logr.Logger,
) *Controller {
	return &Controller{
		experimentStore: experimentStore,
		eventStore:      eventStore,
		scheduleStore:   scheduleStore,
		workflowStore:   workflowStore,
		ttlconfig:       ttlconfig,
		logger:          logger,
	}
}

// Register periodically calls function runWorker to delete the data.
func Register(ctx context.Context, c *Controller) {
	defer utilruntime.HandleCrash()

	c.logger.Info("Starting database TTL controller")

	go wait.Until(c.runWorker, c.ttlconfig.ResyncPeriod, ctx.Done())
}

// runWorker is a long-running function that will be called in order to delete the events, archives, schedule, and workflow.
func (c *Controller) runWorker() {
	c.logger.Info("Deleting expired data from the database")

	ctx := context.Background()

	_ = c.eventStore.DeleteByDuration(ctx, c.ttlconfig.EventTTL)
	c.experimentStore.DeleteByFinishTime(ctx, c.ttlconfig.ExperimentTTL)
	c.scheduleStore.DeleteByFinishTime(ctx, c.ttlconfig.ScheduleTTL)
	c.workflowStore.DeleteByFinishTime(ctx, c.ttlconfig.WorkflowTTL)
}
