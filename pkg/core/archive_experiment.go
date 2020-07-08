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

package core

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
)

// ExperimentStore defines operations for working with archive experiments
type ExperimentStore interface {
	// List returns an archive experiment list from the datastore.
	List(ctx context.Context, kind, namespace, name string) ([]*ArchiveExperiment, error)

	// ListMeta returns an archive experiment metadata list from the datastore.
	ListMeta(ctx context.Context, kind, namespace, name string) ([]*ArchiveExperimentMeta, error)

	// Find returns an archive experiment by ID.
	Find(context.Context, int64) (*ArchiveExperiment, error)

	// Create persists a new archive experiment to the datastore.
	Create(context.Context, *ArchiveExperiment) error

	// Delete deletes the experiment from the datastore.
	Delete(context.Context, *ArchiveExperiment) error

	// DetailList returns a list of archive experiments from the datastore.
	DetailList(ctx context.Context, kind, namespace, name, uid string) ([]*ArchiveExperiment, error)

	// FindByUID returns an archive experiment by UID.
	FindByUID(context.Context, string) (*ArchiveExperiment, error)

	// FindMetaByUID returns an archive experiment by UID.
	FindMetaByUID(context.Context, string) (*ArchiveExperimentMeta, error)

	// DeleteByFinishTime deletes experiments whose time difference is greater than the given time from FinishTime.
	DeleteByFinishTime(context.Context, time.Duration) error
}

// ArchiveExperiment represents an experiment instance.
type ArchiveExperiment struct {
	ArchiveExperimentMeta
	Experiment string `gorm:"size:2048"`
}

// ArchiveExperimentMeta defines the meta data for ArchiveExperiment.
type ArchiveExperimentMeta struct {
	gorm.Model
	Name       string
	Namespace  string
	Kind       string
	Action     string
	UID        string `gorm:"index:uid"`
	StartTime  time.Time
	FinishTime time.Time
}

// TODO: implement parse functions
func (e *ArchiveExperiment) ParsePodChaos() (*v1alpha1.PodChaos, error)       { return nil, nil }
func (e *ArchiveExperiment) ParseNetChaos() (*v1alpha1.NetworkChaos, error)   { return nil, nil }
func (e *ArchiveExperiment) ParseIOChaos() (*v1alpha1.IoChaos, error)         { return nil, nil }
func (e *ArchiveExperiment) ParseTimeChaos() (*v1alpha1.TimeChaos, error)     { return nil, nil }
func (e *ArchiveExperiment) ParseKernelChaos() (*v1alpha1.KernelChaos, error) { return nil, nil }
func (e *ArchiveExperiment) ParseStressChaos() (*v1alpha1.StressChaos, error) { return nil, nil }
