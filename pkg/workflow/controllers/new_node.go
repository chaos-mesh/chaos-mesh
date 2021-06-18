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
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var (
	isController       = true
	blockOwnerDeletion = true
	ApiVersion         = v1alpha1.GroupVersion.String()
	KindWorkflow       = "Workflow"
	KindWorkflowNode   = "WorkflowNode"
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

			if template.Deadline != nil {
				duration, err := time.ParseDuration(*template.Deadline)
				if err != nil {
					// TODO: logger
					return nil, err
				}
				copiedDuration := metav1.NewTime(now.DeepCopy().Add(duration))
				deadline = &copiedDuration
			}

			renderedNode := v1alpha1.WorkflowNode{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    workflow.Namespace,
					GenerateName: fmt.Sprintf("%s-", template.Name),
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					TemplateName:        template.Name,
					WorkflowName:        workflow.Name,
					Type:                template.Type,
					StartTime:           &now,
					Deadline:            deadline,
					Children:            template.Children,
					Task:                template.Task,
					ConditionalBranches: template.ConditionalBranches,
					EmbedChaos:          template.EmbedChaos,
					Schedule:            conversionSchedule(template.Schedule),
				},
			}

			// if parent is specified, use parent as owner, else use workflow as owner.
			if parent != nil {
				renderedNode.OwnerReferences = append(renderedNode.OwnerReferences, metav1.OwnerReference{
					APIVersion:         ApiVersion,
					Kind:               KindWorkflowNode,
					Name:               parent.Name,
					UID:                parent.UID,
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				})
				if renderedNode.Labels == nil {
					renderedNode.Labels = make(map[string]string)
				}
				renderedNode.Labels[v1alpha1.LabelControlledBy] = parent.Name
			} else {
				renderedNode.OwnerReferences = append(renderedNode.OwnerReferences, metav1.OwnerReference{
					APIVersion:         ApiVersion,
					Kind:               KindWorkflow,
					Name:               workflow.Name,
					UID:                workflow.UID,
					Controller:         &isController,
					BlockOwnerDeletion: &blockOwnerDeletion,
				})
				if renderedNode.Labels == nil {
					renderedNode.Labels = make(map[string]string)
				}
				renderedNode.Labels[v1alpha1.LabelControlledBy] = workflow.Name
			}

			renderedNode.Labels[v1alpha1.LabelWorkflow] = workflow.Name
			renderedNode.Finalizers = append(renderedNode.Finalizers, metav1.FinalizerDeleteDependents)

			result = append(result, &renderedNode)
			continue
		}
		return nil, fmt.Errorf(
			"workflow %s do not contains template called %s",
			workflow.Name,
			name,
		)
	}
	return result, nil
}

func conversionSchedule(origin *v1alpha1.ChaosOnlyScheduleSpec) *v1alpha1.ScheduleSpec {
	if origin == nil {
		return nil
	}
	return &v1alpha1.ScheduleSpec{
		Schedule:                origin.Schedule,
		StartingDeadlineSeconds: origin.StartingDeadlineSeconds,
		ConcurrencyPolicy:       origin.ConcurrencyPolicy,
		HistoryLimit:            origin.HistoryLimit,
		Type:                    origin.Type,
		ScheduleItem: v1alpha1.ScheduleItem{
			EmbedChaos: v1alpha1.EmbedChaos{
				AwsChaos:     origin.EmbedChaos.AwsChaos,
				DNSChaos:     origin.EmbedChaos.DNSChaos,
				GcpChaos:     origin.EmbedChaos.GcpChaos,
				HTTPChaos:    origin.EmbedChaos.HTTPChaos,
				IOChaos:      origin.EmbedChaos.IOChaos,
				JVMChaos:     origin.EmbedChaos.JVMChaos,
				KernelChaos:  origin.EmbedChaos.KernelChaos,
				NetworkChaos: origin.EmbedChaos.NetworkChaos,
				PodChaos:     origin.EmbedChaos.PodChaos,
				StressChaos:  origin.EmbedChaos.StressChaos,
				TimeChaos:    origin.EmbedChaos.TimeChaos,
			},
			Workflow: nil,
		},
	}
}
