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
	"encoding/json"
	"time"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// ExperimentStore defines operations for working with archive experiments
type ExperimentStore interface {
	// List returns an archive experiment list from the datastore.
	List(ctx context.Context, kind, namespace, name string) ([]*ArchiveExperiment, error)

	// ListMeta returns an archive experiment metadata list from the datastore.
	ListMeta(ctx context.Context, kind, namespace, name string) ([]*ArchiveExperimentMeta, error)

	// Find returns an archive experiment by ID.
	Find(context.Context, int64) (*ArchiveExperiment, error)

	// Delete deletes the experiment from the datastore.
	Delete(context.Context, *ArchiveExperiment) error

	// DetailList returns a list of archive experiments from the datastore.
	DetailList(ctx context.Context, kind, namespace, name, uid string) ([]*ArchiveExperiment, error)

	// DeleteByFinishTime deletes experiments whose time difference is greater than the given time from FinishTime.
	DeleteByFinishTime(context.Context, time.Duration) error

	// Archive archives experiments whose "archived" field is false,
	Archive(ctx context.Context, namespace, name string) error

	// Set sets the experiment.
	Set(context.Context, *ArchiveExperiment) error

	// FindByUID returns an experiment record by the UID of the experiment.
	FindByUID(ctx context.Context, UID string) (*ArchiveExperiment, error)

	// FindMetaByUID returns an archive experiment by UID.
	FindMetaByUID(context.Context, string) (*ArchiveExperimentMeta, error)
}

// ArchiveExperiment represents an experiment instance.
type ArchiveExperiment struct {
	ArchiveExperimentMeta
	Experiment string `gorm:"size:2048"`
}

// ArchiveExperimentMeta defines the meta data for ArchiveExperiment.
type ArchiveExperimentMeta struct {
	ID         uint       `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `sql:"index" json:"deleted_at"`
	Name       string     `json:"name"`
	Namespace  string     `json:"namespace"`
	Kind       string     `json:"kind"`
	Action     string     `json:"action"`
	UID        string     `gorm:"index:uid" json:"uid"`
	StartTime  time.Time  `json:"start_time"`
	FinishTime time.Time  `json:"finish_time"`
	Archived   bool       `json:"archived"`
}

// ExperimentInfo defines a form data of Experiment from API.
type ExperimentInfo struct {
	Name        string            `json:"name" binding:"required,NameValid"`
	Namespace   string            `json:"namespace" binding:"required,NameValid"`
	Labels      map[string]string `json:"labels" binding:"MapSelectorsValid"`
	Annotations map[string]string `json:"annotations" binding:"MapSelectorsValid"`
	Scope       ScopeInfo         `json:"scope"`
	Target      TargetInfo        `json:"target"`
	Scheduler   SchedulerInfo     `json:"scheduler"`
}

// ScopeInfo defines the scope of the Experiment.
type ScopeInfo struct {
	SelectorInfo
	Mode  string `json:"mode" binding:"oneof='' 'one' 'all' 'fixed' 'fixed-percent' 'random-max-percent'"`
	Value string `json:"value" binding:"ValueValid"`
}

// TODO: consider moving this to a common package
// SelectorInfo defines the selector options of the Experiment.
type SelectorInfo struct {
	NamespaceSelectors  []string          `json:"namespace_selectors" binding:"NamespaceSelectorsValid"`
	LabelSelectors      map[string]string `json:"label_selectors" binding:"MapSelectorsValid"`
	AnnotationSelectors map[string]string `json:"annotation_selectors" binding:"MapSelectorsValid"`
	FieldSelectors      map[string]string `json:"field_selectors" binding:"MapSelectorsValid"`
	PhaseSelector       []string          `json:"phase_selectors" binding:"PhaseSelectorsValid"`

	// Pods is a map of string keys and a set values that used to select pods.
	// The key defines the namespace which pods belong,
	// and the each values is a set of pod names.
	Pods map[string][]string `json:"pods" binding:"PodsValid"`
}

// ParseSelector parses SelectorInfo to v1alpha1.SelectorSpec
func (s *SelectorInfo) ParseSelector() v1alpha1.SelectorSpec {
	selector := v1alpha1.SelectorSpec{}

	for _, ns := range s.NamespaceSelectors {
		selector.Namespaces = append(selector.Namespaces, ns)
	}

	selector.LabelSelectors = make(map[string]string)
	for key, val := range s.LabelSelectors {
		selector.LabelSelectors[key] = val
	}

	selector.AnnotationSelectors = make(map[string]string)
	for key, val := range s.AnnotationSelectors {
		selector.AnnotationSelectors[key] = val
	}

	selector.FieldSelectors = make(map[string]string)
	for key, val := range s.FieldSelectors {
		selector.FieldSelectors[key] = val
	}

	for _, ph := range s.PhaseSelector {
		selector.PodPhaseSelectors = append(selector.PodPhaseSelectors, ph)
	}

	if s.Pods != nil {
		selector.Pods = s.Pods
	}

	return selector
}

// TargetInfo defines the information of target objects.
type TargetInfo struct {
	Kind         string            `json:"kind" binding:"required,oneof=PodChaos NetworkChaos IoChaos KernelChaos TimeChaos StressChaos"`
	PodChaos     *PodChaosInfo     `json:"pod_chaos,omitempty" binding:"RequiredFieldEqual=Kind:PodChaos"`
	NetworkChaos *NetworkChaosInfo `json:"network_chaos,omitempty" binding:"RequiredFieldEqual=Kind:NetworkChaos"`
	IOChaos      *IOChaosInfo      `json:"io_chaos,omitempty" binding:"RequiredFieldEqual=Kind:IoChaos"`
	KernelChaos  *KernelChaosInfo  `json:"kernel_chaos,omitempty" binding:"RequiredFieldEqual=Kind:KernelChaos"`
	TimeChaos    *TimeChaosInfo    `json:"time_chaos,omitempty" binding:"RequiredFieldEqual=Kind:TimeChaos"`
	StressChaos  *StressChaosInfo  `json:"stress_chaos,omitempty" binding:"RequiredFieldEqual=Kind:StressChaos"`
}

// SchedulerInfo defines the scheduler information.
type SchedulerInfo struct {
	Cron     string `json:"cron" binding:"CronValid"`
	Duration string `json:"duration" binding:"DurationValid"`
}

// PodChaosInfo defines the basic information of pod chaos for creating a new PodChaos.
type PodChaosInfo struct {
	Action        string `json:"action" binding:"oneof='' 'pod-kill' 'pod-failure' 'container-kill'"`
	ContainerName string `json:"container_name"`
}

// NetworkChaosInfo defines the basic information of network chaos for creating a new NetworkChaos.
type NetworkChaosInfo struct {
	Action      string                  `json:"action" binding:"oneof='' 'netem' 'delay' 'loss' 'duplicate' 'corrupt' 'partition' 'bandwidth'"`
	Delay       *v1alpha1.DelaySpec     `json:"delay" binding:"RequiredFieldEqual=Action:delay"`
	Loss        *v1alpha1.LossSpec      `json:"loss" binding:"RequiredFieldEqual=Action:loss"`
	Duplicate   *v1alpha1.DuplicateSpec `json:"duplicate" binding:"RequiredFieldEqual=Action:duplicate"`
	Corrupt     *v1alpha1.CorruptSpec   `json:"corrupt" binding:"RequiredFieldEqual=Action:corrupt"`
	Bandwidth   *v1alpha1.BandwidthSpec `json:"bandwidth" binding:"RequiredFieldEqual=Action:bandwidth"`
	Direction   string                  `json:"direction" binding:"oneof='' 'to' 'from' 'both'"`
	TargetScope *ScopeInfo              `json:"target_scope"`
}

// IOChaosInfo defines the basic information of io chaos for creating a new IOChaos.
type IOChaosInfo struct {
	Action     string                     `json:"action" binding:"oneof='' 'latency' 'fault' 'attrOverride'"`
	Delay      string                     `json:"delay"`
	Errno      uint32                     `json:"errno"`
	Attr       *v1alpha1.AttrOverrideSpec `json:"attr"`
	Path       string                     `json:"path"`
	Percent    int                        `json:"percent"`
	Methods    []v1alpha1.IoMethod        `json:"methods"`
	VolumePath string                     `json:"volume_path"`
}

// KernelChaosInfo defines the basic information of kernel chaos for creating a new KernelChaos.
type KernelChaosInfo struct {
	FailKernRequest v1alpha1.FailKernRequest `json:"fail_kern_request"`
}

// TimeChaosInfo defines the basic information of time chaos for creating a new TimeChaos.
type TimeChaosInfo struct {
	TimeOffset     string   `json:"time_offset"`
	ClockIDs       []string `json:"clock_ids"`
	ContainerNames []string `json:"container_names"`
}

// StressChaosInfo defines the basic information of stress chaos for creating a new StressChaos.
type StressChaosInfo struct {
	Stressors         *v1alpha1.Stressors `json:"stressors"`
	StressngStressors string              `json:"stressng_stressors,omitempty"`
	ContainerName     *string             `json:"container_name,omitempty"`
}

func (e *ArchiveExperiment) ParsePodChaos() (ExperimentInfo, error) {
	chaos := &v1alpha1.PodChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return ExperimentInfo{}, err
	}

	info := ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: ScopeInfo{
			SelectorInfo: SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: TargetInfo{
			Kind: v1alpha1.KindPodChaos,
			PodChaos: &PodChaosInfo{
				Action:        string(chaos.Spec.Action),
				ContainerName: chaos.Spec.ContainerName,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return info, nil
}

func (e *ArchiveExperiment) ParseNetworkChaos() (ExperimentInfo, error) {
	chaos := &v1alpha1.NetworkChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return ExperimentInfo{}, err
	}

	info := ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: ScopeInfo{
			SelectorInfo: SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: TargetInfo{
			Kind: v1alpha1.KindNetworkChaos,
			NetworkChaos: &NetworkChaosInfo{
				Action:    string(chaos.Spec.Action),
				Delay:     chaos.Spec.Delay,
				Loss:      chaos.Spec.Loss,
				Duplicate: chaos.Spec.Duplicate,
				Corrupt:   chaos.Spec.Corrupt,
				Bandwidth: chaos.Spec.Bandwidth,
				Direction: string(chaos.Spec.Direction),
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	if chaos.Spec.Target != nil {
		info.Target.NetworkChaos.TargetScope.SelectorInfo = SelectorInfo{
			NamespaceSelectors:  chaos.Spec.Target.TargetSelector.Namespaces,
			LabelSelectors:      chaos.Spec.Target.TargetSelector.LabelSelectors,
			AnnotationSelectors: chaos.Spec.Target.TargetSelector.AnnotationSelectors,
			FieldSelectors:      chaos.Spec.Target.TargetSelector.FieldSelectors,
			PhaseSelector:       chaos.Spec.Target.TargetSelector.PodPhaseSelectors,
		}
		info.Target.NetworkChaos.TargetScope.Mode = string(chaos.Spec.Target.TargetMode)
		info.Target.NetworkChaos.TargetScope.Value = chaos.Spec.Target.TargetValue
	}

	return info, nil
}
func (e *ArchiveExperiment) ParseIOChaos() (ExperimentInfo, error) {
	chaos := &v1alpha1.IoChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return ExperimentInfo{}, err
	}

	var methods []string
	for _, method := range chaos.Spec.Methods {
		methods = append(methods, string(method))
	}
	info := ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: ScopeInfo{
			SelectorInfo: SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: TargetInfo{
			Kind: v1alpha1.KindIOChaos,
			IOChaos: &IOChaosInfo{
				Action:     string(chaos.Spec.Action),
				Delay:      chaos.Spec.Delay,
				Errno:      chaos.Spec.Errno,
				Attr:       chaos.Spec.Attr,
				Path:       chaos.Spec.Path,
				Percent:    chaos.Spec.Percent,
				Methods:    chaos.Spec.Methods,
				VolumePath: chaos.Spec.VolumePath,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return info, nil
}
func (e *ArchiveExperiment) ParseTimeChaos() (ExperimentInfo, error) {
	chaos := &v1alpha1.TimeChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return ExperimentInfo{}, err
	}

	info := ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: ScopeInfo{
			SelectorInfo: SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: TargetInfo{
			Kind: v1alpha1.KindTimeChaos,
			TimeChaos: &TimeChaosInfo{
				TimeOffset:     chaos.Spec.TimeOffset,
				ClockIDs:       chaos.Spec.ClockIds,
				ContainerNames: chaos.Spec.ContainerNames,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return info, nil
}
func (e *ArchiveExperiment) ParseKernelChaos() (ExperimentInfo, error) {
	chaos := &v1alpha1.KernelChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return ExperimentInfo{}, err
	}

	info := ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: ScopeInfo{
			SelectorInfo: SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: TargetInfo{
			Kind: v1alpha1.KindKernelChaos,
			KernelChaos: &KernelChaosInfo{
				FailKernRequest: chaos.Spec.FailKernRequest,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	return info, nil
}
func (e *ArchiveExperiment) ParseStressChaos() (ExperimentInfo, error) {
	chaos := &v1alpha1.StressChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return ExperimentInfo{}, err
	}

	info := ExperimentInfo{
		Name:        chaos.Name,
		Namespace:   chaos.Namespace,
		Labels:      chaos.Labels,
		Annotations: chaos.Annotations,
		Scope: ScopeInfo{
			SelectorInfo: SelectorInfo{
				NamespaceSelectors:  chaos.Spec.Selector.Namespaces,
				LabelSelectors:      chaos.Spec.Selector.LabelSelectors,
				AnnotationSelectors: chaos.Spec.Selector.AnnotationSelectors,
				FieldSelectors:      chaos.Spec.Selector.FieldSelectors,
				PhaseSelector:       chaos.Spec.Selector.PodPhaseSelectors,
				Pods:                chaos.Spec.Selector.Pods,
			},
			Mode:  string(chaos.Spec.Mode),
			Value: chaos.Spec.Value,
		},
		Target: TargetInfo{
			Kind: v1alpha1.KindStressChaos,
			StressChaos: &StressChaosInfo{
				Stressors:         chaos.Spec.Stressors,
				StressngStressors: chaos.Spec.StressngStressors,
			},
		},
	}

	if chaos.Spec.Scheduler != nil {
		info.Scheduler.Cron = chaos.Spec.Scheduler.Cron
	}

	if chaos.Spec.Duration != nil {
		info.Scheduler.Duration = *chaos.Spec.Duration
	}

	if chaos.Spec.ContainerName != nil {
		info.Target.StressChaos.ContainerName = chaos.Spec.ContainerName
	}

	return info, nil
}
