// Copyright 2021 Chaos Mesh Authors.
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
//

package controllers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestTaskContainerNameForCollection(t *testing.T) {
	node := v1alpha1.WorkflowNode{
		Spec: v1alpha1.WorkflowNodeSpec{
			Task: &v1alpha1.Task{
				Container: &corev1.Container{
					Name: "main",
				},
			},
		},
	}
	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "istio-proxy"},
				{Name: "main"},
			},
		},
	}

	if got := taskContainerNameForCollection(node, pod); got != "main" {
		t.Fatalf("taskContainerNameForCollection() = %q, want %q", got, "main")
	}
}

func TestPersistEvaluatedResultDoesNotStoreContextInAnnotations(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("add client-go scheme: %v", err)
	}
	if err := v1alpha1.SchemeBuilder.AddToScheme(scheme); err != nil {
		t.Fatalf("add workflow scheme: %v", err)
	}

	node := &v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "main-entry",
			Namespace: "default",
		},
		Spec: v1alpha1.WorkflowNodeSpec{
			Type: v1alpha1.TypeEphemeralTask,
			ConditionalBranches: []v1alpha1.ConditionalBranch{
				{
					Target:     "success",
					Expression: `exitCode == 0 && stdout == "ok"`,
				},
				{
					Target:     "failure",
					Expression: `exitCode != 0`,
				},
			},
		},
	}

	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1alpha1.WorkflowNode{}).
		WithObjects(node).
		Build()

	reconciler := &EphemeralTaskReconciler{
		TaskReconciler: &TaskReconciler{
			kubeClient:    kubeClient,
			eventRecorder: recorder.NewDebugRecorder(),
			logger:        logr.Discard(),
		},
	}

	largeStdout := strings.Repeat("x", 300*1024)
	if err := reconciler.persistEvaluatedResult(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "main-entry",
	}, map[string]interface{}{
		"exitCode": 0,
		"stdout":   largeStdout,
	}); err != nil {
		t.Fatalf("persistEvaluatedResult() error = %v", err)
	}

	updated := &v1alpha1.WorkflowNode{}
	if err := kubeClient.Get(context.Background(), types.NamespacedName{
		Namespace: "default",
		Name:      "main-entry",
	}, updated); err != nil {
		t.Fatalf("get updated node: %v", err)
	}

	if updated.Annotations[ephemeralTaskResultCollectedAnnotation] != "true" {
		t.Fatalf("result-collected annotation = %q, want true", updated.Annotations[ephemeralTaskResultCollectedAnnotation])
	}
	if _, exists := updated.Annotations["workflow.chaos-mesh.org/ephemeral-task-context"]; exists {
		t.Fatalf("unexpected context annotation present")
	}
	if updated.Status.ConditionalBranchesStatus == nil || len(updated.Status.ConditionalBranchesStatus.Context) != 1 {
		t.Fatalf("status context was not persisted")
	}
	if len(updated.Status.ConditionalBranchesStatus.Context[0]) == 0 {
		t.Fatalf("status context is empty")
	}

	contextValue := map[string]interface{}{}
	if err := json.Unmarshal([]byte(updated.Status.ConditionalBranchesStatus.Context[0]), &contextValue); err != nil {
		t.Fatalf("unmarshal context: %v", err)
	}

	stdout, ok := contextValue["stdout"].(string)
	if !ok {
		t.Fatalf("persisted stdout is missing or not a string: %#v", contextValue["stdout"])
	}
	if len(stdout) != maxPersistedStdoutBytes {
		t.Fatalf("persisted stdout length = %d, want %d", len(stdout), maxPersistedStdoutBytes)
	}
	if contextValue["stdoutTruncated"] != true {
		t.Fatalf("stdoutTruncated = %#v, want true", contextValue["stdoutTruncated"])
	}
	if got, ok := contextValue["stdoutOriginalBytes"].(float64); !ok || int(got) != len(largeStdout) {
		t.Fatalf("stdoutOriginalBytes = %#v, want %d", contextValue["stdoutOriginalBytes"], len(largeStdout))
	}
}

func TestTaskSyncChildNodesNoopWhenWorkflowNodeFinished(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("add client-go scheme: %v", err)
	}
	if err := v1alpha1.SchemeBuilder.AddToScheme(scheme); err != nil {
		t.Fatalf("add workflow scheme: %v", err)
	}

	kubeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	reconciler := &TaskReconciler{
		ChildNodesFetcher: NewChildNodesFetcher(kubeClient, logr.Discard()),
		kubeClient:        kubeClient,
		eventRecorder:     recorder.NewDebugRecorder(),
		logger:            logr.Discard(),
	}

	node := v1alpha1.WorkflowNode{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "finished-ephemeral",
			Namespace: "default",
		},
		Status: v1alpha1.WorkflowNodeStatus{
			Conditions: []v1alpha1.WorkflowNodeCondition{
				{
					Type:   v1alpha1.ConditionDeadlineExceed,
					Status: corev1.ConditionTrue,
					Reason: v1alpha1.NodeDeadlineExceed,
				},
			},
			ConditionalBranchesStatus: &v1alpha1.ConditionalBranchesStatus{
				Branches: []v1alpha1.ConditionalBranchStatus{
					{
						Target:           "child",
						EvaluationResult: corev1.ConditionTrue,
					},
				},
			},
		},
	}

	if err := reconciler.syncChildNodes(context.Background(), node); err != nil {
		t.Fatalf("syncChildNodes() error = %v", err)
	}
}
