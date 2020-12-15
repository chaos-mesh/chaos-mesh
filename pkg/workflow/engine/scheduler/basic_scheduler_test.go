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
	assert.True(t, errors.Is(err, workflowerr.ErrNoSuchTemplate))
	assert.Equal(t, 0, len(nextTemplates))
	assert.Equal(t, "", parentNode)
}

func TestScheduleTheFirstEntry(t *testing.T) {
	workflowName := "mocked-workflow-0"
	entryTemplateName := "just-a-fake-entry"
	entryNodeName := entryTemplateName + "-0000"

	entryTemplate := mocktemplate.NewMockTemplate()
	entryTemplate.SetName(entryTemplateName)
	entryTemplate.SetTemplateType(template.IoChaos)
	entryTemplate.SetDuration(3 * time.Minute)

	mockedWorkflowSpec := mockworkflow.NewMockWorkflowSpec()
	mockedWorkflowSpec.SetName(workflowName)
	mockedWorkflowSpec.SetEntry(entryTemplateName)
	mockedWorkflowSpec.SetTemplates(mocktemplate.NewMockedTemplates([]template.Template{entryTemplate}))

	mockWorkflowStatus := mockworkflow.NewMockWorkflowStatus()

	basicScheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)

	// schedule entry
	nextTemplates, parentNode, err := basicScheduler.ScheduleNext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, entryTemplateName, nextTemplates[0].GetName())
	assert.Equal(t, "", parentNode)

	// update workflow status
	entryNode := mocknode.NewMockNode(
		entryNodeName,
		node.Succeed,
		"",
		entryTemplate.GetName(),
		entryTemplate.GetTemplateType())
	mockWorkflowStatus.SetNodes([]node.Node{entryNode})

	// the second schedule
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())
	assert.Error(t, err)
	assert.True(t, IsNoNeedSchedule(err))
	assert.Equal(t, "", parentNode)
	assert.Equal(t, 0, len(nextTemplates))
}

func TestNextedTwoLayerSerial(t *testing.T) {

	workflowName := "mocked-workflow-1"

	entryTemplateName := "another-entry"
	entryNodeName := entryTemplateName + "-0000"

	firstTemplateName := "normal-template-0"
	secondTemplateName := "normal-template-1"
	thirdTemplateName := "normal-template-2"

	firstNodeName := firstTemplateName + "-0000"
	secondNodeName := secondTemplateName + "-0000"
	thirdNodeName := thirdTemplateName + "-0000"

	firstTemplate := mocktemplate.NewMockTemplate()
	firstTemplate.SetName(firstTemplateName)
	firstTemplate.SetTemplateType(template.IoChaos)
	firstTemplate.SetDuration(5 * time.Minute)

	secondTemplate := mocktemplate.NewMockTemplate()
	secondTemplate.SetName(secondTemplateName)
	secondTemplate.SetTemplateType(template.Task)
	secondTemplate.SetDuration(1 * time.Minute)

	thirdTemplate := mocktemplate.NewMockTemplate()
	thirdTemplate.SetName(thirdTemplateName)
	thirdTemplate.SetTemplateType(template.Suspend)
	thirdTemplate.SetDuration(1 * time.Minute)

	entryTemplate := mocktemplate.NewMockSerialTemplate()
	entryTemplate.SetName(entryTemplateName)
	entryTemplate.SetTemplateType(template.Serial)
	entryTemplate.SetDeadline(20 * time.Minute)
	entryTemplate.SetSerialChildrenList([]template.Template{firstTemplate, secondTemplate, thirdTemplate})

	mockedWorkflowSpec := mockworkflow.NewMockWorkflowSpec()
	mockedWorkflowSpec.SetName(workflowName)
	mockedWorkflowSpec.SetEntry(entryTemplateName)
	mockedWorkflowSpec.SetTemplates(mocktemplate.NewMockedTemplates([]template.Template{entryTemplate}))

	mockWorkflowStatus := mockworkflow.NewMockWorkflowStatus()

	basicScheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)

	// schedule entry
	nextTemplates, parentNode, err := basicScheduler.ScheduleNext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, entryTemplateName, nextTemplates[0].GetName())
	assert.Equal(t, "", parentNode)

	// update workflow status
	entryNode := mocknode.NewMockNode(
		entryNodeName,
		node.WaitingForSchedule,
		"",
		entryTemplate.GetName(),
		entryTemplate.GetTemplateType())

	mockWorkflowStatus.SetNodes([]node.Node{entryNode})
	entryNodeTreeNode := mocknode.NewMockTreeNode(
		entryNode.GetName(),
		entryNode.GetTemplateName(),
		mocknode.NewMockNodeTreeChildren(nil))
	mockWorkflowStatus.SetRootNode(entryNodeTreeNode)

	// the second schedule, it expected to return the first child in Serial
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entryNodeName, parentNode)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, firstTemplateName, nextTemplates[0].GetName())

	// update status
	firstNode := mocknode.NewMockNode(
		firstNodeName,
		node.Succeed,
		entryNode.GetName(),
		firstTemplate.GetName(),
		firstTemplate.GetTemplateType(),
	)

	firstTreeNode := mocknode.NewMockTreeNode(
		firstNode.GetName(),
		firstNode.GetTemplateName(),
		mocknode.NewMockNodeTreeChildren(nil),
	)

	entryNodeTreeNode.SetChildren(mocknode.NewMockNodeTreeChildren(
		map[string]node.NodeTreeNode{
			firstNodeName: firstTreeNode,
		}))

	mockWorkflowStatus.SetNodes([]node.Node{entryNode, firstNode})

	// the third schedule, it expected to return the second child in Serial
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entryNodeName, parentNode)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, secondTemplateName, nextTemplates[0].GetName())

	// update status
	secondNode := mocknode.NewMockNode(
		secondNodeName,
		node.Succeed,
		entryNode.GetName(),
		secondTemplate.GetName(),
		secondTemplate.GetTemplateType(),
	)

	secondTreeNode := mocknode.NewMockTreeNode(
		secondNode.GetName(),
		secondNode.GetTemplateName(),
		mocknode.NewMockNodeTreeChildren(nil),
	)

	entryNodeTreeNode.SetChildren(mocknode.NewMockNodeTreeChildren(
		map[string]node.NodeTreeNode{
			firstNodeName:  firstTreeNode,
			secondNodeName: secondTreeNode,
		}))

	mockWorkflowStatus.SetNodes([]node.Node{entryNode, firstNode, secondNode})

	// the forth schedule, it expected to return the third child in Serial
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entryNodeName, parentNode)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, thirdTemplateName, nextTemplates[0].GetName())

	// update status

	thirdNode := mocknode.NewMockNode(
		thirdNodeName,
		node.Succeed,
		entryNode.GetName(),
		thirdTemplate.GetName(),
		thirdTemplate.GetTemplateType(),
	)

	thirdTreeNode := mocknode.NewMockTreeNode(
		thirdNode.GetName(),
		thirdNode.GetTemplateName(),
		mocknode.NewMockNodeTreeChildren(nil),
	)

	entryNodeTreeNode.SetChildren(mocknode.NewMockNodeTreeChildren(
		map[string]node.NodeTreeNode{
			firstNodeName:  firstTreeNode,
			secondNodeName: secondTreeNode,
			thirdNodeName:  thirdTreeNode,
		}))

	mockWorkflowStatus.SetNodes([]node.Node{entryNode, firstNode, secondNode, thirdNode})
	entryNode.SetPhase(node.Succeed)

	// the fifth schedule, it expected to return no need schedule error
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, workflowerr.ErrNoNeedSchedule))
}
