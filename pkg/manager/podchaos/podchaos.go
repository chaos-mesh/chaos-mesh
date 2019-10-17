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
	"k8s.io/client-go/util/retry"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"
	listers "github.com/pingcap/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type podChaosManager struct {
	base      manager.ManagerBaseInterface
	kubeCli   kubernetes.Interface
	cli       versioned.Interface
	podLister corelisters.PodLister
	pcLister  listers.PodChaosLister
}

// NewPodChaosManager returns a instance of podChaosManager.
// This manager will manage all PodChaos task.
func NewPodChaosManager(
	kubeCli kubernetes.Interface,
	cli versioned.Interface,
	base manager.ManagerBaseInterface,
	podLister corelisters.PodLister,
	lister listers.PodChaosLister,
) *podChaosManager {
	return &podChaosManager{
		kubeCli:   kubeCli,
		cli:       cli,
		base:      base,
		podLister: podLister,
		pcLister:  lister,
	}
}

// Sync syncs the PodChaos resource to manager.
func (m *podChaosManager) Sync(pc *v1alpha1.PodChaos) error {
	var (
		err    error
		key    string
		runner *manager.Runner
	)

	defer func() {
		if pc.Status.Phase == v1alpha1.ChaosPhaseNone {
			pc.Status.Phase = v1alpha1.ChaosPhaseNormal
			pc.Status.Reason = ""
		}

		if err != nil {
			pc.Status.Phase = v1alpha1.ChaosPhaseAbnormal
			pc.Status.Reason = err.Error()
		}

		if err := m.updatePodChaosStatus(pc); err != nil {
			glog.Errorf("fail to update PodChaos %s status, %v", key, err)
		}
	}()

	key, err = cache.MetaNamespaceKeyFunc(pc)
	if err != nil {
		return err
	}

	runner, err = m.newRunner(pc)
	if err != nil {
		return err
	}

	if rn, exist := m.base.GetRunner(key); exist {
		if rn.Equal(runner) {
			return nil
		}

		glog.Infof("Update the runner %s", key)
		err = m.base.UpdateRunner(runner)
	} else {
		glog.Infof("Add a new runner for %s", key)
		err = m.base.AddRunner(runner)
	}

	return err
}

func (m *podChaosManager) Delete(key string) error {
	glog.Infof("Delete the runner %s", key)
	return m.base.DeleteRunner(key)
}

func (m *podChaosManager) newRunner(pc *v1alpha1.PodChaos) (*manager.Runner, error) {
	var job manager.Job

	switch pc.Spec.Action {
	case v1alpha1.PodKillAction:
		job = &PodKillJob{
			podChaos:  pc,
			kubeCli:   m.kubeCli,
			cli:       m.cli,
			podLister: m.podLister,
		}
	case v1alpha1.PodFailureAction:
		job = &PodFailureJob{
			podChaos:  pc,
			kubeCli:   m.kubeCli,
			cli:       m.cli,
			podLister: m.podLister,
			stopC:     make(chan struct{}),
		}
	default:
		return nil, fmt.Errorf("PodChaos action %s not supported", pc.Spec.Action)
	}

	name, err := cache.MetaNamespaceKeyFunc(pc)
	if err != nil {
		return nil, err
	}

	return &manager.Runner{
		Name: name,
		Rule: pc.Spec.Scheduler.Cron,
		Job:  job,
	}, nil
}

func (m *podChaosManager) updatePodChaosStatus(pc *v1alpha1.PodChaos) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		npc, err := m.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).
			Get(pc.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		npc.Status.Phase = pc.Status.Phase
		npc.Status.Reason = pc.Status.Reason

		_, err = m.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Update(npc)
		return err
	})
}
