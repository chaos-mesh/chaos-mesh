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
	"testing"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect/resolver"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/mock_workflowrepo"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_template"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/mock/engine/model/mock_workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

func TestScheduleSingleOne(t *testing.T) {

	// preparing mocks
	mockctl := gomock.NewController(t)
	mockRepo := mock_workflowrepo.NewMockWorkflowRepo(mockctl)

	const namespace = "mock-ns"
	const workflowName = "testing-workflow"
	const entryName = "just-a-entry"
	const entryNodeName = entryName + "-0000"
	const layer1Template0Name = "layer1-0"
	const layer1Node0Name = layer1Template0Name + "-0000"

	mockNameGenerator := mock_node.NewMockNodeNameGenerator(mockctl)
	mockEntryTemplate := mock_template.NewMockSerialTemplate(mockctl)
	mockWorkflowSpec := mock_workflow.NewMockWorkflowSpec(mockctl)
	mockWorkflowStatus := mock_workflow.NewMockWorkflowStatus(mockctl)
	mockEntryTreeNode := mock_node.NewMockNodeTreeNode(mockctl)
	mockEntryTreeNodeChildren := mock_node.NewMockNodeTreeChildren(mockctl)

	mockLayer1Template0 := mock_template.NewMockSuspendTemplate(mockctl)

	// mocked repo needs this trigger
	mockRepo.EXPECT().FetchWorkflow(gomock.Eq(namespace), gomock.Eq(workflowName)).AnyTimes().Return(mockWorkflowSpec, mockWorkflowStatus, nil)

	mockNameGenerator.EXPECT().GenerateNodeName(gomock.Eq(entryName)).Return(entryNodeName).AnyTimes()
	mockNameGenerator.EXPECT().GenerateNodeName(gomock.Eq(layer1Template0Name)).Return(layer1Node0Name).AnyTimes()

	mockWorkflowSpec.EXPECT().GetName().Return(workflowName).AnyTimes()
	mockWorkflowSpec.EXPECT().GetEntry().Return(entryName).AnyTimes()
	mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(entryName)).Return(mockEntryTemplate, nil).AnyTimes()
	mockWorkflowSpec.EXPECT().FetchTemplateByName(gomock.Eq(layer1Template0Name)).Return(mockLayer1Template0, nil).AnyTimes()

	mockEntryTemplate.EXPECT().GetTemplateType().Return(template.Serial).AnyTimes()
	mockEntryTemplate.EXPECT().GetSerialChildrenList().Return([]template.Template{mockLayer1Template0}).AnyTimes()

	mockLayer1Template0.EXPECT().GetName().Return(layer1Template0Name).AnyTimes()
	mockLayer1Template0.EXPECT().GetTemplateType().Return(template.Suspend).AnyTimes()

	mockEntryNode := mock_node.NewMockNode(mockctl)
	mockEntryNode.EXPECT().GetName().Return(entryNodeName).AnyTimes()
	mockEntryNode.EXPECT().GetTemplateName().Return(entryName).AnyTimes()
	gomock.InOrder(
		mockEntryNode.EXPECT().GetNodePhase().Return(node.Init).Times(2),
	)
	mockEntryNodeWaitingForSchedule := mock_node.NewMockNode(mockctl)
	mockEntryNodeWaitingForSchedule.EXPECT().GetName().Return(entryNodeName).AnyTimes()
	mockEntryNodeWaitingForSchedule.EXPECT().GetTemplateName().Return(entryName).AnyTimes()
	gomock.InOrder(
		mockEntryNodeWaitingForSchedule.EXPECT().GetNodePhase().Return(node.WaitingForSchedule).Times(2),
	)

	layer1Node0Node := mock_node.NewMockNode(mockctl)

	repoTrigger := trigger.NewOperableTrigger()
	compositeResolver, err := resolver.NewCompositeResolverWith(resolver.NewNotifyNewEventResolver(repoTrigger), resolver.NewCreateNewNodeResolver(mockRepo), resolver.NewUpdateNodePhaseResolver(mockRepo))
	if err != nil {
		t.Fatal(err)
	}
	controllerTrigger := trigger.NewOperableTrigger()
	manager := NewBasicManager("testing-manager", mockRepo, zap.New().WithName("testing-manager"), mockNameGenerator, compositeResolver, repoTrigger, controllerTrigger)

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	t.Log("triggering workflow created")
	err = controllerTrigger.Notify(trigger.NewEvent(namespace, workflowName, "", trigger.WorkflowCreated))

	t.Log("handle handle WorkflowCreated")
	event, err := manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.WorkflowCreated, event.GetEventType())
	gomock.InOrder(
		// resolve workflow created
		mockRepo.EXPECT().UpdateWorkflowPhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(workflow.Running)).Times(1),
		mockRepo.EXPECT().CreateNodes(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(""), gomock.Eq(entryNodeName), gomock.Eq(entryName)).Return(nil).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	t.Log("handle entry node NodeCreated")
	event, err = manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.NodeCreated, event.GetEventType())
	assert.Equal(t, entryNodeName, event.GetNodeName())
	gomock.InOrder(
		mockWorkflowStatus.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryNode, nil).Times(1),
		mockWorkflowStatus.EXPECT().GetNodesTree().Return(mockEntryTreeNode).Times(1),
		mockEntryTreeNode.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryTreeNode).Times(1),
		mockRepo.EXPECT().UpdateNodePhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(entryNodeName), gomock.Eq(node.WaitingForSchedule)).Return(nil).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	t.Log("handle entry node waiting for schedule")
	event, err = manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.NodePickChildToSchedule, event.GetEventType())
	assert.Equal(t, entryNodeName, event.GetNodeName())
	gomock.InOrder(
		// resolve entry node created
		mockWorkflowStatus.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryNodeWaitingForSchedule, nil).Times(1),
		mockWorkflowStatus.EXPECT().GetNodesTree().Return(mockEntryTreeNode).Times(1),
		mockEntryTreeNode.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryTreeNode).Times(1),
		mockEntryTreeNode.EXPECT().GetChildren().Return(mockEntryTreeNodeChildren).Times(1),
		mockEntryTreeNodeChildren.EXPECT().Length().Return(0).Times(1),
		mockEntryTreeNode.EXPECT().GetChildren().Return(mockEntryTreeNodeChildren).Times(1),
		mockEntryTreeNodeChildren.EXPECT().ContainsTemplate(gomock.Eq(layer1Template0Name)).Return(false).Times(1),
		mockRepo.EXPECT().UpdateNodePhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(entryNodeName), gomock.Eq(node.WaitingForChild)).Return(nil).Times(1),
		mockRepo.EXPECT().CreateNodes(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(entryNodeName), gomock.Eq(layer1Node0Name), gomock.Eq(layer1Template0Name)).Return(nil).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	t.Log("handle layer 1 template 0 NodeCreated")
	event, err = manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.NodeCreated, event.GetEventType())
	assert.Equal(t, layer1Node0Name, event.GetNodeName())
	gomock.InOrder(
		// resolve layer1 node 0 created
		mockWorkflowStatus.EXPECT().FetchNodeByName(layer1Node0Name).Return(layer1Node0Node, nil).Times(1),
		layer1Node0Node.EXPECT().GetTemplateName().Return(layer1Template0Name).Times(1),
		layer1Node0Node.EXPECT().GetNodePhase().Return(node.Init).Times(1),
		layer1Node0Node.EXPECT().GetTemplateName().Return(layer1Template0Name).Times(1),
		mockLayer1Template0.EXPECT().GetDuration().Return(170*time.Millisecond, nil).Times(1),
		layer1Node0Node.EXPECT().GetName().Return(layer1Node0Name).Times(1),
		layer1Node0Node.EXPECT().GetNodePhase().Return(node.Init).Times(1),
		layer1Node0Node.EXPECT().GetName().Return(layer1Node0Name).Times(1),
		mockRepo.EXPECT().UpdateNodePhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(layer1Node0Name), gomock.Eq(node.Holding)).Return(nil).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	t.Log("handle layer 1 template 0 NodeHoldingAwake")
	start := time.Now()
	event, err = manager.acquire(ctx)
	delay := time.Now().Sub(start)

	t.Logf("NodeHoldingAwake expected: 170ms actual %dms", delay.Milliseconds())
	assert.NoError(t, err)
	assert.Equal(t, trigger.NodeHoldingAwake, event.GetEventType())
	assert.Equal(t, layer1Node0Name, event.GetNodeName())
	gomock.InOrder(
		mockWorkflowStatus.EXPECT().FetchNodeByName(gomock.Eq(layer1Node0Name)).Return(layer1Node0Node, nil).Times(1),
		layer1Node0Node.EXPECT().GetTemplateName().Return(layer1Template0Name).Times(1),
		layer1Node0Node.EXPECT().GetName().Return(layer1Node0Name).Times(1),
		layer1Node0Node.EXPECT().GetNodePhase().Return(node.Init).Times(1),
		layer1Node0Node.EXPECT().GetParentNodeName().Return(entryNodeName).Times(1),
		mockRepo.EXPECT().UpdateNodePhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(layer1Node0Name), gomock.Eq(node.Succeed)).Return(nil).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	t.Log("handle entry node ChildNodeSucceed")
	event, err = manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.ChildNodeSucceed, event.GetEventType())
	assert.Equal(t, entryNodeName, event.GetNodeName())
	gomock.InOrder(
		mockWorkflowStatus.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryNode, nil).Times(1),
		mockWorkflowStatus.EXPECT().GetNodesTree().Return(mockEntryTreeNode).Times(1),
		mockEntryTreeNode.EXPECT().FetchNodeByName(entryNodeName).Return(mockEntryTreeNode).Times(1),
		mockEntryTreeNode.EXPECT().GetChildren().Return(mockEntryTreeNodeChildren).Times(1),
		mockEntryTreeNodeChildren.EXPECT().Length().Return(1).Times(1),
		mockEntryTreeNode.EXPECT().GetTemplateName().Return(entryName).Times(1),
		mockEntryTreeNode.EXPECT().GetName().Return(entryNodeName).Times(1),
		mockEntryNode.EXPECT().GetNodePhase().Return(node.WaitingForChild).Times(1),
		mockEntryNode.EXPECT().GetParentNodeName().Return("").Times(1),
		mockRepo.EXPECT().UpdateNodePhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(entryNodeName), gomock.Eq(node.Succeed)).Return(nil).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	event, err = manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.ChildNodeSucceed, event.GetEventType())
	assert.Equal(t, "", event.GetNodeName())
	gomock.InOrder()
	err = manager.consume(ctx, event)
	assert.NoError(t, err)

	event, err = manager.acquire(ctx)
	assert.NoError(t, err)
	assert.Equal(t, trigger.WorkflowFinished, event.GetEventType())
	gomock.InOrder()
	gomock.InOrder(
		mockRepo.EXPECT().UpdateWorkflowPhase(gomock.Eq(namespace), gomock.Eq(workflowName), gomock.Eq(workflow.Succeed)).Times(1),
	)
	err = manager.consume(ctx, event)
	assert.NoError(t, err)
}
