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
	"math/rand"

	"github.com/cwen0/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/cwen0/chaos-operator/pkg/manager"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type podKillJob struct {
	podChaos  *v1alpha1.PodChaos
	kubeCli   kubernetes.Interface
	podLister corelisters.PodLister
}

func (p podKillJob) Run() {
	var err error

	// TODO: support more modes
	switch p.podChaos.Spec.Mode {
	case v1alpha1.OnePodMode:
		err = p.deleteRandomPod()
	default:
		err = fmt.Errorf("pod-kill mode %s not supported", p.podChaos.Spec.Mode)
	}

	utilruntime.HandleError(err)
}

func (p *podKillJob) deleteRandomPod() error {
	pods, err := manager.SelectPods(p.podChaos.Spec.Selector, p.podLister, p.kubeCli)
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("selected pods is empty")
	}

	index := rand.Intn(len(pods))

	return p.deletePod(pods[index])
}

func (p *podKillJob) deletePod(pod *v1.Pod) error {
	deleteOpts := p.getDeleteOptsForPod(pod)

	return p.kubeCli.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOpts)
}

// Creates the DeleteOptions object for the pod. Grace period is calculated as the higher
// of configured grace period and termination grace period set on the pod
func (p *podKillJob) getDeleteOptsForPod(pod *v1.Pod) *metav1.DeleteOptions {
	gracePeriodSec := &p.podChaos.Spec.GracePeriodSeconds

	if pod.Spec.TerminationGracePeriodSeconds != nil &&
		*pod.Spec.TerminationGracePeriodSeconds > *gracePeriodSec {
		gracePeriodSec = pod.Spec.TerminationGracePeriodSeconds
	}

	return &metav1.DeleteOptions{
		GracePeriodSeconds: gracePeriodSec,
	}
}
