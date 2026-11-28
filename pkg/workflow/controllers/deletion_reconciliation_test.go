// Copyright Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

const deletionTestNamespace = "default"

type createTrackingClient struct {
	client.Client
	created []client.Object
}

func (c *createTrackingClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	c.created = append(c.created, obj.DeepCopyObject().(client.Object))
	return c.Client.Create(ctx, obj, opts...)
}

func TestTerminatingOwnersDoNotCreateDependents(t *testing.T) {
	workflow := func(name string, templates ...v1alpha1.Template) *v1alpha1.Workflow {
		return &v1alpha1.Workflow{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: deletionTestNamespace},
			Spec:       v1alpha1.WorkflowSpec{Templates: templates},
		}
	}
	terminatingNode := func(name string, spec v1alpha1.WorkflowNodeSpec) *v1alpha1.WorkflowNode {
		return &v1alpha1.WorkflowNode{ObjectMeta: terminatingObjectMeta(name), Spec: spec}
	}

	tests := []struct {
		name          string
		objects       []client.Object
		requestName   string
		newReconciler func(client.Client) reconcile.Reconciler
	}{
		{
			name: "workflow entry node",
			objects: []client.Object{
				&v1alpha1.Workflow{
					ObjectMeta: terminatingObjectMeta("workflow"),
					Spec: v1alpha1.WorkflowSpec{
						Entry: "entry",
						Templates: []v1alpha1.Template{{
							Name: "entry",
							Type: v1alpha1.TypeSerial,
						}},
					},
				},
			},
			requestName: "workflow",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewWorkflowEntryReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "serial child",
			objects: []client.Object{
				workflow("workflow", v1alpha1.Template{Name: "child", Type: v1alpha1.TypeSerial}),
				terminatingNode("serial", v1alpha1.WorkflowNodeSpec{
					Type:         v1alpha1.TypeSerial,
					WorkflowName: "workflow",
					Children:     []string{"child"},
				}),
			},
			requestName: "serial",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewSerialNodeReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "parallel child",
			objects: []client.Object{
				workflow("workflow", v1alpha1.Template{Name: "child", Type: v1alpha1.TypeParallel}),
				terminatingNode("parallel", v1alpha1.WorkflowNodeSpec{
					Type:         v1alpha1.TypeParallel,
					WorkflowName: "workflow",
					Children:     []string{"child"},
				}),
			},
			requestName: "parallel",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewParallelNodeReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "chaos custom resource",
			objects: []client.Object{
				terminatingNode("chaos", v1alpha1.WorkflowNodeSpec{
					Type: v1alpha1.TypePodChaos,
					EmbedChaos: &v1alpha1.EmbedChaos{
						PodChaos: &v1alpha1.PodChaosSpec{},
					},
				}),
			},
			requestName: "chaos",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewChaosNodeReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "schedule custom resource",
			objects: []client.Object{
				terminatingNode("schedule", v1alpha1.WorkflowNodeSpec{
					Type:     v1alpha1.TypeSchedule,
					Schedule: &v1alpha1.ScheduleSpec{},
				}),
			},
			requestName: "schedule",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewChaosNodeReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "task pod",
			objects: []client.Object{
				workflow("workflow"),
				terminatingNode("task", v1alpha1.WorkflowNodeSpec{
					Type: v1alpha1.TypeTask,
					Task: &v1alpha1.Task{Container: &corev1.Container{
						Name:  "task",
						Image: "task:latest",
					}},
				}),
			},
			requestName: "task",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewTaskReconciler(c, nil, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "task child",
			objects: []client.Object{
				workflow("workflow", v1alpha1.Template{Name: "child", Type: v1alpha1.TypeSerial}),
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "task-pod",
						Namespace: deletionTestNamespace,
						Labels:    map[string]string{v1alpha1.LabelControlledBy: "task"},
					},
					Status: corev1.PodStatus{Phase: corev1.PodSucceeded},
				},
				&v1alpha1.WorkflowNode{
					ObjectMeta: terminatingObjectMeta("task"),
					Spec: v1alpha1.WorkflowNodeSpec{
						Type:         v1alpha1.TypeTask,
						WorkflowName: "workflow",
						ConditionalBranches: []v1alpha1.ConditionalBranch{{
							Target: "child",
						}},
					},
					Status: v1alpha1.WorkflowNodeStatus{
						ConditionalBranchesStatus: &v1alpha1.ConditionalBranchesStatus{
							Branches: []v1alpha1.ConditionalBranchStatus{{
								Target:           "child",
								EvaluationResult: corev1.ConditionTrue,
							}},
						},
					},
				},
			},
			requestName: "task",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewTaskReconciler(c, nil, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
		{
			name: "status check",
			objects: []client.Object{
				workflow("workflow"),
				terminatingNode("status-check", v1alpha1.WorkflowNodeSpec{
					Type: v1alpha1.TypeStatusCheck,
					StatusCheck: &v1alpha1.StatusCheckSpec{
						Type: v1alpha1.TypeHTTP,
					},
				}),
			},
			requestName: "status-check",
			newReconciler: func(c client.Client) reconcile.Reconciler {
				return NewStatusCheckReconciler(c, recorder.NewDebugRecorder(), logr.Discard())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := newCreateTrackingClient(t, tt.objects...)
			request := reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      tt.requestName,
				Namespace: deletionTestNamespace,
			}}

			if _, err := tt.newReconciler(kubeClient).Reconcile(context.Background(), request); err != nil {
				t.Fatalf("reconcile terminating owner: %v", err)
			}
			if len(kubeClient.created) != 0 {
				t.Fatalf("created %d dependent(s) for terminating owner", len(kubeClient.created))
			}

			// The same fixture must exercise creation when its owner is not deleting.
			// This ensures the deletion assertion does not pass for an unrelated reason.
			activeObjects := make([]client.Object, len(tt.objects))
			for i, obj := range tt.objects {
				activeObjects[i] = obj.DeepCopyObject().(client.Object)
				if activeObjects[i].GetName() == tt.requestName {
					activeObjects[i].SetDeletionTimestamp(nil)
				}
			}
			activeClient := newCreateTrackingClient(t, activeObjects...)
			if _, err := tt.newReconciler(activeClient).Reconcile(context.Background(), request); err != nil {
				t.Fatalf("reconcile active owner: %v", err)
			}
			if len(activeClient.created) != 1 {
				t.Fatalf("created %d dependent(s) for active owner, want 1", len(activeClient.created))
			}
		})
	}
}

func terminatingObjectMeta(name string) metav1.ObjectMeta {
	now := metav1.Now()
	return metav1.ObjectMeta{
		Name:              name,
		Namespace:         deletionTestNamespace,
		DeletionTimestamp: &now,
		Finalizers:        []string{"test.chaos-mesh.org/deletion"},
		Labels:            map[string]string{v1alpha1.LabelWorkflow: "workflow"},
	}
}

func newCreateTrackingClient(t *testing.T, objects ...client.Object) *createTrackingClient {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core API to scheme: %v", err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add Chaos Mesh API to scheme: %v", err)
	}

	return &createTrackingClient{Client: fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.Workflow{}, &v1alpha1.WorkflowNode{}).
		WithObjects(objects...).
		Build()}
}
