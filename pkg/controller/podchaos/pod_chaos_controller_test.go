// Copyright 2019 PingCAP, Inc.
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

package podchaos

import (
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned/fake"
	informers "github.com/pingcap/chaos-operator/pkg/client/informers/externalversions"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
)

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

func TestCreatePodChaos(t *testing.T) {
	f := newFixture(t)

	pc := newPodChaos("pod-kill-test")
	f.podChaosLister = append(f.podChaosLister, pc)
	f.objects = append(f.objects, pc)

	f.expectCreatePodChaosAction(pc)
	f.run(getKey(pc, f.g))
}

func TestEnqueuePodChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	pc := newPodChaos("test")

	client := fake.NewSimpleClientset()
	kubeclient := kubefake.NewSimpleClientset()

	i := informers.NewSharedInformerFactory(client, noResyncPeriodFunc())
	kubeI := kubeinformers.NewSharedInformerFactory(kubeclient, noResyncPeriodFunc())

	managerBase := &fakeManagerBase{}
	c := NewController(kubeclient, client, kubeI, i, managerBase)

	c.enqueuePodChaos(pc)
	g.Expect(c.queue.Len()).To(Equal(1))
}

func newPodChaos(name string) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: corev1.NamespaceDefault,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{"chaos-testing"},
			},
			Scheduler: v1alpha1.SchedulerSpec{
				Cron: "@every 1m",
			},
			Action: v1alpha1.PodKillAction,
			Mode:   v1alpha1.OnePodMode,
		},
	}
}

type fixture struct {
	g          *GomegaWithT
	client     *fake.Clientset
	kubeclient *kubefake.Clientset

	podChaosLister []*v1alpha1.PodChaos

	actions []core.Action

	objects []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}

	f.g = NewGomegaWithT(t)
	f.objects = []runtime.Object{}

	return f
}

func (f *fixture) newController() (*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = kubefake.NewSimpleClientset()

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	kubeI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())

	managerBase := &fakeManagerBase{}
	c := NewController(f.kubeclient, f.client, kubeI, i, managerBase)

	c.pcsSynced = alwaysReady
	c.podsSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, f := range f.podChaosLister {
		_ = i.Pingcap().V1alpha1().PodChaoses().Informer().GetIndexer().Add(f)
	}

	return c, i, kubeI
}

func (f *fixture) run(pcName string) {
	f.runController(pcName)
}

func (f *fixture) runExpectError(pcName string) {
	f.runController(pcName)
}

func (f *fixture) runController(pcName string) {
	c, i, kubeI := f.newController()

	stopCh := make(chan struct{})
	defer close(stopCh)

	i.Start(stopCh)
	kubeI.Start(stopCh)

	err := c.syncHandler(pcName)
	f.g.Expect(err).ShouldNot(HaveOccurred())

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		f.g.Expect(f.actions).Should(BeNumerically(">=", i+1))

		expectAction := f.actions[i]
		checkAction(expectAction, action, f.g)
	}

	// TODO: check update action
	// f.g.Expect(len(f.actions)).Should(BeNumerically("<=", len(actions)))
}

func (f *fixture) expectCreatePodChaosAction(pc *v1alpha1.PodChaos) {
	f.actions = append(f.actions, core.NewCreateAction(schema.GroupVersionResource{Resource: "podchaoses"}, pc.Namespace, pc))
}

func (f *fixture) expectUpdatePodChaosAction(pc *v1alpha1.PodChaos) {
	f.actions = append(f.actions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "podchaoses"}, pc.Namespace, pc))
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, g *GomegaWithT) {
	g.Expect(expected.Matches(actual.GetVerb(), actual.GetResource().Resource)).To(BeTrue())
	g.Expect(actual.GetSubresource()).To(Equal(expected.GetSubresource()))

	g.Expect(reflect.TypeOf(actual)).To(Equal(reflect.TypeOf(expected)))

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		g.Expect(object).To(Equal(expObject))
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		g.Expect(object).To(Equal(expObject))
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		g.Expect(patch).To(Equal(expPatch))
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	var ret []core.Action
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "podchaoses") ||
				action.Matches("watch", "podchaoses")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func getKey(pc *v1alpha1.PodChaos, g *GomegaWithT) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(pc)
	g.Expect(err).ShouldNot(HaveOccurred())
	return key
}
