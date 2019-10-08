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
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned/fake"
	informers "github.com/pingcap/chaos-operator/pkg/client/informers/externalversions"
	"github.com/pingcap/chaos-operator/pkg/manager"

	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestUpdatePodChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	pc := newPodChaos("pod-kill")
	control := newFakePodChaosControl()

	g.Expect(control.UpdatePodChaos(pc)).Should(Succeed())
}

func TestDeletePodChaos(t *testing.T) {
	g := NewGomegaWithT(t)

	control := newFakePodChaosControl()

	g.Expect(control.DeletePodChaos("test")).Should(Succeed())
}

func newFakePodChaosControl() *podChaosControl {
	kubeCli := kubefake.NewSimpleClientset()
	cli := fake.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, 0)
	informerFactory := informers.NewSharedInformerFactory(cli, 0)

	podLister := kubeInformerFactory.Core().V1().Pods().Lister()
	pcLister := informerFactory.Pingcap().V1alpha1().PodChaoses().Lister()

	control := NewPodChaosControl(
		kubeCli,
		nil,
		podLister,
		pcLister)

	control.mgr = &fakePodChaosManager{base: &fakeManagerBase{}}

	return control
}

type fakeManagerBase struct{}

func (m *fakeManagerBase) AddRunner(_ *manager.Runner) error { return nil }

func (m *fakeManagerBase) DeleteRunner(_ string) error { return nil }

func (m *fakeManagerBase) UpdateRunner(_ *manager.Runner) error { return nil }

func (m *fakeManagerBase) GetRunner(_ string) (*manager.Runner, bool) { return nil, false }

type fakePodChaosManager struct {
	base manager.ManagerBaseInterface
}

func (p *fakePodChaosManager) Sync(_ *v1alpha1.PodChaos) error { return nil }

func (p *fakePodChaosManager) Delete(_ string) error { return nil }
