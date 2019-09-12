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
	"github.com/cwen0/chaos-operator/pkg/apis/pingcap.com/v1alpha1"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
)

// ManagerInterface defines manager functions to manager pod chaos job.
type ManagerInterface interface {
	Sync(pc *v1alpha1.PodChaos) error
}

// StatusUpdaterInterface defines a function to update PodChaos status.
type StatusUpdaterInterface interface {
	UpdateStatus() error
}

type podChaosControl struct {
	statusUpdater StatusUpdaterInterface
	mgr           ManagerInterface
}

// NewPodChaosControl returns a new instance of podChaosControl.
func NewPodChaosControl() *podChaosControl {
	return &podChaosControl{}
}

// UpdatePodChaos executes the core logic loop for a PodChaos.
// It will sync the PodChaos resource to PodChaosManager and update status to kubernetes.
func (p *podChaosControl) UpdatePodChaos(pc *v1alpha1.PodChaos) error {
	var errs []error
	oldStatus := pc.Status.DeepCopy()

	if err := p.sync(pc); err != nil {
		errs = append(errs, err)
	}

	if apiequality.Semantic.DeepEqual(&pc.Status, oldStatus) {
		return errorutils.NewAggregate(errs)
	}

	if err := p.statusUpdater.UpdateStatus(); err != nil {
		errs = append(errs, err)
	}

	return errorutils.NewAggregate(errs)
}

// sync a PodChaos object to manager.
func (p *podChaosControl) sync(pc *v1alpha1.PodChaos) error {
	return p.mgr.Sync(pc)
}
