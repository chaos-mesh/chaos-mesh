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

// TargetInfo defines the information of target objects.
type TargetInfo struct {
	Kind         string            `json:"kind" binding:"required,oneof=PodChaos NetworkChaos IOChaos KernelChaos TimeChaos StressChaos DNSChaos AwsChaos GcpChaos"`
	PodChaos     *PodChaosInfo     `json:"pod_chaos,omitempty" binding:"RequiredFieldEqual=Kind:PodChaos"`
	NetworkChaos *NetworkChaosInfo `json:"network_chaos,omitempty" binding:"RequiredFieldEqual=Kind:NetworkChaos"`
	IOChaos      *IOChaosInfo      `json:"io_chaos,omitempty" binding:"RequiredFieldEqual=Kind:IOChaos"`
	KernelChaos  *KernelChaosInfo  `json:"kernel_chaos,omitempty" binding:"RequiredFieldEqual=Kind:KernelChaos"`
	TimeChaos    *TimeChaosInfo    `json:"time_chaos,omitempty" binding:"RequiredFieldEqual=Kind:TimeChaos"`
	StressChaos  *StressChaosInfo  `json:"stress_chaos,omitempty" binding:"RequiredFieldEqual=Kind:StressChaos"`
	DNSChaos     *DNSChaosInfo     `json:"dns_chaos,omitempty" binding:"RequiredFieldEqual=Kind:DNSChaos"`
	AwsChaos     *AwsChaosInfo     `json:"aws_chaos,omitempty" binding:"RequiredFieldEqual=Kind:AwsChaos"`
	GcpChaos     *GcpChaosInfo     `json:"gcp_chaos,omitempty" binding:"RequiredFieldEqual=Kind:GcpChaos"`
}

// SchedulerInfo defines the scheduler information.
type SchedulerInfo struct {
	Duration string `json:"duration" binding:"DurationValid"`
}

// PodChaosInfo defines the basic information of pod chaos for creating a new PodChaos.
type PodChaosInfo struct {
	Action         string   `json:"action" binding:"oneof='' 'pod-kill' 'pod-failure' 'container-kill'"`
	ContainerNames []string `json:"container_names,omitempty"`
	GracePeriod    int64    `json:"grace_period"`
}

// NetworkChaosInfo defines the basic information of network chaos for creating a new NetworkChaos.
type NetworkChaosInfo struct {
	Action          string                  `json:"action" binding:"oneof='' 'netem' 'delay' 'loss' 'duplicate' 'corrupt' 'partition' 'bandwidth'"`
	Delay           *v1alpha1.DelaySpec     `json:"delay" binding:"RequiredFieldEqual=Action:delay"`
	Loss            *v1alpha1.LossSpec      `json:"loss" binding:"RequiredFieldEqual=Action:loss"`
	Duplicate       *v1alpha1.DuplicateSpec `json:"duplicate" binding:"RequiredFieldEqual=Action:duplicate"`
	Corrupt         *v1alpha1.CorruptSpec   `json:"corrupt" binding:"RequiredFieldEqual=Action:corrupt"`
	Bandwidth       *v1alpha1.BandwidthSpec `json:"bandwidth" binding:"RequiredFieldEqual=Action:bandwidth"`
	Direction       string                  `json:"direction" binding:"oneof='' 'to' 'from' 'both'"`
	TargetScope     *ScopeInfo              `json:"target_scope"`
	ExternalTargets []string                `json:"external_targets"`
}

// IOChaosInfo defines the basic information of io chaos for creating a new IOChaos.
type IOChaosInfo struct {
	Action        string                     `json:"action" binding:"oneof='' 'latency' 'fault' 'attrOverride'"`
	Delay         string                     `json:"delay"`
	Errno         uint32                     `json:"errno"`
	Attr          *v1alpha1.AttrOverrideSpec `json:"attr"`
	Mistake       *v1alpha1.MistakeSpec      `json:"mistake"`
	Path          string                     `json:"path"`
	Percent       int                        `json:"percent"`
	Methods       []v1alpha1.IoMethod        `json:"methods"`
	VolumePath    string                     `json:"volume_path"`
	ContainerName string                     `json:"container_name"`
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

// DNSChaosInfo defines the basic information of dns chaos for creating a new DNSChaos.
type DNSChaosInfo struct {
	Action             string   `json:"action" binding:"oneof='error' 'random'"`
	DomainNamePatterns []string `json:"patterns"`
	ContainerNames     []string `json:"container_names,omitempty"`
}

// AwsChaosInfo defines the basic information of aws chaos for creating a new AwsChaos.
type AwsChaosInfo struct {
	Action      string  `json:"action" binding:"oneof='ec2-stop' 'ec2-restart' 'detach-volume'"`
	SecretName  *string `json:"secretName,omitempty"`
	AwsRegion   string  `json:"awsRegion"`
	Ec2Instance string  `json:"ec2Instance"`
	EbsVolume   *string `json:"volumeID,omitempty"`
	DeviceName  *string `json:"deviceName,omitempty"`
}

// GcpChaosInfo defines the basic information of aws chaos for creating a new AwsChaos.
type GcpChaosInfo struct {
	Action      string    `json:"action" binding:"oneof='node-stop' 'node-reset' 'disk-loss'"`
	SecretName  *string   `json:"secretName,omitempty"`
	Project     string    `json:"project"`
	Zone        string    `json:"zone"`
	Instance    string    `json:"instance"`
	DeviceNames *[]string `json:"deviceNames,omitempty"`
}

// ParsePodChaos Parse PodChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParsePodChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.PodChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseNetworkChaos Parse NetworkChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseNetworkChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.NetworkChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseIOChaos Parse IOChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseIOChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.IOChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseTimeChaos Parse TimeChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseTimeChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.TimeChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseKernelChaos Parse KernelChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseKernelChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.KernelChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseStressChaos Parse StressChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseStressChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.StressChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseDNSChaos Parse DNSChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseDNSChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.DNSChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseAwsChaos Parse AwsChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseAwsChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.AwsChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}

// ParseGcpChaos Parse GcpChaos JSON string into KubeObjectDesc.
func (e *Experiment) ParseGcpChaos() (KubeObjectDesc, error) {
	chaos := &v1alpha1.GcpChaos{}
	if err := json.Unmarshal([]byte(e.Experiment), &chaos); err != nil {
		return KubeObjectDesc{}, err
	}

	return KubeObjectDesc{
		TypeMeta: metav1.TypeMeta{
			APIVersion: chaos.APIVersion,
			Kind:       chaos.Kind,
		},
		Meta: KubeObjectMeta{
			Name:        chaos.Name,
			Namespace:   chaos.Namespace,
			Labels:      chaos.Labels,
			Annotations: chaos.Annotations,
		},
		Spec: chaos.Spec,
	}, nil
}
