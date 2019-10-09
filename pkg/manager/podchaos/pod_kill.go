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
	"math"
	"math/rand"
	"strconv"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/manager"

	"golang.org/x/sync/errgroup"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

// PodKillJob defines a job to do pod-kill chaos experiment.
type PodKillJob struct {
	podChaos  *v1alpha1.PodChaos
	kubeCli   kubernetes.Interface
	podLister corelisters.PodLister
}

// Run is the core logic to execute pod-kill chaos experiment.
func (p *PodKillJob) Run() {
	var err error

	pods, err := manager.SelectPods(p.podChaos.Spec.Selector, p.podLister, p.kubeCli)
	if err != nil {
		glog.Errorf("%s, fail to get selected pods, %v", p.logPrefix(), err)
		return
	}

	if pods == nil || len(pods) == 0 {
		glog.Errorf("%s, no pod is selected", p.logPrefix())
		return
	}

	switch p.podChaos.Spec.Mode {
	case v1alpha1.OnePodMode:
		glog.Infof("%s, Try to select one pod to do pod-kill job randomly", p.logPrefix())
		err = p.deleteRandomPod(pods)
	case v1alpha1.AllPodMode:
		glog.Infof("%s, Try to do pod-kill action on all filtered pods", p.logPrefix())
		err = p.deleteAllPods(pods)
	case v1alpha1.FixedPodMode:
		glog.Infof("%s, Try to do pod-kill action on %s pods", p.logPrefix(), p.podChaos.Spec.Value)
		err = p.deleteFixedPods(pods)
	case v1alpha1.FixedPercentPodMode:
		glog.Infof("%s, Try to do pod-kill action on %s%% pods", p.logPrefix(), p.podChaos.Spec.Value)
		err = p.deleteFixedPercentagePods(pods)
	case v1alpha1.RandomMaxPercentPodMode:
		glog.Infof("%s, Try to do pod-kill action on max %s%% pods", p.logPrefix(), p.podChaos.Spec.Value)
		err = p.deleteMaxPercentagePods(pods)
	default:
		err = fmt.Errorf("pod-kill mode %s not supported", p.podChaos.Spec.Mode)
	}

	if err != nil {
		glog.Errorf("%s, fail to run action, %v", p.logPrefix(), err)
	}
}

// Equal returns true when the two jobs have same PodChaos.
// It can be used to judge if the job need to update this job.
func (p *PodKillJob) Equal(job manager.Job) bool {
	pjob, ok := job.(*PodKillJob)
	if !ok {
		return false
	}

	if p.podChaos.Name != pjob.podChaos.Name ||
		p.podChaos.Namespace != pjob.podChaos.Namespace {
		return false
	}

	// judge ResourceVersion,
	// If them are same, we can think that the PodChaos resource has not changed.
	if p.podChaos.ResourceVersion != pjob.podChaos.ResourceVersion {
		return false
	}

	return true
}

// Stop stops pod-kill job, because pod-kill is a transient operation
// and the pod will be maintained by kubernetes,
// so we don't need clean anything.
func (p *PodKillJob) Close() error { return nil }

func (p *PodKillJob) deleteAllPods(pods []v1.Pod) error {
	g := errgroup.Group{}
	for _, pod := range pods {
		pod := pod
		g.Go(func() error {
			return p.deletePod(pod)
		})
	}

	return g.Wait()
}

func (p *PodKillJob) deleteFixedPods(pods []v1.Pod) error {
	killNum, err := strconv.Atoi(p.podChaos.Spec.Value)
	if err != nil {
		return err
	}

	glog.Infof("%s, Try to delete %d pods", p.logPrefix(), killNum)

	if len(pods) < killNum {
		glog.Infof("%s, fixed number is less the count of the selected pods", p.logPrefix())
		killNum = len(pods)
	}

	return p.concurrentDeletePods(pods, killNum)
}

func (p *PodKillJob) deleteFixedPercentagePods(pods []v1.Pod) error {
	killPercentage, err := strconv.Atoi(p.podChaos.Spec.Value)
	if err != nil {
		return err
	}

	if killPercentage == 0 {
		glog.V(6).Infof("%s, Not terminating any pods to do pod-kill action as fixed percentage is 0",
			p.logPrefix())
		return nil
	}

	if killPercentage < 0 || killPercentage > 100 {
		return fmt.Errorf("fixed percentage value of %d is invalid, Must be [0-100]", killPercentage)
	}

	killNum := int(math.Floor(float64(len(pods)) * float64(killPercentage) / 100))

	return p.concurrentDeletePods(pods, killNum)
}

func (p *PodKillJob) deleteMaxPercentagePods(pods []v1.Pod) error {
	maxPercentage, err := strconv.Atoi(p.podChaos.Spec.Value)
	if err != nil {
		return err
	}

	if maxPercentage == 0 {
		glog.V(6).Infof("%s, Not terminating any pods to do pod-kill action as fixed percentage is 0",
			p.logPrefix())
		return nil
	}

	if maxPercentage < 0 || maxPercentage > 100 {
		return fmt.Errorf("fixed percentage value of %d is invalid, Must be [0-100]", maxPercentage)
	}

	killPercentage := rand.Intn(maxPercentage + 1) // + 1 because Intn works with half open interval [0,n) and we want [0,n]
	killNum := int(math.Floor(float64(len(pods)) * float64(killPercentage) / 100))

	return p.concurrentDeletePods(pods, killNum)
}

func (p *PodKillJob) deleteRandomPod(pods []v1.Pod) error {
	if len(pods) == 0 {
		return nil
	}

	index := rand.Intn(len(pods))
	return p.deletePod(pods[index])
}

func (p *PodKillJob) deletePod(pod v1.Pod) error {
	glog.Infof("%s, Try to delete pod %s/%s", p.logPrefix(), pod.Namespace, pod.Name)

	deleteOpts := p.getDeleteOptsForPod(pod)

	return p.kubeCli.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOpts)
}

func (p *PodKillJob) concurrentDeletePods(pods []v1.Pod, killNum int) error {
	if killNum < 0 {
		return nil
	}

	killIndexes := manager.RandomFixedIndexes(0, uint(len(pods)), uint(killNum))

	g := errgroup.Group{}
	for _, index := range killIndexes {
		index := index
		g.Go(func() error {
			return p.deletePod(pods[index])
		})
	}

	return g.Wait()
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
