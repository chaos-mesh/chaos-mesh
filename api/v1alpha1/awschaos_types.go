package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +chaos-mesh:base

// AWSChaos is the Schema for the AWSChaos API
type AWSChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AWSChaosSpec   `json:"spec,omitempty"`
	Status AWSChaosStatus `json:"status,omitempty"`
}

type AWSChaosSpec struct {
	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`
	// Scheduler defines some schedule rules to
	// control the running time of the chaos experiment about pods.
	Scheduler *SchedulerSpec `json:"scheduler,omitempty"`

	// Action defines the specific chaos action.
	// Supported action:
	// +kubebuilder:validation:Enum=stop
	Action AWSChaosAction `json:"action"`

	Service AWSService `json:"service"`

	Resource string `json:"resource"`

	Selector AWSSelector `json:"selector"`

	Config AWSConfig `json:"config"`
}

type AWSConfig struct {
	Region string `json:"region"`

	Credential *corev1.LocalObjectReference `json:"credential"`
}

const (
	AWSAccessKeyID     = "accessKeyID"
	AWSSecretAccessKey = "secretAccessKey"
)

type AWSSelector struct {
	IDs []string `json:"ids"`

	Filters []AWSFilter `json:"filters"`
}

type AWSFilter struct {
	Name   *string  `json:"name"`
	Values []string `json:"values"`
}

type AWSService string

const (
	AWSServiceEC2     AWSService = "EC2"
	AWSServiceS3      AWSService = "S3"
	AWSServiceRoute53 AWSService = "Route53"
)

type AWSChaosAction string

const (
	// AWSActionBlock will block service of aws
	AWSActionBlock AWSChaosAction = "block"
	// AWSActionStop will stop specified aws resource
	AWSActionStop AWSChaosAction = "stop"
)

type AWSChaosStatus struct {
	ChaosStatus `json:",inline"`
	Snapshot    *AWSStatusSnapshot `json:"snapshot,omitempty"`
}

type AWSStatusSnapshot struct {
	Resources []AWSResource `json:"resources"`
}

type AWSResource struct {
	Tuple []string `json:"tuple"`
}
