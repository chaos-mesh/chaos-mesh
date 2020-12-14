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

package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	workflowerr "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	mocknode "github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/node"
	mocktemplate "github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/template"
	mockworkflow "github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/workflow"
)

func TestEmptySpec(t *testing.T) {
	mockedWorkflowSpec := mockworkflow.NewMockWorkflowSpec()
	mockWorkflowStatus := mockworkflow.NewMockWorkflowStatus()
	scheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err := scheduler.ScheduleNext(context.Background())
	assert.Error(t, err, "must be failed")
	assert.True(t, errors.Is(err, workflowerr.ErrNoTemplates))
	assert.Equal(t, 0, len(nextTemplates))
	assert.Equal(t, "", parentNode)
}

func TestScheduleTheFirstEntry(t *testing.T) {
	workflowName := "mocked-workflow-0"
	entryName := "just-a-fake-entry"
	nodeName := entryName + "-0000"

	entryTemplate := mocktemplate.NewMockTemplate()
	entryTemplate.SetName(entryName)
	entryTemplate.SetTemplateType(template.IoChaos)
	entryTemplate.SetDuration(3 * time.Minute)

	mockedWorkflowSpec := mockworkflow.NewMockWorkflowSpec()
	mockedWorkflowSpec.SetName(workflowName)
	mockedWorkflowSpec.SetEntry(entryName)
	mockedWorkflowSpec.SetTemplates(mocktemplate.NewMockedTemplates([]template.Template{entryTemplate}))

	mockWorkflowStatus := mockworkflow.NewMockWorkflowStatus()

	basicScheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)

	// schedule entry
	nextTemplates, parentNode, err := basicScheduler.ScheduleNext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, entryName, nextTemplates[0].GetName())
	assert.Equal(t, "", parentNode)

	// update workflow status
	entryNode := mocknode.NewMockNode()
	entryNode.SetNodeName(nodeName)
	entryNode.SetPhase(node.Succeed)
	mockWorkflowStatus.SetNodes([]node.Node{entryNode})

	// the second schedule
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())
	assert.Error(t, err)
	assert.True(t, IsNoNeedSchedule(err))
	assert.Equal(t, "", parentNode)
	assert.Equal(t, 0, len(nextTemplates))
}
