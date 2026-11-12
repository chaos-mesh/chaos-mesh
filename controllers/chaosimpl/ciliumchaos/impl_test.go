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

package ciliumchaos

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func ciliumPodOn(name, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "kube-system",
			Labels:    map[string]string{"k8s-app": "cilium"},
		},
		Spec: corev1.PodSpec{NodeName: nodeName},
	}
}

func newCiliumChaos() *v1alpha1.CiliumChaos {
	return &v1alpha1.CiliumChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "az-failure", Namespace: "chaos"},
		Spec: v1alpha1.CiliumChaosSpec{
			CiliumPodSelector: &v1alpha1.CiliumPodSelectorSpec{
				Namespace:      "kube-system",
				LabelSelectors: map[string]string{"k8s-app": "cilium"},
			},
		},
	}
}

func TestRecoverSkipsDeletedNode(t *testing.T) {
	g := NewWithT(t)

	// node-b was removed while the chaos was injected, taking its cilium pod with it
	objs := []client.Object{
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}},
		ciliumPodOn("cilium-a", "node-a"),
	}
	impl := &Impl{
		Reader: fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(objs...).Build(),
		Log:    logr.Discard(),
	}
	records := []*v1alpha1.Record{{Id: "node-b", Phase: v1alpha1.Injected}}

	phase, err := impl.Recover(context.TODO(), 0, records, newCiliumChaos())

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func TestRecoverRetriesWhenCiliumPodMissingOnExistingNode(t *testing.T) {
	g := NewWithT(t)

	objs := []client.Object{
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-b"}},
		ciliumPodOn("cilium-a", "node-a"),
	}
	impl := &Impl{
		Reader: fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(objs...).Build(),
		Log:    logr.Discard(),
	}
	records := []*v1alpha1.Record{{Id: "node-b", Phase: v1alpha1.Injected}}

	phase, err := impl.Recover(context.TODO(), 0, records, newCiliumChaos())

	g.Expect(err).To(MatchError(ContainSubstring("received unexpected number of cilium pods: 0")))
	g.Expect(phase).To(Equal(v1alpha1.Injected))
}
