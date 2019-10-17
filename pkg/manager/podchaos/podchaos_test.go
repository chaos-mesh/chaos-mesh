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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned/fake"
	informers "github.com/pingcap/chaos-operator/pkg/client/informers/externalversions"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"github.com/robfig/cron/v3"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestPodChaosManagerSync(t *testing.T) {
	g := NewGomegaWithT(t)

	pcManager := newFakePodChaosManager()

	type TestCase struct {
		name           string
		podchaos       *v1alpha1.PodChaos
		expectedResult resultF
		isChange       bool
	}

	tcsNew := []TestCase{
		{
			name:           "new podchaos 1",
			podchaos:       newPodChaos2("pod-chaos-1", "@every 2m", v1alpha1.SelectorSpec{Namespaces: []string{"chaos-testing"}}),
			expectedResult: Succeed,
		},
		{
			name:           "new podchaos 2",
			podchaos:       newPodChaos2("pod-chaos-2", "@every 2m", v1alpha1.SelectorSpec{Namespaces: []string{"chaos-testing"}}),
			expectedResult: Succeed,
		},
	}

	for _, tc := range tcsNew {
		key, err := cache.MetaNamespaceKeyFunc(tc.podchaos)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)

		_, exist := pcManager.base.GetRunner(key)
		g.Expect(exist).To(Equal(false), tc.name)

		g.Expect(pcManager.Sync(tc.podchaos)).Should(tc.expectedResult(), tc.name)

		_, exist = pcManager.base.GetRunner(key)
		g.Expect(exist).To(Equal(true), tc.name)
	}

	tcsUpdate := []TestCase{
		{
			name:           "same pochaos",
			podchaos:       newPodChaos2("pod-chaos-1", "@every 2m", v1alpha1.SelectorSpec{Namespaces: []string{"chaos-testing"}}),
			expectedResult: Succeed,
			isChange:       false,
		},
		{
			name:           "different rule",
			podchaos:       newPodChaos2("pod-chaos-1", "@every 4m", v1alpha1.SelectorSpec{Namespaces: []string{"chaos-testing"}}),
			expectedResult: Succeed,
			isChange:       true,
		},
		{
			name:           "different selector",
			podchaos:       newPodChaos2("pod-chaos-2", "@every 2m", v1alpha1.SelectorSpec{Namespaces: []string{"chaos"}}),
			expectedResult: Succeed,
			isChange:       true,
		},
	}

	for _, tc := range tcsUpdate {
		key, err := cache.MetaNamespaceKeyFunc(tc.podchaos)
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)

		expectedID, exist := pcManager.base.GetRunner(key)
		g.Expect(exist).To(Equal(true), tc.name)

		g.Expect(pcManager.Sync(tc.podchaos)).Should(tc.expectedResult(), tc.name)

		getID, exist := pcManager.base.GetRunner(key)
		g.Expect(exist).To(Equal(true), tc.name)

		if tc.isChange {
			g.Expect(getID).NotTo(Equal(expectedID), tc.name)
		} else {
			g.Expect(getID).To(Equal(expectedID), tc.name)
		}
	}
}

func TestPodChaosManagerDelete(t *testing.T) {
	g := NewGomegaWithT(t)

	pcManager := newFakePodChaosManager()

	pcs := []*v1alpha1.PodChaos{newPodChaos("pc-1"), newPodChaos("pc-2"), newPodChaos("pc-3")}

	for _, pc := range pcs {
		g.Expect(pcManager.Sync(pc)).Should(Succeed(), pc.Name)
	}

	type TestCase struct {
		name           string
		key            string
		expectedResult resultF
		isExist        bool
	}

	tcsNew := []TestCase{
		{
			name:           "delete pc-1",
			isExist:        true,
			key:            fmt.Sprintf("%s/pc-1", metav1.NamespaceDefault),
			expectedResult: Succeed,
		},
		{
			name:           "delete pc-2",
			isExist:        true,
			key:            fmt.Sprintf("%s/pc-2", metav1.NamespaceDefault),
			expectedResult: Succeed,
		},
		{
			name:           "podchaos not exist",
			isExist:        false,
			key:            fmt.Sprintf("%s/pc-not-exist", metav1.NamespaceDefault),
			expectedResult: Succeed,
		},
	}

	for _, tc := range tcsNew {
		_, exist := pcManager.base.GetRunner(tc.key)
		g.Expect(exist).To(Equal(tc.isExist), tc.name)

		g.Expect(pcManager.Delete(tc.key)).Should(tc.expectedResult(), tc.name)

		_, exist = pcManager.base.GetRunner(tc.key)
		g.Expect(exist).To(Equal(false), tc.name)
	}
}

func newFakePodChaosManager() *podChaosManager {
	kubeCli := kubefake.NewSimpleClientset()
	cli := fake.NewSimpleClientset()
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeCli, 0)
	informerFactory := informers.NewSharedInformerFactory(cli, 0)

	podLister := kubeInformerFactory.Core().V1().Pods().Lister()
	pcLister := informerFactory.Pingcap().V1alpha1().PodChaoses().Lister()

	cronEngine := cron.New()
	cronEngine.Start()

	managerBase := manager.NewManagerBase(cronEngine)
	pcManager := NewPodChaosManager(kubeCli, cli, managerBase, podLister, pcLister)

	return pcManager
}

func newPodChaos2(name string, rule string, selector v1alpha1.SelectorSpec) *v1alpha1.PodChaos {
	return &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: selector,
			Scheduler: v1alpha1.SchedulerSpec{
				Cron: rule,
			},
			Action: v1alpha1.PodKillAction,
		},
	}
}

func TestUpdatePodChaosStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	pc := newPodChaos("test")

	pcm := newFakePodChaosManager()
	pcm.cli = fake.NewSimpleClientset(pc)

	pct := pc.DeepCopy()
	pct.Status.Phase = v1alpha1.ChaosPhaseAbnormal
	pct.Status.Reason = "t"
	pct.Status.Experiment = v1alpha1.PodChaosExperimentStatus{
		StartTime: metav1.Now(),
	}

	g.Expect(pcm.updatePodChaosStatus(pct)).ShouldNot(HaveOccurred())
	getPc, err := pcm.cli.PingcapV1alpha1().PodChaoses(pct.Namespace).
		Get(pct.Name, metav1.GetOptions{})
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(getPc.Status.Phase).To(Equal(pct.Status.Phase))
	g.Expect(getPc.Status.Reason).To(Equal(pct.Status.Reason))
	g.Expect(getPc.Status.Experiment).ToNot(Equal(pct.Status.Experiment))
}
