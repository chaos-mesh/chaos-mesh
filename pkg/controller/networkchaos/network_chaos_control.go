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

// TODO: some of these codes are copied directly from podchaos. Refractor is needed for reusing code

package networkchaos

import (
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"
	listers "github.com/pingcap/chaos-operator/pkg/client/listers/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"github.com/pingcap/chaos-operator/pkg/manager/networkchaos"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// NetworkChaosManagerInterface defines manager functions to manager pod chaos job.
type NetworkChaosManagerInterface interface {
	Sync(nc *v1alpha1.NetworkChaos) error
	Delete(key string) error
}

type networkChaosControl struct {
	mgr NetworkChaosManagerInterface
}

// NewNetworkChaosControl returns a new instance of networkChaosControl.
func NewNetworkChaosControl(
	kubeCli kubernetes.Interface,
	cli versioned.Interface,
	mgr manager.ManagerBaseInterface,
	podLister corelisters.PodLister,
	lister listers.NetworkChaosLister,
) *networkChaosControl {
	return &networkChaosControl{
		mgr: networkchaos.NewNetworkChaosManager(kubeCli, cli, mgr, podLister, lister),
	}
}

// UpdateNetworkChaos executes the core logic loop for a NetworkChaos.
// It will sync the NetworkChaos resource to NetworkChaosManager and update status to kubernetes.
func (n *networkChaosControl) UpdateNetworkChaos(nc *v1alpha1.NetworkChaos) error {
	var errs []error
	oldStatus := nc.Status.DeepCopy()

	if err := n.mgr.Sync(nc); err != nil {
		errs = append(errs, err)
	}

	if apiequality.Semantic.DeepEqual(&nc.Status, oldStatus) {
		return errorutils.NewAggregate(errs)
	}

	// if err := p.statusUpdater.UpdateStatus(); err != nil {
	// 	errs = append(errs, err)
	// }

	return errorutils.NewAggregate(errs)
}

func (n *networkChaosControl) DeleteNetworkChaos(key string) error {
	return n.mgr.Delete(key)
}
