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
	"github.com/golang/glog"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

type PodKillJob struct {
	podChaos  *v1alpha1.PodChaos
	kubeCli   kubernetes.Interface
	podLister corelisters.PodLister
}

func (p PodKillJob) Run() {
	var err error

	pods, err := manager.SelectPods(p.podChaos.Spec.Selector, p.podLister, p.kubeCli)
	if err != nil {
		glog.Errorf("%s, fail to get selected pods, %v", p.logPrefix(), err)
		return
	}

	if len(pods) == 0 {
		glog.Errorf("%s, no pod is selected", p.logPrefix())
		return
	}

	// TODO: support more modes
	switch p.podChaos.Spec.Mode {
	case v1alpha1.OnePodMode:
		glog.Infof("%s, Try to select one pod to do pod-kill job randomly", p.logPrefix())
		err = p.deleteRandomPod(pods)
	case v1alpha1.AllPodMode:
		glog.Infof("%s, Try to do pod-kill action on all filtered pods", p.logPrefix())
		err = p.deleteAllPods(pods)
	default:
		err = fmt.Errorf("pod-kill mode %s not supported", p.podChaos.Spec.Mode)
	}

	if err != nil {
		glog.Errorf("%s, fail to run action, %v", p.logPrefix(), err)
	}
}

func (p PodKillJob) Equal(job manager.Job) bool {
	pjob, ok := job.(PodKillJob)
	if !ok {
		return false
	}

	if p.podChaos.ResourceVersion != pjob.podChaos.ResourceVersion {
		return false
	}

	return true
}

func (p *PodKillJob) deleteAllPods(pods []v1.Pod) error {
	for _, pod := range pods {
		if err := p.deletePod(pod); err != nil {
			return err
		}
	}

	return nil
}

func (p *PodKillJob) deleteRandomPod(pods []v1.Pod) error {
	index := rand.Intn(len(pods))
	return p.deletePod(pods[index])
}

func (p *PodKillJob) deletePod(pod v1.Pod) error {
	glog.Infof("%s, Try to delete pod %s/%s", p.logPrefix(), pod.Namespace, pod.Name)

	deleteOpts := p.getDeleteOptsForPod(pod)

	return p.kubeCli.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOpts)
}

// Creates the DeleteOptions object for the pod. Grace period is calculated as the higher
// of configured grace period and termination grace period set on the pod
func (p *PodKillJob) getDeleteOptsForPod(pod v1.Pod) *metav1.DeleteOptions {
	gracePeriodSec := &p.podChaos.Spec.GracePeriodSeconds

	if pod.Spec.TerminationGracePeriodSeconds != nil &&
		*pod.Spec.TerminationGracePeriodSeconds > *gracePeriodSec {
		gracePeriodSec = pod.Spec.TerminationGracePeriodSeconds
	}

	return &metav1.DeleteOptions{
		GracePeriodSeconds: gracePeriodSec,
	}
}

func (p *PodKillJob) logPrefix() string {
	return fmt.Sprintf("[%s/%s] [action:%s]", p.podChaos.Namespace, p.podChaos.Name, p.podChaos.Spec.Action)
}
