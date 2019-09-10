// Copyright 2019 PingCAP, Inc.
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

// SelectorSpec defines the some selectors to select objects.
// If the all selectors are empty, all objects will be used in chaos experiment.
type SelectorSpec struct {
	// Namespaces is a set of namespace to which objects belong.
	// +optional
	Namespaces []string `json:"namespaces"`

	// Nodes is a set of node name and objects must belong to these nodes.
	// +optional
	Nodes []string `json:"nodes"`

	// Pods is a map of string keys and a set values that used to select pods.
	// The key defines the namespace which pods belong,
	// and the each values is a set of pod names.
	// +optional
	Pods map[string][]string `json:"pods"`

	// Map of string keys and values that can be used to select nodes.
	// Selector which must match a node's labels,
	// and objects must belong to these selected nodes.
	// +optional
	NodeSelectors map[string]string `json:"nodeSelectors"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on fields.
	// +optional
	FieldSelectors map[string]string `json:"fieldSelectors"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on labels.
	// +optional
	LabelSelectors map[string]string `json:"labelSelectors"`

	// Map of string keys and values that can be used to select objects.
	// A selector based on annotations.
	// +optional
	AnnotationSelectors map[string]string `json:"annotationSelectors"`
}

// SchedulerSpec defines information about schedule of the chaos experiment.
type SchedulerSpec struct {
	// Period two iterations of a specific chaos experiment.
	// This rule will overwrite the cron rule.
	Interval string `json:"interval"`

	// Cron defines a cron job rule.
	//
	// Some rule examples:
	// "0 30 * * * *" means to "Every hour on the half hour"
	// "@hourly"      means to "Every hour"
	// "@every 1h30m" means to "Every hour thirty"
	//
	// More rule info: https://godoc.org/github.com/robfig/cron
	Cron string `json:"cron"`
}
