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
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"
	listers "github.com/pingcap/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"github.com/pingcap/chaos-operator/pkg/manager/podchaos"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// PodChaosManagerInterface defines manager functions to manager pod chaos job.
type PodChaosManagerInterface interface {
	Sync(pc *v1alpha1.PodChaos) error
	Delete(key string) error
}

type podChaosControl struct {
	mgr PodChaosManagerInterface
}

// NewPodChaosControl returns a new instance of podChaosControl.
func NewPodChaosControl(
	kubeCli kubernetes.Interface,
	cli versioned.Interface,
	mgr manager.ManagerBaseInterface,
	podLister corelisters.PodLister,
	lister listers.PodChaosLister,
) *podChaosControl {
	return &podChaosControl{
		mgr: podchaos.NewPodChaosManager(kubeCli, cli, mgr, podLister, lister),
	}
}

// UpdatePodChaos executes the core logic loop for a PodChaos.
// It will sync the PodChaos resource to PodChaosManager and update status to kubernetes.
func (p *podChaosControl) UpdatePodChaos(pc *v1alpha1.PodChaos) error {
	var errs []error
	oldStatus := pc.Status.DeepCopy()

	if err := p.mgr.Sync(pc); err != nil {
		errs = append(errs, err)
	}

	if apiequality.Semantic.DeepEqual(&pc.Status, oldStatus) {
		return errorutils.NewAggregate(errs)
	}

	return errorutils.NewAggregate(errs)
}

func (p *podChaosControl) DeletePodChaos(key string) error {
	return p.mgr.Delete(key)
}
