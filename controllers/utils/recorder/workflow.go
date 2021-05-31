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
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type WorkflowInvalidEntry struct {
	EntryTemplate string
}

func (it WorkflowInvalidEntry) Type() string {
	return corev1.EventTypeWarning
}

func (it WorkflowInvalidEntry) Reason() string {
	return v1alpha1.InvalidEntry
}

func (it WorkflowInvalidEntry) Message() string {
	return fmt.Sprintf("failed to spawn new entry node of workflow, entry: %s", it.EntryTemplate)
}

type WorkflowEntryCreated struct {
	Entry string
}

func (it WorkflowEntryCreated) Type() string {
	return corev1.EventTypeNormal
}

func (it WorkflowEntryCreated) Reason() string {
	return v1alpha1.EntryCreated
}

func (it WorkflowEntryCreated) Message() string {
	return fmt.Sprintf("entry node created, entry node %s", it.Entry)
}
