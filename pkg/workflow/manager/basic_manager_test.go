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

package manager

import (
	"context"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/mock_manager"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

func TestScheduleSingleOne(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	// preparing mocks
	mockctl := gomock.NewController(t)
	mockRepo := mock_manager.NewMockWorkflowRepo(mockctl)

	const workflowName = "testing-workflow"
	const entryName = "just-a-entry"
	const entryNodeName = entryName + "-0000"
	const layer1Template0Name = "layer1-0"
	const layer1Node0Name = layer1Template0Name + "-0000"

	mockEntryTemplate := mock_template.NewMockSerialTemplate(mockctl)
	mockWorkflowSpec := mock_workflow.NewMockWorkflowSpec(mockctl)
	mockWorkflowStatus := mock_workflow.NewMockWorkflowStatus(mockctl)
	mockEntryNode := mock_node.NewMockNode(mockctl)
	mockEntryTreeNode := mock_node.NewMockNodeTreeNode(mockctl)
	mockEntryTreeNodeChildren := mock_node.NewMockNodeTreeChildren(mockctl)

	mockLayer1Template0 := mock_template.NewMockTemplate(mockctl)

	// mocked repo needs this trigger
	repoTrigger := trigger.NewOperableTrigger()
	mockRepo.EXPECT().FetchWorkflow(gomock.Eq(workflowName)).AnyTimes().Return(mockWorkflowSpec, mockWorkflowStatus, nil)

	// mocked methods
	gomock.InOrder(
		mockWorkflowStatus.EXPECT().FetchNodesMap().Return(nil).Times(1),
		mockWorkflowSpec.EXPECT().GetEntry().Return(entryName).Times(1),
		mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(entryName)).Return(mockEntryTemplate, nil).Times(1),
		mockRepo.EXPECT().CreateNodes(gomock.Eq(workflowName), gomock.Eq([]template.Template{mockEntryTemplate}), gomock.Eq("")).Return([]string{entryNodeName}, nil).Times(1),
		mockWorkflowStatus.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryNode, nil).Times(1),
		mockEntryNode.EXPECT().GetTemplateName().Return(entryName).Times(1),
		mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(entryName)).Return(mockEntryTemplate, nil).Times(1),
		mockEntryTemplate.EXPECT().GetTemplateType().Return(template.Serial).Times(1),
		mockRepo.EXPECT().UpdateNodesToWaitingForSchedule(gomock.Eq(workflowName), gomock.Eq(entryNodeName)).DoAndReturn(func(mockedWorkflowName, mockedEntryNodeName string) error {
			err := repoTrigger.Notify(trigger.NewEvent(mockedWorkflowName, mockedEntryNodeName, trigger.NodeWaitingForSchedule))
			return err
		}).Times(1),
		mockWorkflowStatus.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryNode, nil).Times(1),
		mockEntryNode.EXPECT().GetTemplateName().Return(entryName).Times(1),
		mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(entryName)).Return(mockEntryTemplate, nil).Times(1),
		mockEntryTemplate.EXPECT().GetTemplateType().Return(template.Serial).Times(1),
		mockWorkflowStatus.EXPECT().FetchNodesMap().Return(map[string]node.Node{
			entryNodeName: mockEntryNode,
		}).Times(1),
		mockWorkflowStatus.EXPECT().GetNodesTree().Return(mockEntryTreeNode).Times(1),
		mockEntryTreeNode.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryTreeNode).Times(1),
		mockEntryNode.EXPECT().GetTemplateName().Return(entryName).Times(1),
		mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(entryName)).Return(mockEntryTemplate, nil).Times(1),
		mockEntryTemplate.EXPECT().GetTemplateType().Return(template.Serial).Times(1),
		mockEntryTemplate.EXPECT().GetSerialChildrenList().Return([]template.Template{mockLayer1Template0}).Times(1),
		mockEntryTreeNode.EXPECT().GetChildren().Return(mockEntryTreeNodeChildren).Times(1),
		mockEntryTreeNodeChildren.EXPECT().Length().Return(0).Times(1),
		mockEntryTreeNode.EXPECT().GetChildren().Return(mockEntryTreeNodeChildren).Times(1),
		mockLayer1Template0.EXPECT().GetName().Return(layer1Template0Name).Times(1),
		mockEntryTreeNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(layer1Template0Name)).Return(false).Times(1),
		mockRepo.EXPECT().UpdateNodesToWaitingForChild(gomock.Eq(workflowName), gomock.Eq(entryNodeName)).DoAndReturn(func(mockedWorkflowName, mockedEntryNodeName string) error {
			// TODO trigger others
			return nil
		}).Times(1),
		mockRepo.EXPECT().CreateNodes(gomock.Eq(workflowName), gomock.Eq([]template.Template{mockLayer1Template0}), gomock.Eq(entryNodeName)).Return([]string{layer1Node0Name}, nil).Times(1).
			Do(func(_, _, _ interface{}) {
				wg.Done()
			}),
	)

	controllerTrigger := trigger.NewOperableTrigger()
	manager := NewBasicManager("testing-manager", mockRepo, repoTrigger, controllerTrigger)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		manager.Run(ctx)
	}()

	err := controllerTrigger.Notify(trigger.NewEvent(workflowName, "", trigger.WorkflowCreated))

	assert.NoError(t, err, "no error with notify")
	wg.Wait()
}
