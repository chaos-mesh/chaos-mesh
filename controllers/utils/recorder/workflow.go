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

func init() {
	register(
		InvalidEntry{},
		EntryCreated{},
		NodesCreated{},
		ChaosCustomResourceCreated{},
		ChaosCustomResourceCreateFailed{},
		DeadlineExceed{},
		WorkflowAccomplished{},
	)
}
