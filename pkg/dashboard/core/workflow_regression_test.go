// Copyright 2026 Chaos Mesh Authors.
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

package core

import (
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestWorkflowTopologyDistinguishesTemplatePrefixes(t *testing.T) {
	for _, nodeType := range []v1alpha1.TemplateType{v1alpha1.TypeParallel, v1alpha1.TypeTask} {
		t.Run(string(nodeType), func(t *testing.T) {
			wf := v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "workflow", Namespace: "default"}}
			node := v1alpha1.WorkflowNode{
				ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "default", Labels: map[string]string{v1alpha1.LabelWorkflow: wf.Name}},
				Spec: v1alpha1.WorkflowNodeSpec{
					Type: nodeType, Children: []string{"check", "check-extra"},
					ConditionalBranches: []v1alpha1.ConditionalBranch{{Target: "check", Expression: "true"}, {Target: "check-extra", Expression: "false"}},
				},
				Status: v1alpha1.WorkflowNodeStatus{ActiveChildren: []corev1.LocalObjectReference{{Name: "check-extra-abcde"}}},
			}
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&wf, &node).Build()
			got, err := NewKubeWorkflowRepository(kube).Get(context.Background(), "default", "workflow")
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Topology.Nodes) != 1 {
				t.Fatalf("topology = %#v", got.Topology)
			}
			converted := got.Topology.Nodes[0]
			if nodeType == v1alpha1.TypeParallel {
				want := []NodeNameWithTemplate{{Template: "check"}, {Name: "check-extra-abcde", Template: "check-extra"}}
				if !reflect.DeepEqual(converted.Parallel, want) {
					t.Fatalf("parallel = %#v, want %#v", converted.Parallel, want)
				}
			} else {
				want := []ConditionalBranch{
					{NodeNameWithTemplate: NodeNameWithTemplate{Template: "check"}, Expression: "true"},
					{NodeNameWithTemplate: NodeNameWithTemplate{Name: "check-extra-abcde", Template: "check-extra"}, Expression: "false"},
				}
				if !reflect.DeepEqual(converted.ConditionalBranches, want) {
					t.Fatalf("branches = %#v, want %#v", converted.ConditionalBranches, want)
				}
			}
		})
	}
}

func TestWorkflowEntityConversions(t *testing.T) {
	wf := &v1alpha1.Workflow{
		TypeMeta:   metav1.TypeMeta{Kind: "Workflow", APIVersion: v1alpha1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "default", UID: "workflow-uid", Labels: map[string]string{"app": "test"}, Annotations: map[string]string{"note": "preserved"}},
		Spec:       v1alpha1.WorkflowSpec{Entry: "wait", Templates: []v1alpha1.Template{{Name: "wait", Type: v1alpha1.TypeSuspend}}},
	}
	entity, err := WorkflowCR2WorkflowEntity(wf)
	if err != nil {
		t.Fatal(err)
	}
	entity.ID = 42
	entity.Archived = true
	roundtrip, err := WorkflowEntity2WorkflowCR(entity)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(roundtrip, wf) {
		t.Fatalf("roundtrip = %#v, want %#v", roundtrip, wf)
	}
	detail, err := WorkflowEntity2WorkflowDetail(entity)
	if err != nil {
		t.Fatal(err)
	}
	if detail.ID != 42 || !detail.Archived || detail.UID != "workflow-uid" || !reflect.DeepEqual(detail.KubeObject.Meta.Labels, wf.Labels) || !reflect.DeepEqual(detail.KubeObject.Meta.Annotations, wf.Annotations) || !reflect.DeepEqual(detail.KubeObject.Spec, wf.Spec) {
		t.Fatalf("detail lost persisted metadata or Kubernetes object: %#v", detail)
	}
	for _, raw := range []string{"", "{", `{"spec":false}`} {
		t.Run("invalid JSON "+raw, func(t *testing.T) {
			if got, err := WorkflowEntity2WorkflowCR(&WorkflowEntity{Workflow: raw}); err == nil || got != nil {
				t.Fatalf("CR = %#v, error = %v", got, err)
			}
			if got, err := WorkflowEntity2WorkflowDetail(&WorkflowEntity{Workflow: raw}); err == nil || got != nil {
				t.Fatalf("detail = %#v, error = %v", got, err)
			}
		})
	}
}

func TestWorkflowEntityConversionsNil(t *testing.T) {
	if got, err := WorkflowCR2WorkflowEntity(nil); got != nil || err != nil {
		t.Fatalf("entity = %#v, error = %v", got, err)
	}
	if got, err := WorkflowEntity2WorkflowCR(nil); got != nil || err != nil {
		t.Fatalf("CR = %#v, error = %v", got, err)
	}
	if got, err := WorkflowEntity2WorkflowDetail(nil); got != nil || err != nil {
		t.Fatalf("detail = %#v, error = %v", got, err)
	}
}

func TestKubeWorkflowRepositoryLifecycleAndIsolation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	unrelated := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "unrelated", Namespace: "other"}}
	nodes := []v1alpha1.WorkflowNode{
		{ObjectMeta: metav1.ObjectMeta{Name: "selected", Namespace: "default", Labels: map[string]string{v1alpha1.LabelWorkflow: "sample"}}, Spec: v1alpha1.WorkflowNodeSpec{Type: v1alpha1.TypeSuspend}},
		{ObjectMeta: metav1.ObjectMeta{Name: "different-workflow", Namespace: "default", Labels: map[string]string{v1alpha1.LabelWorkflow: "unrelated"}}, Spec: v1alpha1.WorkflowNodeSpec{Type: "invalid"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "different-namespace", Namespace: "other", Labels: map[string]string{v1alpha1.LabelWorkflow: "sample"}}, Spec: v1alpha1.WorkflowNodeSpec{Type: "invalid"}},
	}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(unrelated, &nodes[0], &nodes[1], &nodes[2]).Build()
	repository := NewKubeWorkflowRepository(kube)
	wf := v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "default"}, Spec: v1alpha1.WorkflowSpec{Entry: "initial"}}
	created, err := repository.Create(ctx, wf)
	if err != nil {
		t.Fatal(err)
	}
	if created.Name != "sample" || created.Entry != "initial" || len(created.Topology.Nodes) != 1 || created.Topology.Nodes[0].Name != "selected" {
		t.Fatalf("created detail = %#v", created)
	}
	if _, err := repository.Create(ctx, wf); !apierrors.IsAlreadyExists(err) {
		t.Fatalf("duplicate creation error = %v", err)
	}
	filtered, err := repository.ListByNamespace(ctx, "default")
	if err != nil || len(filtered) != 1 || filtered[0].Name != "sample" {
		t.Fatalf("namespace list = %#v, error = %v", filtered, err)
	}
	all, err := repository.List(ctx)
	if err != nil || len(all) != 2 {
		t.Fatalf("all workflows = %#v, error = %v", all, err)
	}
	wf.Spec.Entry = "updated"
	updated, err := repository.Update(ctx, "default", "sample", wf)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Entry != "updated" {
		t.Fatalf("updated workflow = %#v", updated)
	}
	stored := &v1alpha1.Workflow{}
	if err := kube.Get(ctx, client.ObjectKeyFromObject(&wf), stored); err != nil {
		t.Fatal(err)
	}
	if stored.Spec.Entry != "updated" {
		t.Fatalf("persisted entry = %q", stored.Spec.Entry)
	}
	if err := repository.Delete(ctx, "default", "sample"); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Get(ctx, "default", "sample"); !apierrors.IsNotFound(err) {
		t.Fatalf("deleted workflow get error = %v", err)
	}
	if err := repository.Delete(ctx, "default", "sample"); !apierrors.IsNotFound(err) {
		t.Fatalf("missing workflow delete error = %v", err)
	}
	if _, err := repository.Update(ctx, "default", "sample", wf); !apierrors.IsNotFound(err) {
		t.Fatalf("missing workflow update error = %v", err)
	}
	if _, err := repository.Get(ctx, "other", "unrelated"); err != nil {
		t.Fatalf("unrelated workflow lost: %v", err)
	}
}

func TestWorkflowRepositoryRejectsUnknownNodeType(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	wf := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{Name: "sample", Namespace: "default"}}
	node := &v1alpha1.WorkflowNode{ObjectMeta: metav1.ObjectMeta{Name: "unknown", Namespace: "default", Labels: map[string]string{v1alpha1.LabelWorkflow: "sample"}}, Spec: v1alpha1.WorkflowNodeSpec{Type: "unsupported"}}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(wf, node).Build()
	if got, err := NewKubeWorkflowRepository(kube).Get(context.Background(), "default", "sample"); err == nil || len(got.Topology.Nodes) != 0 {
		t.Fatalf("invalid node detail = %#v, error = %v", got, err)
	}
}
