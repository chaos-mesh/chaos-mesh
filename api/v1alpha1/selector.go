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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// LabelSelectorRequirements is list of LabelSelectorRequirement
type LabelSelectorRequirements []metav1.LabelSelectorRequirement

// PodMode represents the mode to run pod chaos action.
type PodMode string

const (
	// OnePodMode represents that the system will do the chaos action on one pod selected randomly.
	OnePodMode PodMode = "one"
	// AllPodMode represents that the system will do the chaos action on all pods
	// regardless of status (not ready or not running pods includes).
	// Use this label carefully.
	AllPodMode PodMode = "all"
	// FixedPodMode represents that the system will do the chaos action on a specific number of running pods.
	FixedPodMode PodMode = "fixed"
	// FixedPercentPodMode to specify a fixed % that can be inject chaos action.
	FixedPercentPodMode PodMode = "fixed-percent"
	// RandomMaxPercentPodMode to specify a maximum % that can be inject chaos action.
	RandomMaxPercentPodMode PodMode = "random-max-percent"
)

// PodSelectorSpec defines the some selectors to select objects.
// If the all selectors are empty, all objects will be used in chaos experiment.
type PodSelectorSpec struct {
	// Namespaces is a set of namespace to which objects belong.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// Nodes is a set of node name and objects must belong to these nodes.
	// +optional
	Nodes []string `json:"nodes,omitempty"`

	// Pods is a map of string keys and a set values that used to select pods.
	// The key defines the namespace which pods belong,
	// and the each values is a set of pod names.
	// +optional
	Pods map[string][]string `json:"pods,omitempty"`

	// Map of string keys and values that can be used to select nodes.
	// Selector which must match a node's labels,
	// and objects must belong to these selected nodes.
	// +optional
	NodeSelectors map[string]string `json:"nodeSelectors,omitempty"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on fields.
	// +optional
	FieldSelectors map[string]string `json:"fieldSelectors,omitempty"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on labels.
	// +optional
	LabelSelectors map[string]string `json:"labelSelectors,omitempty"`

	// a slice of label selector expressions that can be used to select objects.
	// A list of selectors based on set-based label expressions.
	// +optional
	ExpressionSelectors LabelSelectorRequirements `json:"expressionSelectors,omitempty"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on annotations.
	// +optional
	AnnotationSelectors map[string]string `json:"annotationSelectors,omitempty"`

	// PodPhaseSelectors is a set of condition of a pod at the current time.
	// supported value: Pending / Running / Succeeded / Failed / Unknown
	// +optional
	PodPhaseSelectors []string `json:"podPhaseSelectors,omitempty"`
}

func (in *PodSelectorSpec) DefaultNamespace(namespace string) {
	if len(in.Namespaces) == 0 {
		in.Namespaces = []string{namespace}
	}
}

type PodSelector struct {
	// Selector is used to select pods that are used to inject chaos action.
	Selector PodSelectorSpec `json:"selector"`

	// Mode defines the mode to run chaos action.
	// Supported mode: one / all / fixed / fixed-percent / random-max-percent
	// +kubebuilder:validation:Enum=one;all;fixed;fixed-percent;random-max-percent
	Mode PodMode `json:"mode"`

	// Value is required when the mode is set to `FixedPodMode` / `FixedPercentPodMod` / `RandomMaxPercentPodMod`.
	// If `FixedPodMode`, provide an integer of pods to do chaos action.
	// If `FixedPercentPodMod`, provide a number from 0-100 to specify the percent of pods the server can do chaos action.
	// IF `RandomMaxPercentPodMod`,  provide a number from 0-100 to specify the max percent of pods to do chaos action
	// +optional
	Value string `json:"value,omitempty"`
}

type ContainerSelector struct {
	PodSelector `json:",inline"`

	// ContainerNames indicates list of the name of affected container.
	// If not set, all containers will be injected
	// +optional
	ContainerNames []string `json:"containerNames,omitempty"`
}

// ClusterScoped returns true if the selector selects Pods in the cluster
func (in PodSelectorSpec) ClusterScoped() bool {
	// in fact, this will never happened, will add namespace if it is empty, so len(s.Namespaces) can not be 0,
	// but still add judgentment here for safe
	// https://github.com/chaos-mesh/chaos-mesh/blob/478d00d01bb0f9fb08a1085428a7da8c8f9df4e8/api/v1alpha1/common_webhook.go#L22
	if len(in.Namespaces) == 0 && len(in.Pods) == 0 {
		return true
	}

	return false
}

// AffectedNamespaces returns all the namespaces which the selector effect
func (in PodSelectorSpec) AffectedNamespaces() []string {
	affectedNamespacesMap := make(map[string]struct{})
	affectedNamespacesArray := make([]string, 0, 2)

	for namespace := range in.Pods {
		affectedNamespacesMap[namespace] = struct{}{}
	}

	for _, namespace := range in.Namespaces {
		affectedNamespacesMap[namespace] = struct{}{}
	}

	for namespace := range affectedNamespacesMap {
		affectedNamespacesArray = append(affectedNamespacesArray, namespace)
	}

	return affectedNamespacesArray
}
