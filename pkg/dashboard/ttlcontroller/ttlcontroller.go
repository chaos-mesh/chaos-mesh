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

	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type Controller struct {
	logger     logr.Logger
	event      core.EventStore
	experiment core.ExperimentStore
	schedule   core.ScheduleStore
	workflow   core.WorkflowStore
	ttlconfig  *TTLConfig
}

// TTLConfig defines all the TTL-related configurations.
type TTLConfig struct {
	// ResyncPeriod defines the period of cleaning data.
	ResyncPeriod time.Duration

	// TTL of events.
	EventTTL time.Duration
	// TTL of experiments.
	ExperimentTTL time.Duration
	// TTL of schedules.
	ScheduleTTL time.Duration
	// TTL of workflows.
	WorkflowTTL time.Duration
}

// TTLConfigWithStringTime defines all the TTL-related configurations with string type time.
type TTLConfigWithStringTime struct {
	ResyncPeriod string `envconfig:"CLEAN_SYNC_PERIOD" default:"12h"`

	EventTTL      string `envconfig:"TTL_EVENT"         default:"168h"` // one week
	ExperimentTTL string `envconfig:"TTL_EXPERIMENT"    default:"336h"` // two weeks
	ScheduleTTL   string `envconfig:"TTL_EXPERIMENT"    default:"336h"`
	WorkflowTTL   string `envconfig:"TTL_EXPERIMENT"    default:"336h"`
}

func (config *TTLConfigWithStringTime) parse() (*TTLConfig, error) {
	syncPeriod, err := time.ParseDuration(config.ResyncPeriod)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration sync period")
	}

	event, err := time.ParseDuration(config.EventTTL)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for event")
	}

	experiment, err := time.ParseDuration(config.ExperimentTTL)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for experiment")
	}

	schedule, err := time.ParseDuration(config.ScheduleTTL)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for schedule")
	}

	workflow, err := time.ParseDuration(config.WorkflowTTL)
	if err != nil {
		return nil, errors.Wrap(err, "parse configuration TTL for workflow")
	}

	return &TTLConfig{
		ResyncPeriod:  syncPeriod,
		EventTTL:      event,
		ExperimentTTL: experiment,
		ScheduleTTL:   schedule,
		WorkflowTTL:   workflow,
	}, nil
}

// NewController returns a new database ttl controller
func NewController(
	event core.EventStore,
	experiment core.ExperimentStore,
	schedule core.ScheduleStore,
	workflow core.WorkflowStore,
	ttlconfig *TTLConfig,
	logger logr.Logger,
) *Controller {
	return &Controller{
		experiment: experiment,
		event:      event,
		schedule:   schedule,
		workflow:   workflow,
		ttlconfig:  ttlconfig,
		logger:     logger,
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

	_ = c.event.DeleteByDuration(ctx, c.ttlconfig.EventTTL)
	c.experiment.DeleteByFinishTime(ctx, c.ttlconfig.ExperimentTTL)
	c.schedule.DeleteByFinishTime(ctx, c.ttlconfig.ScheduleTTL)
	c.workflow.DeleteByFinishTime(ctx, c.ttlconfig.WorkflowTTL)
}
