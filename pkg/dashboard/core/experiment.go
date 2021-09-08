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

package core

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/jinzhu/gorm"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// ExperimentStore defines operations for working with experiments.
type ExperimentStore interface {
	// ListMeta returns experiment metadata list from the datastore.
	ListMeta(ctx context.Context, kind, namespace, name string, archived bool) ([]*ExperimentMeta, error)

	// FindByUID returns an experiment by UID.
	FindByUID(ctx context.Context, UID string) (*Experiment, error)

	// FindMetaByUID returns an experiment metadata by UID.
	FindMetaByUID(context.Context, string) (*ExperimentMeta, error)

	// Set saves the experiment to datastore.
	Set(context.Context, *Experiment) error

	// Archive archives experiments which "archived" field is false.
	Archive(ctx context.Context, namespace, name string) error

	// Delete deletes the archive from the datastore.
	Delete(context.Context, *Experiment) error

	// DeleteByFinishTime deletes archives which time difference is greater than the given time from FinishTime.
	DeleteByFinishTime(context.Context, time.Duration) error

	// DeleteByUIDs deletes archives by the uid list.
	DeleteByUIDs(context.Context, []string) error

	// DeleteIncompleteExperiments deletes all incomplete experiments.
	// If the chaos-dashboard was restarted and the experiment is completed during the restart,
	// which means the experiment would never save the finish_time.
	// DeleteIncompleteExperiments can be used to delete all incomplete experiments to avoid this case.
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
	UID        string    `gorm:"index:uid" json:"uid"`
	Kind       string    `json:"kind"`
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Action     string    `json:"action"`
	StartTime  time.Time `json:"start_time"`
	FinishTime time.Time `json:"finish_time"`
	Archived   bool      `json:"archived"`
}

// SelectorInfo defines the selector options of the Experiment.
type SelectorInfo struct {
	Namespaces          []string                          `json:"namespaces,omitempty" binding:"NamespaceSelectorsValid"`
	Nodes               []string                          `json:"nodes,omitempty"`
	NodeSelectors       map[string]string                 `json:"node_selectors,omitempty"`
	FieldSelectors      map[string]string                 `json:"field_selectors,omitempty" binding:"MapSelectorsValid"`
	LabelSelectors      map[string]string                 `json:"label_selectors,omitempty" binding:"MapSelectorsValid"`
	ExpressionSelectors []metav1.LabelSelectorRequirement `json:"expression_selectors,omitempty" binding:"RequirementSelectorsValid"`
	AnnotationSelectors map[string]string                 `json:"annotation_selectors,omitempty" binding:"MapSelectorsValid"`
	PodPhaseSelectors   []string                          `json:"phase_selectors,omitempty" binding:"PhaseSelectorsValid"`

	// Pods is a map of string keys and a set values that used to select pods.
	// The key defines the namespace which pods belong,
	// and the each values is a set of pod names.
	Pods map[string][]string `json:"pods,omitempty" binding:"PodsValid"`
}

// ParseSelector parses SelectorInfo to v1alpha1.SelectorSpec
func (s *SelectorInfo) ParseSelector() v1alpha1.PodSelectorSpec {
	selector := v1alpha1.PodSelectorSpec{}
	selector.Namespaces = append(selector.Namespaces, s.Namespaces...)

	selector.LabelSelectors = make(map[string]string)
	for key, val := range s.LabelSelectors {
		selector.LabelSelectors[key] = val
	}

	selector.ExpressionSelectors = append(selector.ExpressionSelectors, s.ExpressionSelectors...)

	selector.AnnotationSelectors = make(map[string]string)
	for key, val := range s.AnnotationSelectors {
		selector.AnnotationSelectors[key] = val
	}

	selector.FieldSelectors = make(map[string]string)
	for key, val := range s.FieldSelectors {
		selector.FieldSelectors[key] = val
	}

	selector.PodPhaseSelectors = append(selector.PodPhaseSelectors, s.PodPhaseSelectors...)

	if s.Pods != nil {
		selector.Pods = s.Pods
	}

	return selector
}
