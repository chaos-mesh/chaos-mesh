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
	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// DelayJob represents a job to add delay on pod
type DelayJob struct {
	networkChaos *v1alpha1.NetworkChaos
	kubeCli      kubernetes.Interface
	podLister    corelisters.PodLister
}

// Run is the core logic to execute delay chaos experiment.
func (d *DelayJob) Run() {
	glog.Info("DELAY RUNNING")
}

// Equal returns true when the two jobs have same NetworkChaos.
// It can be used to judge if the job need to update this job.
func (d *DelayJob) Equal(job manager.Job) bool {
	djob, ok := job.(*DelayJob)
	if !ok {
		return false
	}

	if d.networkChaos.Name != djob.networkChaos.Name ||
		d.networkChaos.Namespace != djob.networkChaos.Namespace {
		return false
	}

	// judge ResourceVersion,
	// If them are same, we can think that the NetworkChaos resource has not changed.
	if d.networkChaos.ResourceVersion != djob.networkChaos.ResourceVersion {
		return false
	}

	return true
}

// Close stops delay job, we need to clean up running delay job (by removing every netem)
func (d *DelayJob) Close() error {
	// TODO: clean up
	return nil
}
