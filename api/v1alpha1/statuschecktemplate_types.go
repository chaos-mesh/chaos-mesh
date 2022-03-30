// Copyright Chaos Mesh Authors.
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

package v1alpha1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

const (
	// TemplateTypeLabelKey is a label that represents the template type.
	TemplateTypeLabelKey = "template.chaos-mesh.org/type"
	// ManagedByLabelKey is a label that represents the tool being used
	// to manage the operation of the object.
	ManagedByLabelKey = "app.kubernetes.io/managed-by"
	// ManagedByLabelValue is the value that represents the object is
	// managed by Chaos Mesh.
	ManagedByLabelValue = "chaos-mesh"

	// TemplateNameAnnotationKey is an annotation that represents
	// the real name of the template.
	TemplateNameAnnotationKey = "template.chaos-mesh.org/name"
	// TemplateDescriptionAnnotationKey is an annotation that represents
	// the description of the template.
	TemplateDescriptionAnnotationKey = "template.chaos-mesh.org/description"

	// PrefixStatusCheckTemplate is the prefix of the name of a StatusCheckTemplate.
	PrefixStatusCheckTemplate = "template-status-check"
	// StatusCheckTemplateKey is the key that status check spec
	// saved in the template ConfigMap.
	StatusCheckTemplateKey = "spec"
)

// StatusCheckTemplate represents a template of status check.
// A statusCheckTemplate would save in the ConfigMap named `template-status-check-<template-name>`.
// +kubebuilder:object:generate=false
type StatusCheckTemplate struct {
	StatusCheckSpec `json:",inline"`
}

func GetTemplateName(cm v1.ConfigMap) string {
	return cm.Annotations[TemplateNameAnnotationKey]
}

func GetTemplateDescription(cm v1.ConfigMap) string {
	return cm.Annotations[TemplateDescriptionAnnotationKey]
}

func GenerateTemplateName(name string) string {
	return fmt.Sprintf("%s-%s", PrefixStatusCheckTemplate, name)
}

func IsStatusCheckTemplate(cm v1.ConfigMap) bool {
	return cm.Labels[ManagedByLabelKey] == ManagedByLabelValue &&
		cm.Labels[TemplateTypeLabelKey] == KindStatusCheck &&
		cm.Name == GenerateTemplateName(cm.Annotations[TemplateNameAnnotationKey])
}

func (in *StatusCheckTemplate) Validate() error {
	statusCheck := &StatusCheck{
		Spec: in.StatusCheckSpec,
	}
	return statusCheck.Validate()
}

func (in *StatusCheckTemplate) Default() {
	if in == nil {
		return
	}

	statusCheck := &StatusCheck{
		Spec: in.StatusCheckSpec,
	}
	statusCheck.Default()
	in.StatusCheckSpec = *statusCheck.Spec.DeepCopy()
}
