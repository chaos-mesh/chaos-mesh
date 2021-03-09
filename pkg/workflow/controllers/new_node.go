// Copyright 2021 Chaos Mesh Authors.
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

package controllers

import (
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var (
	isController       = true
	blockOwnerDeletion = true
)

// renderNodesByTemplates will render the nodes one by one, will setup owner by given parent. If parent is nil, it will use workflow as its owner.
func renderNodesByTemplates(workflow *v1alpha1.Workflow, parent *v1alpha1.WorkflowNode, templates ...string) ([]*v1alpha1.WorkflowNode, error) {
	templateNameSet := make(map[string]v1alpha1.Template)
	for _, template := range workflow.Spec.Templates {
		templateNameSet[template.Name] = template
	}
	var result []*v1alpha1.WorkflowNode
	for _, name := range templates {
		if template, ok := templateNameSet[name]; ok {

			now := metav1.NewTime(time.Now())
			var deadline *metav1.Time = nil

			if template.Duration != nil {
				duration, err := time.ParseDuration(*template.Duration)
				if err != nil {
					// TODO: logger
					return nil, err
				}
				result := metav1.NewTime(now.DeepCopy().Add(duration))
				deadline = &result
			}

			renderedNode := v1alpha1.WorkflowNode{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    workflow.Namespace,
					GenerateName: fmt.Sprintf("%s-", template.Name),
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					TemplateName: template.Name,
					WorkflowName: workflow.Name,
					Type:         template.Type,
					StartTime:    &now,
					Deadline:     deadline,
					Tasks:        template.Tasks,
					EmbedChaos:   template.EmbedChaos,
				},
			}

			if parent != nil {
				renderedNode.OwnerReferences = append(renderedNode.OwnerReferences, metav1.OwnerReference{
					APIVersion:         parent.APIVersion,
					Kind:               parent.Kind,
					Name:               parent.Name,
					UID:                parent.UID,
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				})
			} else {
				renderedNode.OwnerReferences = append(renderedNode.OwnerReferences, metav1.OwnerReference{
					APIVersion:         workflow.APIVersion,
					Kind:               workflow.Kind,
					Name:               workflow.Name,
					UID:                workflow.UID,
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				})
			}

			renderedNode.Finalizers = append(renderedNode.Finalizers, metav1.FinalizerDeleteDependents)

			result = append(result, &renderedNode)
			continue
		}
		return nil, errors.New(
			fmt.Sprintf("workflow %s do not contains template claled %s",
				workflow.Name,
				name,
			))
	}
	return result, nil
}
