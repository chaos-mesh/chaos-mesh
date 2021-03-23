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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func Test_conversionWorkflow(t *testing.T) {
	type args struct {
		kubeWorkflow v1alpha1.Workflow
	}
	tests := []struct {
		name string
		args args
		want Workflow
	}{
		{
			name: "simple workflow",
			args: args{
				v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "fake-workflow-0",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry: "an-entry",
					},
					Status: v1alpha1.WorkflowStatus{},
				},
			},
			want: Workflow{
				Name:   "fake-workflow-0",
				Entry:  "an-entry",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conversionWorkflow(tt.args.kubeWorkflow); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("conversionWorkflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_conversionWorkflowDetail(t *testing.T) {
	type args struct {
		kubeWorkflow v1alpha1.Workflow
		nodes        []v1alpha1.WorkflowNode
	}
	tests := []struct {
		name string
		args args
		want WorkflowDetail
	}{
		{
			name: "simple workflow detail with no nodes",
			args: args{
				kubeWorkflow: v1alpha1.Workflow{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "another-fake-workflow",
					},
					Spec: v1alpha1.WorkflowSpec{
						Entry:     "another-entry",
						Templates: nil,
					},
					Status: v1alpha1.WorkflowStatus{},
				},
				nodes: nil,
			},
			want: WorkflowDetail{
				Workflow: Workflow{
					Name:   "another-fake-workflow",
					Entry:  "another-entry",
				},
				Topology: Topology{
					Nodes: []Node{},
				},
				CurrentNodes: []Node{},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conversionWorkflowDetail(tt.args.kubeWorkflow, tt.args.nodes...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("conversionWorkflowDetail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_conversionWorkflowNode(t *testing.T) {
	type args struct {
		kubeWorkflowNode v1alpha1.WorkflowNode
	}
	tests := []struct {
		name string
		args args
		want Node
	}{
		{
			name: "simple node",
			args: args{kubeWorkflowNode: v1alpha1.WorkflowNode{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "fake-node-0",
				},
				Spec: v1alpha1.WorkflowNodeSpec{
					WorkflowName: "fake-workflow-0",
					TemplateName: "fake-template-0",
					Type:         v1alpha1.TypeJvmChaos,
				},
				Status: v1alpha1.WorkflowNodeStatus{},
			}},
			want: Node{
				Name:     "fake-node-0",
				Type:     ChaosNode,
				Serial:   NodeSerial{[]string{}},
				Parallel: NodeParallel{[]string{}},
				Template: "fake-template-0",
			},
		}, {
			name: "serial node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "fake-serial-node-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "fake-serial-node",
						WorkflowName: "fake-workflow-0",
						Type:         v1alpha1.TypeSerial,
						Tasks:        []string{"child-0", "child-1"},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name: "fake-serial-node-0",
				Type: SerialNode,
				Serial: NodeSerial{
					Tasks: []string{"child-0", "child-1"},
				},
				Parallel: NodeParallel{[]string{}},
				Template: "fake-serial-node",
			},
		},
		{
			name: "parallel node",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "parallel-node-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "parallel-node",
						WorkflowName: "another-fake-workflow",
						Type:         v1alpha1.TypeParallel,
						Tasks:        []string{"child-1", "child-0"},
					},
					Status: v1alpha1.WorkflowNodeStatus{},
				},
			},
			want: Node{
				Name:   "parallel-node-0",
				Type:   ParallelNode,
				Serial: NodeSerial{[]string{}},
				Parallel: NodeParallel{
					Tasks: []string{"child-1", "child-0"},
				},
				Template: "parallel-node",
			},
		},
		{
			name: "some chaos",
			args: args{
				kubeWorkflowNode: v1alpha1.WorkflowNode{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name: "io-chaos-0",
					},
					Spec: v1alpha1.WorkflowNodeSpec{
						TemplateName: "io-chaos",
						WorkflowName: "another-workflow-0",
						Type:         v1alpha1.TypeIoChaos,
						EmbedChaos: &v1alpha1.EmbedChaos{
							IoChaos: &v1alpha1.IoChaosSpec{
								Mode:       v1alpha1.OnePodMode,
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
				Serial:   NodeSerial{[]string{}},
				Parallel: NodeParallel{[]string{}},
				Template: "io-chaos",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := conversionWorkflowNode(tt.args.kubeWorkflowNode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("conversionWorkflowNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
