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

package recorder

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type InvalidEntry struct {
	EntryTemplate string
}

func (it InvalidEntry) Type() string {
	return corev1.EventTypeWarning
}

func (it InvalidEntry) Reason() string {
	return v1alpha1.InvalidEntry
}

func (it InvalidEntry) Message() string {
	return fmt.Sprintf("failed to spawn new entry node of workflow, entry: %s", it.EntryTemplate)
}

type EntryCreated struct {
	Entry string
}

func (it EntryCreated) Type() string {
	return corev1.EventTypeNormal
}

func (it EntryCreated) Reason() string {
	return v1alpha1.EntryCreated
}

func (it EntryCreated) Message() string {
	return fmt.Sprintf("entry node created, entry node %s", it.Entry)
}

type NodesCreated struct {
	ChildNodes []string
}

func (it NodesCreated) Type() string {
	return corev1.EventTypeNormal
}

func (it NodesCreated) Reason() string {
	return v1alpha1.NodesCreated
}

func (it NodesCreated) Message() string {
	return fmt.Sprintf("child nodes created, %s", strings.Join(it.ChildNodes, ","))
}

type ChaosCustomResourceCreated struct {
	Name string
	Kind string
}

func (it ChaosCustomResourceCreated) Type() string {
	return corev1.EventTypeNormal
}

func (it ChaosCustomResourceCreated) Reason() string {
	return v1alpha1.ChaosCRCreated
}

func (it ChaosCustomResourceCreated) Message() string {
	return fmt.Sprintf("chaos CR %s created", it.Name)
}

type ChaosCustomResourceCreateFailed struct {
}

func (it ChaosCustomResourceCreateFailed) Type() string {
	return corev1.EventTypeWarning
}

func (it ChaosCustomResourceCreateFailed) Reason() string {
	return v1alpha1.ChaosCRCreateFailed
}

func (it ChaosCustomResourceCreateFailed) Message() string {
	return "failed to create chaos CR"
}

type ChaosCustomResourceDeleted struct {
	Name string
	Kind string
}

func (it ChaosCustomResourceDeleted) Type() string {
	return corev1.EventTypeNormal
}

func (it ChaosCustomResourceDeleted) Reason() string {
	return v1alpha1.ChaosCRDeleted
}

func (it ChaosCustomResourceDeleted) Message() string {
	return fmt.Sprintf("chaos CR %s deleted", it.Name)
}

type ChaosCustomResourceDeleteFailed struct {
	Name string
	Kind string
}

func (it ChaosCustomResourceDeleteFailed) Type() string {
	return corev1.EventTypeWarning
}

func (it ChaosCustomResourceDeleteFailed) Reason() string {
	return v1alpha1.ChaosCRDeleteFailed
}

func (it ChaosCustomResourceDeleteFailed) Message() string {
	return fmt.Sprintf("chaos CR %s delete failed", it.Name)
}

type DeadlineExceed struct {
}

func (it DeadlineExceed) Type() string {
	return corev1.EventTypeNormal
}

func (it DeadlineExceed) Reason() string {
	return v1alpha1.NodeDeadlineExceed
}

func (it DeadlineExceed) Message() string {
	return "deadline exceed"
}

type ParentNodeDeadlineExceed struct {
	ParentNodeName string
}

func (it ParentNodeDeadlineExceed) Type() string {
	return corev1.EventTypeNormal
}

func (it ParentNodeDeadlineExceed) Reason() string {
	return v1alpha1.ParentNodeDeadlineExceed
}

func (it ParentNodeDeadlineExceed) Message() string {
	return fmt.Sprintf("deadline exceed bscause parent node %s deadline exceed", it.ParentNodeName)
}

type WorkflowAccomplished struct {
}

func (it WorkflowAccomplished) Type() string {
	return corev1.EventTypeNormal
}

func (it WorkflowAccomplished) Reason() string {
	return v1alpha1.WorkflowAccomplished
}

func (it WorkflowAccomplished) Message() string {
	return "workflow accomplished"
}

type NodeAccomplished struct {
}

func (it NodeAccomplished) Type() string {
	return corev1.EventTypeNormal
}

func (it NodeAccomplished) Reason() string {
	return v1alpha1.NodeAccomplished
}

func (it NodeAccomplished) Message() string {
	return "node accomplished"
}

type TaskPodSpawned struct {
	PodName string
}

func (it TaskPodSpawned) Type() string {
	return corev1.EventTypeNormal
}

func (it TaskPodSpawned) Reason() string {
	return v1alpha1.TaskPodSpawned
}

func (it TaskPodSpawned) Message() string {
	return fmt.Sprintf("pod %s spawned for task", it.PodName)
}

type TaskPodSpawnFailed struct {
}

func (it TaskPodSpawnFailed) Type() string {
	return corev1.EventTypeWarning
}

func (it TaskPodSpawnFailed) Reason() string {
	return v1alpha1.TaskPodSpawnFailed
}

func (it TaskPodSpawnFailed) Message() string {
	return "failed to create pod for task"
}

type TaskPodPodCompleted struct {
	PodName string
}

func (it TaskPodPodCompleted) Type() string {
	return corev1.EventTypeNormal
}

func (it TaskPodPodCompleted) Reason() string {
	return v1alpha1.TaskPodPodCompleted
}

func (it TaskPodPodCompleted) Message() string {
	return fmt.Sprintf("pod %s for task node completed", it.PodName)
}

type ConditionalBranchesSelected struct {
	SelectedBranches []string
}

func (it ConditionalBranchesSelected) Type() string {
	return corev1.EventTypeNormal
}

func (it ConditionalBranchesSelected) Reason() string {
	return v1alpha1.ConditionalBranchesSelected
}

func (it ConditionalBranchesSelected) Message() string {
	return fmt.Sprintf("selected branches: %s", it.SelectedBranches)
}

type RerunBySpecChanged struct {
	CleanedChildrenNode []string
}

func (it RerunBySpecChanged) Type() string {
	return corev1.EventTypeNormal
}

func (it RerunBySpecChanged) Reason() string {
	return v1alpha1.RerunBySpecChanged
}

func (it RerunBySpecChanged) Message() string {
	return fmt.Sprintf("rerun by spec changed, remove children nodes: %s", it.CleanedChildrenNode)
}

func init() {
	register(
		InvalidEntry{},
		EntryCreated{},
		NodesCreated{},
		ChaosCustomResourceCreated{},
		ChaosCustomResourceCreateFailed{},
		ChaosCustomResourceDeleted{},
		ChaosCustomResourceDeleteFailed{},
		DeadlineExceed{},
		ParentNodeDeadlineExceed{},
		WorkflowAccomplished{},
		NodeAccomplished{},
		TaskPodSpawned{},
		TaskPodSpawnFailed{},
		TaskPodPodCompleted{},
		ConditionalBranchesSelected{},
		RerunBySpecChanged{},
	)
}
