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

package nodescoped

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
	. "github.com/chaos-mesh/chaos-mesh/pkg/testutils"
)

func ids(targets []Target) []string {
	result := []string{}
	for _, t := range targets {
		result = append(result, t.Id())
	}
	return result
}

func newSelector(objects []runtime.Object, option generic.Option, excluded ...string) *SelectImpl {
	c := fake.NewClientBuilder().
		WithRuntimeObjects(objects...).
		WithIndex(&v1.Pod{}, NodeNameField, func(obj client.Object) []string {
			return []string{obj.(*v1.Pod).Spec.NodeName}
		}).
		Build()
	pods := pod.New(pod.Params{Client: c, Reader: c})
	return NewForTest(c, c, pods, option, excluded...)
}

func TestSelectFansOutToTheVictimsNode(t *testing.T) {
	g := NewWithT(t)

	var objects []runtime.Object
	add := func(arg PodArg) {
		p := NewPod(arg)
		objects = append(objects, &p)
	}
	add(PodArg{Name: "crdb-0", Namespace: "db", Labels: map[string]string{"app": "crdb"}, Nodename: "node-a"})
	add(PodArg{Name: "crdb-1", Namespace: "db", Labels: map[string]string{"app": "crdb"}, Nodename: "node-b"})
	add(PodArg{Name: "web", Namespace: "apps", Nodename: "node-a"})
	add(PodArg{Name: "pending", Namespace: "apps", Nodename: "node-a", Status: v1.PodPending})
	add(PodArg{Name: "daemon", Namespace: "chaos-mesh", Nodename: "node-a"})
	add(PodArg{Name: "other-node", Namespace: "apps", Nodename: "node-b"})
	hostNet := NewPod(PodArg{Name: "cilium", Namespace: "kube-system", Nodename: "node-a"})
	hostNet.Spec.HostNetwork = true
	objects = append(objects, &hostNet)

	impl := newSelector(objects, generic.Option{ClusterScoped: true}, "chaos-mesh")
	victim := v1alpha1.PodSelector{
		Selector: v1alpha1.PodSelectorSpec{
			Pods: map[string][]string{"db": {"crdb-0"}},
		},
		Mode: v1alpha1.AllMode,
	}

	targets, err := impl.Select(context.Background(), &v1alpha1.NodeScopedPodSelector{PodSelector: victim})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ids(targets)).To(Equal([]string{"apps/web", "db/crdb-0"}))

	targets, err = impl.Select(context.Background(), &v1alpha1.NodeScopedPodSelector{PodSelector: victim, IncludeHost: true})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ids(targets)).To(Equal([]string{"apps/web", "db/crdb-0", "node:node-a"}))
}

func TestSelectCoversEveryVictimNode(t *testing.T) {
	g := NewWithT(t)

	var objects []runtime.Object
	for _, arg := range []PodArg{
		{Name: "crdb-0", Namespace: "db", Labels: map[string]string{"app": "crdb"}, Nodename: "node-a"},
		{Name: "crdb-1", Namespace: "db", Labels: map[string]string{"app": "crdb"}, Nodename: "node-b"},
		{Name: "web", Namespace: "apps", Nodename: "node-b"},
	} {
		p := NewPod(arg)
		objects = append(objects, &p)
	}

	impl := newSelector(objects, generic.Option{ClusterScoped: true})
	targets, err := impl.Select(context.Background(), &v1alpha1.NodeScopedPodSelector{
		PodSelector: v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
					LabelSelectors: map[string]string{"app": "crdb"},
				},
			},
			Mode: v1alpha1.AllMode,
		},
		IncludeHost: true,
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ids(targets)).To(Equal([]string{"db/crdb-0", "apps/web", "db/crdb-1", "node:node-a", "node:node-b"}))
}

func TestSelectRespectsTargetNamespace(t *testing.T) {
	g := NewWithT(t)

	var objects []runtime.Object
	for _, arg := range []PodArg{
		{Name: "crdb-0", Namespace: "db", Nodename: "node-a"},
		{Name: "web", Namespace: "apps", Nodename: "node-a"},
	} {
		p := NewPod(arg)
		objects = append(objects, &p)
	}

	impl := newSelector(objects, generic.Option{ClusterScoped: false, TargetNamespace: "db"})
	targets, err := impl.Select(context.Background(), &v1alpha1.NodeScopedPodSelector{
		PodSelector: v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{Pods: map[string][]string{"db": {"crdb-0"}}},
			Mode:     v1alpha1.AllMode,
		},
	})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ids(targets)).To(Equal([]string{"db/crdb-0"}))
}
