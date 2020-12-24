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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	workflowerr "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_workflow"
)

func TestEmptySpec(t *testing.T) {
	mockctl := gomock.NewController(t)
	defer mockctl.Finish()

	mockedTemplates := mock_template.NewMockTemplates(mockctl)
	mockedTemplates.EXPECT().GetTemplateByName(gomock.Any()).Return(nil, workflowerr.ErrNoSuchTemplate)

	mockedWorkflowSpec := mock_workflow.NewMockWorkflowSpec(mockctl)
	mockedWorkflowSpec.EXPECT().GetTemplates().Return(mockedTemplates, nil)
	mockedWorkflowSpec.EXPECT().GetEntry().Return("")

	mockWorkflowStatus := mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().FetchNodesMap().Return(nil)
	scheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err := scheduler.ScheduleNext(context.Background())
	assert.Error(t, err, "must be failed")
	assert.True(t, errors.Is(err, workflowerr.ErrNoSuchTemplate))
	assert.Equal(t, 0, len(nextTemplates))
	assert.Equal(t, "", parentNode)
}

func TestScheduleTheFirstEntry(t *testing.T) {
	mockctl := gomock.NewController(t)
	defer mockctl.Finish()

	workflowName := "mocked-workflow-0"
	entryTemplateName := "just-a-fake-entry"
	entryNodeName := entryTemplateName + "-0000"

	entryTemplate := mock_template.NewMockTemplate(mockctl)
	entryTemplate.EXPECT().GetName().AnyTimes().Return(entryTemplateName)
	entryTemplate.EXPECT().GetTemplateType().Return(template.IoChaos)

	mockTemplates := mock_template.NewMockTemplates(mockctl)
	mockTemplates.EXPECT().GetTemplateByName(gomock.Eq(entryTemplateName)).AnyTimes().Return(entryTemplate, nil)

	mockedWorkflowSpec := mock_workflow.NewMockWorkflowSpec(mockctl)
	mockedWorkflowSpec.EXPECT().GetName().Return(workflowName)
	mockedWorkflowSpec.EXPECT().GetEntry().AnyTimes().Return(entryTemplate.GetName())
	mockedWorkflowSpec.EXPECT().GetTemplates().AnyTimes().Return(mockTemplates, nil)

	mockWorkflowStatus := mock_workflow.NewMockWorkflowStatus(mockctl)

	basicScheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	mockWorkflowStatus.EXPECT().FetchNodesMap().AnyTimes().Return(map[string]node.Node{})

	// schedule entry
	nextTemplates, parentNode, err := basicScheduler.ScheduleNext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, entryTemplateName, nextTemplates[0].GetName())
	assert.Equal(t, "", parentNode)

	// update workflow status
	entryNode := mock_node.NewMockNode(mockctl)
	entryNode.EXPECT().GetName().AnyTimes().Return(entryNodeName)
	entryNode.EXPECT().GetTemplateType().AnyTimes().Return(entryTemplate.GetTemplateType())
	entryNode.EXPECT().GetNodePhase().Return(node.Succeed)

	mockWorkflowStatus = mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().GetNodes().Return([]node.Node{entryNode})
	mockWorkflowStatus.EXPECT().FetchNodesMap().AnyTimes().Return(map[string]node.Node{
		entryNode.GetName(): entryNode,
	})

	// the second schedule
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())
	assert.Error(t, err)
	assert.True(t, IsNoNeedSchedule(err))
	assert.Equal(t, "", parentNode)
	assert.Equal(t, 0, len(nextTemplates))
}

func TestNextedTwoLayerSerial(t *testing.T) {
	mockctl := gomock.NewController(t)
	defer mockctl.Finish()

	workflowName := "mocked-workflow-1"

	entryTemplateName := "another-entry"
	entryNodeName := entryTemplateName + "-0000"

	firstTemplateName := "normal-template-0"
	secondTemplateName := "normal-template-1"
	thirdTemplateName := "normal-template-2"

	firstNodeName := firstTemplateName + "-0000"
	secondNodeName := secondTemplateName + "-0000"
	thirdNodeName := thirdTemplateName + "-0000"

	firstTemplate := mock_template.NewMockTemplate(mockctl)
	firstTemplate.EXPECT().GetName().AnyTimes().Return(firstTemplateName)

	secondTemplate := mock_template.NewMockTemplate(mockctl)
	secondTemplate.EXPECT().GetName().AnyTimes().Return(secondTemplateName)

	thirdTemplate := mock_template.NewMockTemplate(mockctl)
	thirdTemplate.EXPECT().GetName().AnyTimes().Return(thirdTemplateName)

	entryTemplate := mock_template.NewMockSerialTemplate(mockctl)
	entryTemplate.EXPECT().GetName().AnyTimes().Return(entryTemplateName)
	entryTemplate.EXPECT().GetTemplateType().Return(template.Serial)
	entryTemplate.EXPECT().GetSerialChildrenList().AnyTimes().Return([]template.Template{firstTemplate, secondTemplate, thirdTemplate})

	mockedTemplates := mock_template.NewMockTemplates(mockctl)
	mockedTemplates.EXPECT().GetTemplateByName(gomock.Eq(entryTemplate.GetName())).AnyTimes().Return(entryTemplate, nil)

	mockedWorkflowSpec := mock_workflow.NewMockWorkflowSpec(mockctl)
	mockedWorkflowSpec.EXPECT().GetName().Return(workflowName)
	mockedWorkflowSpec.EXPECT().GetEntry().AnyTimes().Return(entryTemplate.GetName())
	mockedWorkflowSpec.EXPECT().GetTemplates().AnyTimes().Return(mockedTemplates, nil)

	mockWorkflowStatus := mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().FetchNodesMap().Return(nil)

	basicScheduler := NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)

	// schedule entry
	nextTemplates, parentNode, err := basicScheduler.ScheduleNext(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, entryTemplateName, nextTemplates[0].GetName())
	assert.Equal(t, "", parentNode)

	entryNode := mock_node.NewMockNode(mockctl)
	entryNode.EXPECT().GetName().AnyTimes().Return(entryNodeName)
	entryNode.EXPECT().GetNodePhase().AnyTimes().Return(node.WaitingForSchedule)
	entryNode.EXPECT().GetTemplateName().AnyTimes().Return(entryTemplate.GetName())
	entryNode.EXPECT().GetTemplateType().AnyTimes().Return(entryTemplate.GetTemplateType())

	mockedEntryNodeTreeNode := mock_node.NewMockNodeTreeNode(mockctl)

	// emptyTreeChildren is dummy var
	emptyTreeChildren := mock_node.NewMockNodeTreeChildren(mockctl)
	emptyTreeChildren.EXPECT().Length().AnyTimes().Return(0)
	emptyTreeChildren.EXPECT().ContainsTemplate(gomock.Any()).AnyTimes().Return(false)

	entryNodeChildren := emptyTreeChildren
	mockedEntryNodeTreeNode.EXPECT().GetChildren().AnyTimes().Return(entryNodeChildren)
	mockedEntryNodeTreeNode.EXPECT().FetchNodeByName(gomock.Eq(entryNode.GetName())).AnyTimes().Return(mockedEntryNodeTreeNode)

	mockWorkflowStatus = mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().GetNodes().Return([]node.Node{entryNode})
	mockWorkflowStatus.EXPECT().GetNodesTree().AnyTimes().Return(mockedEntryNodeTreeNode)
	mockWorkflowStatus.EXPECT().FetchNodesMap().AnyTimes().Return(map[string]node.Node{
		entryNode.GetName(): entryNode,
	})

	// the second schedule, it expected to return the first child in Serial
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entryNodeName, parentNode)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, firstTemplateName, nextTemplates[0].GetName())

	// update status
	firstNode := mock_node.NewMockNode(mockctl)
	firstNode.EXPECT().GetName().AnyTimes().Return(firstNodeName)
	firstNode.EXPECT().GetNodePhase().Return(node.Succeed)

	entryNodeChildren = mock_node.NewMockNodeTreeChildren(mockctl)
	entryNodeChildren.EXPECT().Length().Return(1)
	entryNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(firstTemplate.GetName())).Return(true)
	entryNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(secondTemplate.GetName())).Return(false)

	mockedEntryNodeTreeNode = mock_node.NewMockNodeTreeNode(mockctl)
	mockedEntryNodeTreeNode.EXPECT().GetChildren().AnyTimes().Return(entryNodeChildren)
	mockedEntryNodeTreeNode.EXPECT().FetchNodeByName(gomock.Eq(entryNode.GetName())).AnyTimes().Return(mockedEntryNodeTreeNode)

	mockWorkflowStatus = mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().GetNodes().Return([]node.Node{entryNode, firstNode})
	mockWorkflowStatus.EXPECT().GetNodesTree().Return(mockedEntryNodeTreeNode)
	mockWorkflowStatus.EXPECT().FetchNodesMap().AnyTimes().Return(map[string]node.Node{
		entryNode.GetName(): entryNode,
		firstNode.GetName(): firstNode,
	})

	// the third schedule, it expected to return the second child in Serial
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entryNodeName, parentNode)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, secondTemplateName, nextTemplates[0].GetName())

	// update status
	secondNode := mock_node.NewMockNode(mockctl)
	secondNode.EXPECT().GetName().AnyTimes().Return(secondNodeName)
	secondNode.EXPECT().GetNodePhase().Return(node.Succeed)

	entryNodeChildren = mock_node.NewMockNodeTreeChildren(mockctl)
	entryNodeChildren.EXPECT().Length().Return(2)
	entryNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(firstTemplate.GetName())).Return(true)
	entryNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(secondTemplate.GetName())).Return(true)
	entryNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(thirdTemplate.GetName())).Return(false)

	mockedEntryNodeTreeNode = mock_node.NewMockNodeTreeNode(mockctl)
	mockedEntryNodeTreeNode.EXPECT().GetChildren().AnyTimes().Return(entryNodeChildren)
	mockedEntryNodeTreeNode.EXPECT().FetchNodeByName(gomock.Eq(entryNode.GetName())).AnyTimes().Return(mockedEntryNodeTreeNode)

	mockWorkflowStatus = mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().GetNodes().Return([]node.Node{entryNode, firstNode, secondNode})
	mockWorkflowStatus.EXPECT().GetNodesTree().Return(mockedEntryNodeTreeNode)
	mockWorkflowStatus.EXPECT().FetchNodesMap().AnyTimes().Return(map[string]node.Node{
		entryNode.GetName():  entryNode,
		firstNode.GetName():  firstNode,
		secondNode.GetName(): secondNode,
	})

	// the forth schedule, it expected to return the third child in Serial
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, entryNodeName, parentNode)
	assert.Equal(t, 1, len(nextTemplates))
	assert.Equal(t, thirdTemplateName, nextTemplates[0].GetName())

	// update status
	thirdNode := mock_node.NewMockNode(mockctl)
	thirdNode.EXPECT().GetName().AnyTimes().Return(thirdNodeName)
	thirdNode.EXPECT().GetNodePhase().Return(node.Succeed)

	//thirdTreeNode := mock_node.NewMockNodeTreeNode(mockctl)

	entryNodeChildren = mock_node.NewMockNodeTreeChildren(mockctl)

	mockedEntryNodeTreeNode = mock_node.NewMockNodeTreeNode(mockctl)
	mockedEntryNodeTreeNode.EXPECT().GetChildren().AnyTimes().Return(entryNodeChildren)
	mockedEntryNodeTreeNode.EXPECT().FetchNodeByName(gomock.Eq(entryNode.GetName())).AnyTimes().Return(mockedEntryNodeTreeNode)

	entryNode = mock_node.NewMockNode(mockctl)
	entryNode.EXPECT().GetName().AnyTimes().Return(entryNodeName)
	entryNode.EXPECT().GetNodePhase().Return(node.Succeed)

	mockWorkflowStatus = mock_workflow.NewMockWorkflowStatus(mockctl)
	mockWorkflowStatus.EXPECT().GetNodes().Return([]node.Node{entryNode, firstNode, secondNode, thirdNode})
	mockWorkflowStatus.EXPECT().FetchNodesMap().AnyTimes().Return(map[string]node.Node{
		entryNode.GetName():  entryNode,
		firstNode.GetName():  firstNode,
		secondNode.GetName(): secondNode,
		thirdNode.GetName():  thirdNode,
	})

	// the fifth schedule, it expected to return no need schedule error
	basicScheduler = NewBasicScheduler(mockedWorkflowSpec, mockWorkflowStatus)
	nextTemplates, parentNode, err = basicScheduler.ScheduleNext(context.Background())

	assert.Error(t, err)
	assert.True(t, errors.Is(err, workflowerr.ErrNoNeedSchedule))
}
