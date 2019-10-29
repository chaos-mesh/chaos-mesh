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

package networkchaos

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"

	listers "github.com/pingcap/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

type networkChaosManager struct {
	base      manager.ManagerBaseInterface
	kubeCli   kubernetes.Interface
	cli       versioned.Interface
	podLister corelisters.PodLister
	ncLister  listers.NetworkChaosLister
}

// NewNetworkChaosManager returns an instance of networkChaosManager.
// This manager will manage all NetworkChaos task.
func NewNetworkChaosManager(
	kubeCli kubernetes.Interface,
	cli versioned.Interface,
	base manager.ManagerBaseInterface,
	podLister corelisters.PodLister,
	lister listers.NetworkChaosLister,
) *networkChaosManager {
	return &networkChaosManager{
		kubeCli:   kubeCli,
		cli:       cli,
		base:      base,
		podLister: podLister,
		ncLister:  lister,
	}
}

// Sync syncs the NetworkChaos resource to manager.
func (m *networkChaosManager) Sync(nc *v1alpha1.NetworkChaos) error {
	key, err := cache.MetaNamespaceKeyFunc(nc)
	if err != nil {
		return err
	}

	runner, err := m.newRunner(nc)
	if err != nil {
		return err
	}

	if rn, exist := m.base.GetRunner(key); exist {
		if rn.Equal(runner) {
			return nil
		}

		glog.Infof("Update the runner %s", key)
		return m.base.UpdateRunner(runner)
	}

	glog.Infof("Add a new runner for %s", key)
	return m.base.AddRunner(runner)
}

func (m *networkChaosManager) Delete(key string) error {
	glog.Infof("Delete the runner %s", key)
	return m.base.DeleteRunner(key)
}

func (m *networkChaosManager) newRunner(nc *v1alpha1.NetworkChaos) (*manager.Runner, error) {
	var job manager.Job

	switch nc.Spec.Action {
	case v1alpha1.DelayAction:
		job = &DelayJob{
			networkChaos: nc,
			kubeCli:      m.kubeCli,
			podLister:    m.podLister,
		}
	default:
		return nil, fmt.Errorf("NetworkChaos action %s not supported", nc.Spec.Action)
	}

	name, err := cache.MetaNamespaceKeyFunc(nc)
	if err != nil {
		return nil, err
	}

	return &manager.Runner{
		Name: name,
		Rule: nc.Spec.Scheduler.Cron,
		Job:  job,
	}, nil
}
