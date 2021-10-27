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

package core

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"
)

// ExperimentStore defines operations for working with experiments.
type ExperimentStore interface {
	// ListByFilter(context.Context, Filter, bool) ([]*ExperimentMeta, error)

	// ListMeta returns an experiment metadata list from the datastore.
	ListMeta(ctx context.Context, namespace, name, kind string, archived bool) ([]*ExperimentMeta, error)

	// FindByUID returns an experiment by UID.
	FindByUID(ctx context.Context, UID string) (*Experiment, error)

	// Create persists a new experiment to the datastore.
	Create(context.Context, *Experiment) error

	// Archive archives experiments which archived field is false.
	Archive(ctx context.Context, namespace, name string) error

	// Delete deletes the archive from the datastore.
	Delete(context.Context, *Experiment) error

	// DeleteByUIDs deletes archives by the UID list.
	DeleteByUIDs(context.Context, []string) error

	// DeleteByDuration delete experiments that exceed duration.
	DeleteByDuration(context.Context, time.Duration) error

	// DeleteIncompleteExperiments deletes all incomplete experiments.
	// If the chaos-dashboard was restarted and the experiment is completed during the restart,
	// which means the experiment would never update the delete_at field.
	// DeleteIncompleteExperiments can be used to delete all incomplete experiments to avoid this unexpected situation.
	DeleteIncompleteExperiments(context.Context) error
}

// Experiment represents an experiment instance. Use in db.
type Experiment struct {
	ExperimentMeta
	Experiment string `gorm:"size:2048"` // JSON string
}

// ExperimentMeta defines the metadata of an experiment. Use in db.
type ExperimentMeta struct {
	gorm.Model
	UID       string `gorm:"index:uid" json:"uid"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Action    string `json:"action"`
	Archived  bool   `json:"archived"`
}
