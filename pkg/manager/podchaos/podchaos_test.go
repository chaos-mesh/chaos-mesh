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
	"github.com/robfig/cron/v3"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestPodChaosManager(t *testing.T) {
	g := NewGomegaWithT(t)

	pcManager := newFakePodChaosManager()

	pcKill := &v1alpha1.PodChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodChaos",
			APIVersion: "pingcap.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-kill-test",
			Namespace: metav1.NamespaceDefault,
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

	g.Expect(pcManager.Sync(pcKill)).Should(Succeed())
	key, err := cache.MetaNamespaceKeyFunc(pcKill)
	g.Expect(err).ShouldNot(HaveOccurred())

	getRn, exist := pcManager.base.GetRunner(key)
	g.Expect(exist).To(Equal(true))
	g.Expect(getRn.EntryID).NotTo(Equal(0))

	pcKillCopy := pcKill.DeepCopy()
	g.Expect(pcManager.Sync(pcKillCopy)).Should(Succeed())
	getRn2, exist := pcManager.base.GetRunner(key)
	g.Expect(exist).To(Equal(true))
	g.Expect(getRn2.EntryID).To(Equal(getRn.EntryID))

	pcKillCopy.Spec.Scheduler.Cron = "@every 2m"
	g.Expect(pcManager.Sync(pcKillCopy)).Should(Succeed())
	getRn3, exist := pcManager.base.GetRunner(key)
	g.Expect(exist).To(Equal(true))
	g.Expect(getRn3.EntryID).NotTo(Equal(getRn.EntryID))

	g.Expect(pcManager.Delete(key)).Should(Succeed())
	_, exist = pcManager.base.GetRunner(key)
	g.Expect(exist).To(Equal(false))
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
	pcManager := NewPodChaosManager(kubeCli, managerBase, podLister, pcLister)

	return pcManager
}
