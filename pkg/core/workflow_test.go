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

package core

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func Test_convertWorkflow(t *testing.T) {
	type args struct {
		kubeWorkflow v1alpha1.Workflow
	}
	tests := []struct {
		name string
		args args
		want WorkflowMeta
	}{
		{
			name: "simple workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{},
				},
			},
			want: WorkflowMeta{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
				Status:    WorkflowUnknown,
			},
		}, {
			name: "running workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{
						Conditions: []v1alpha1.WorkflowCondition{
							{
								Type:   v1alpha1.WorkflowConditionScheduled,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
						},
					},
				},
			},
			want: WorkflowMeta{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
				Status:    WorkflowRunning,
			},
		}, {
			name: "running workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{
						Conditions: []v1alpha1.WorkflowCondition{
							{
								Type:   v1alpha1.WorkflowConditionAccomplished,
								Status: corev1.ConditionUnknown,
								Reason: "",
							},
							{
								Type:   v1alpha1.WorkflowConditionScheduled,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
						},
					},
				},
			},
			want: WorkflowMeta{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
				Status:    WorkflowRunning,
			},
		}, {
			name: "running workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{
						Conditions: []v1alpha1.WorkflowCondition{
							{
								Type:   v1alpha1.WorkflowConditionAccomplished,
								Status: corev1.ConditionFalse,
								Reason: "",
							},
							{
								Type:   v1alpha1.WorkflowConditionScheduled,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
						},
					},
				},
			},
			want: WorkflowMeta{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
				Status:    WorkflowRunning,
			},
		}, {
			name: "succeed workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{
						Conditions: []v1alpha1.WorkflowCondition{
							{
								Type:   v1alpha1.WorkflowConditionAccomplished,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
							{
								Type:   v1alpha1.WorkflowConditionScheduled,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
						},
					},
				},
			},
			want: WorkflowMeta{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
				Status:    WorkflowSucceed,
			},
		}, {
			name: "converting UID",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-workflow-0",
						UID:       "uid-of-workflow",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{
						Conditions: []v1alpha1.WorkflowCondition{
							{
								Type:   v1alpha1.WorkflowConditionAccomplished,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
							{
								Type:   v1alpha1.WorkflowConditionScheduled,
								Status: corev1.ConditionTrue,
								Reason: "",
							},
						},
					},
				},
			},
			want: WorkflowMeta{
				Namespace: "fake-namespace",
				Name:      "fake-workflow-0",
				Entry:     "an-entry",
				Status:    WorkflowSucceed,
				UID:       "uid-of-workflow",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertWorkflow(tt.args.kubeWorkflow); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertWorkflowDetail(t *testing.T) {
	type args struct {
		kubeWorkflow v1alpha1.Workflow
		kubeNodes    []v1alpha1.WorkflowNode
	}
	tests := []struct {
		name    string
		args    args
		want    WorkflowDetail
		wantErr bool
	}{
		{
			name: "simple workflow detail with no nodes",
			args: args{
				kubeWorkflow: v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "another-namespace",
						Name:      "another-fake-workflow",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry:     "another-entry",
						Templates: nil,
					},
					Status: v1alpha1.WorkflowStatus{},
				},
				kubeNodes: nil,
			},
			want: WorkflowDetail{
				WorkflowMeta: WorkflowMeta{
					Namespace: "another-namespace",
					Name:      "another-fake-workflow",
					Entry:     "another-entry",
					Status:    WorkflowUnknown,
				},
				Topology: Topology{
					Nodes: []Node{},
				},
				KubeObject: KubeObjectDesc{
					Meta: KubeObjectMeta{
						Name:      "another-fake-workflow",
						Namespace: "another-namespace",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "another-entry",
					},
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertWorkflowDetail(tt.args.kubeWorkflow, tt.args.kubeNodes)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertWorkflowDetail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkflowDetail() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertWorkflowNode(t *testing.T) {
	type args struct {
		kubeWorkflowNode v1alpha1.WorkflowNode
	}
	tests := []struct {
		name    string
		args    args
		want    Node
		wantErr bool
	}{
		{
			name: "simple node",
			args: args{kubeWorkflowNode: v1alpha1.WorkflowNode{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fake-namespace",
					Name:      "fake-node-0",
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					WorkflowName: "fake-workflow-0",
					TemplateName: "fake-template-0",
					Type:         v1alpha1.TypeJVMChaos,
				},
				Status: v1alpha1.WorkflowNodeStatus{},
			}},
			want: Node{
				Name:     "fake-node-0",
				Type:     ChaosNode,
				Serial:   nil,
				Parallel: nil,
				Template: "fake-template-0",
				State:    NodeRunning,
			},
		}, {
			name: "serial node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "fake-serial-node-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "fake-serial-node",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeSerial,
						Children:     []string{"child-0", "child-1"},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name: "fake-serial-node-0",
				Type: SerialNode,
				Serial: &NodeSerial{
					Children: []NodeNameWithTemplate{
						{Name: "", Template: "child-0"},
						{Name: "", Template: "child-1"},
					},
				},
				Parallel: nil,
				Template: "fake-serial-node",
				State:    NodeRunning,
			},
		},
		{
			name: "parallel node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "parallel-node-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "parallel-node",
						WorkflowName: "another-fake-workflow",
						Type:         v1alpha1.TypeParallel,
						Children:     []string{"child-1", "child-0"},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name:   "parallel-node-0",
				Type:   ParallelNode,
				Serial: nil,
				Parallel: &NodeParallel{
					Children: []NodeNameWithTemplate{
						{Name: "", Template: "child-1"},
						{Name: "", Template: "child-0"},
					},
				},
				Template: "parallel-node",
				State:    NodeRunning,
			},
		},
		{
			name: "some chaos",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "io-chaos-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "io-chaos",
						WorkflowName: "another-workflow-0",
						Type:         v1alpha1.TypeIOChaos,
						EmbedChaos: &v1alpha1.EmbedChaos{
							IOChaos: &v1alpha1.IOChaosSpec{
								ContainerSelector: v1alpha1.ContainerSelector{
									PodSelector: v1alpha1.PodSelector{
										Mode: v1alpha1.OnePodMode,
									},
								},
								Action:     "delay",
								Delay:      "100ms",
								Path:       "/fake/path",
								Percent:    100,
								VolumePath: "/fake/path",
							},
						},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name:     "io-chaos-0",
				Type:     ChaosNode,
				Serial:   nil,
				Parallel: nil,
				Template: "io-chaos",
				State:    NodeRunning,
			},
		},
		{
			name: "accomplished node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "the-entry-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "the-entry",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeSerial,
						Children:     []string{"unimportant-task-0"},
					},
					Status: v1alpha1.WorkflowNodeStatus{
						Conditions: []v1alpha1.WorkflowNodeCondition{
							{
								Type:   v1alpha1.ConditionAccomplished,
								Status: corev1.ConditionTrue,
								Reason: "unit test mocked true",
							},
						},
					},
				},
			},
			want: Node{
				Name:  "the-entry-0",
				Type:  SerialNode,
				State: NodeSucceed,
				Serial: &NodeSerial{
					Children: []NodeNameWithTemplate{
						{Name: "", Template: "unimportant-task-0"},
					},
				},
				Parallel: nil,
				Template: "the-entry",
			},
		},
		{
			name: "deadline exceed node",
			args: args{kubeWorkflowNode: v1alpha1.WorkflowNode{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "fake-namespace",
					Name:      "deadline-exceed-node-0",
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					TemplateName: "deadline-exceed-node",
					WorkflowName: "some-workflow",
					Type:         v1alpha1.TypePodChaos,
				},
				Status: v1alpha1.WorkflowNodeStatus{
					Conditions: []v1alpha1.WorkflowNodeCondition{
						{
							Type:   v1alpha1.ConditionDeadlineExceed,
							Status: corev1.ConditionTrue,
							Reason: "unit test mocked true",
						},
					},
				},
			}},
			want: Node{
				Name:     "deadline-exceed-node-0",
				Type:     ChaosNode,
				State:    NodeSucceed,
				Serial:   nil,
				Parallel: nil,
				Template: "deadline-exceed-node",
			},
		},
		{
			name: "appending uid",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "fake-namespace",
						Name:      "the-entry-0",
						UID:       "uid-of-workflow-node",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "the-entry",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeSerial,
						Children:     []string{"unimportant-task-0"},
					},
					Status: v1alpha1.WorkflowNodeStatus{
						Conditions: []v1alpha1.WorkflowNodeCondition{
							{
								Type:   v1alpha1.ConditionAccomplished,
								Status: corev1.ConditionTrue,
								Reason: "unit test mocked true",
							},
						},
					},
				},
			},
			want: Node{
				Name:  "the-entry-0",
				Type:  SerialNode,
				State: NodeSucceed,
				Serial: &NodeSerial{
					Children: []NodeNameWithTemplate{
						{Name: "", Template: "unimportant-task-0"},
					},
				},
				Parallel: nil,
				Template: "the-entry",
				UID:      "uid-of-workflow-node",
			},
		},
		{
			name: "task node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mocking-task-node-0",
						Namespace: "mocked-namespace",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "mocking-task-node",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeTask,
						ConditionalBranches: []v1alpha1.ConditionalBranch{
							{
								Target:     "one-node",
								Expression: "exitCode == 0",
							},
							{
								Target:     "another-node",
								Expression: "exitCode != 0",
							},
						},
					},
					Status: v1alpha1.WorkflowNodeStatus{
						ConditionalBranchesStatus: &v1alpha1.ConditionalBranchesStatus{
							Branches: []v1alpha1.ConditionalBranchStatus{
								{
									Target:           "one-node",
									EvaluationResult: corev1.ConditionFalse,
								},
								{
									Target:           "another-node",
									EvaluationResult: corev1.ConditionTrue,
								},
							},
							Context: nil,
						},
						ActiveChildren: []corev1.LocalObjectReference{
							{
								Name: "another-node-0",
							},
						},
					},
				},
			},
			want: Node{
				Name:  "mocking-task-node-0",
				Type:  TaskNode,
				State: NodeRunning,
				ConditionalBranches: []ConditionalBranch{
					{
						NodeNameWithTemplate: NodeNameWithTemplate{
							Template: "one-node",
							Name:     "",
						},
						Expression: "exitCode == 0",
					},
					{
						NodeNameWithTemplate: NodeNameWithTemplate{
							Template: "another-node",
							Name:     "another-node-0",
						},
						Expression: "exitCode != 0",
					},
				},
				Template: "mocking-task-node",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertWorkflowNode(tt.args.kubeWorkflowNode)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertWorkflowNode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("convertWorkflowNode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_composeTaskAndNodes(t *testing.T) {
	type args struct {
		children []string
		nodes    []string
	}
	tests := []struct {
		name string
		args args
		want []NodeNameWithTemplate
	}{
		{
			name: "ordered with serial",
			args: args{
				children: []string{"node-0", "node-1", "node-0", "node-2", "node-3"},
				nodes:    []string{"node-0-instance", "node-1-instance", "node-0-another_instance"},
			},
			want: []NodeNameWithTemplate{
				{
					Name:     "node-0-instance",
					Template: "node-0",
				}, {
					Name:     "node-1-instance",
					Template: "node-1",
				}, {
					Name:     "node-0-another_instance",
					Template: "node-0",
				}, {
					Name:     "",
					Template: "node-2",
				}, {
					Name:     "",
					Template: "node-3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := composeSerialTaskAndNodes(tt.args.children, tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("composeSerialTaskAndNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_composeParallelTaskAndNodes(t *testing.T) {
	type args struct {
		children []string
		nodes    []string
	}
	tests := []struct {
		name string
		args args
		want []NodeNameWithTemplate
	}{
		{
			name: "parallel",
			args: args{
				children: []string{"node-a", "node-b", "node-a", "node-c", "node-d"},
				nodes:    []string{"node-a-instance", "node-a-another_instance", "node-d-instance"},
			},
			want: []NodeNameWithTemplate{
				{
					Name:     "node-a-instance",
					Template: "node-a",
				}, {
					Name:     "",
					Template: "node-b",
				}, {
					Name:     "node-a-another_instance",
					Template: "node-a",
				}, {
					Name:     "",
					Template: "node-c",
				}, {
					Name:     "node-d-instance",
					Template: "node-d",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := composeParallelTaskAndNodes(tt.args.children, tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("composeParallelTaskAndNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_composeTaskConditionalBranches(t *testing.T) {
	type args struct {
		conditionalBranches []v1alpha1.ConditionalBranch
		nodes               []string
	}
	tests := []struct {
		name string
		args args
		want []ConditionalBranch
	}{
		{
			name: "task node all of the branch is selected",
			args: args{
				conditionalBranches: []v1alpha1.ConditionalBranch{
					{
						Target:     "template-a",
						Expression: "a: whatever valid or not",
					},
					{
						Target:     "template-b",
						Expression: "b: whatever valid or not",
					},
					{
						Target:     "template-c",
						Expression: "c: whatever valid or not",
					},
				},
				nodes: []string{
					"template-a-0",
					"template-b-0",
					"template-c-0",
				},
			},
			want: []ConditionalBranch{
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "template-a-0",
						Template: "template-a",
					},
					Expression: "a: whatever valid or not",
				},
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "template-b-0",
						Template: "template-b",
					},
					Expression: "b: whatever valid or not",
				},
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "template-c-0",
						Template: "template-c",
					},
					Expression: "c: whatever valid or not",
				},
			},
		},
		{
			name: "none of the branch is selected",
			args: args{
				conditionalBranches: []v1alpha1.ConditionalBranch{
					{
						Target:     "template-a",
						Expression: "a: whatever valid or not",
					},
					{
						Target:     "template-b",
						Expression: "b: whatever valid or not",
					},
					{
						Target:     "template-c",
						Expression: "c: whatever valid or not",
					},
				},
				nodes: []string{},
			},
			want: []ConditionalBranch{
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "",
						Template: "template-a",
					},
					Expression: "a: whatever valid or not",
				},
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "",
						Template: "template-b",
					},
					Expression: "b: whatever valid or not",
				},
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "",
						Template: "template-c",
					},
					Expression: "c: whatever valid or not",
				},
			},
		},
		{
			name: "part of the branch is selected",
			args: args{
				conditionalBranches: []v1alpha1.ConditionalBranch{
					{
						Target:     "template-a",
						Expression: "a: whatever valid or not",
					},
					{
						Target:     "template-b",
						Expression: "b: whatever valid or not",
					},
					{
						Target:     "template-c",
						Expression: "c: whatever valid or not",
					},
				},
				nodes: []string{
					"template-a-0",
				},
			},
			want: []ConditionalBranch{
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "template-a-0",
						Template: "template-a",
					},
					Expression: "a: whatever valid or not",
				},
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "",
						Template: "template-b",
					},
					Expression: "b: whatever valid or not",
				},
				{
					NodeNameWithTemplate: NodeNameWithTemplate{
						Name:     "",
						Template: "template-c",
					},
					Expression: "c: whatever valid or not",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := composeTaskConditionalBranches(tt.args.conditionalBranches, tt.args.nodes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("composeTaskConditionalBranches() = %v, want %v", got, tt.want)
			}
		})
	}
}
