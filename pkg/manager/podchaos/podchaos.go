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

	"github.com/cwen0/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	listers "github.com/cwen0/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/cwen0/chaos-operator/pkg/manager"
	"github.com/golang/glog"

	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type podChaosManager struct {
	base      manager.ManagerBaseInterface
	kubeCli   kubernetes.Interface
	podLister corelisters.PodLister
	pcLister  listers.PodChaosLister
}

// NewPodChaosManager returns a instance of podChaosManager.
// This manager will manage all PodChaos task.
func NewPodChaosManager(
	kubeCli kubernetes.Interface,
	base manager.ManagerBaseInterface,
	podLister corelisters.PodLister,
	lister listers.PodChaosLister,
) *podChaosManager {
	return &podChaosManager{
		kubeCli:   kubeCli,
		base:      base,
		podLister: podLister,
		pcLister:  lister,
	}
}

// Sync syncs the PodChaos resource to manager.
func (m *podChaosManager) Sync(pc *v1alpha1.PodChaos) error {
	key, err := cache.MetaNamespaceKeyFunc(pc)
	if err != nil {
		return err
	}

	runner, err := m.newRunner(pc)
	if err != nil {
		return err
	}

	if m.base.IsExist(key) {
		glog.Infof("Sync the runner %s", key)
		return m.base.UpdateRunner(runner)
	}

	glog.Infof("Add a new runner for %s", key)
	return m.base.AddRunner(runner)
}

func (m *podChaosManager) Delete(key string) error {
	glog.Infof("Delete the runner %s", key)
	return m.base.DeleteRunner(key)
}

func (m *podChaosManager) newRunner(pc *v1alpha1.PodChaos) (*manager.Runner, error) {
	var job manager.Job

	switch pc.Spec.Action {
	case v1alpha1.PodKillAction:
		job = podKillJob{
			podChaos:  pc,
			kubeCli:   m.kubeCli,
			podLister: m.podLister,
		}
	case v1alpha1.PodFailureAction:
		fallthrough
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
