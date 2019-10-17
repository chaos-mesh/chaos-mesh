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
	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"math"
	"math/rand"
	"reflect"
	"strconv"

	"golang.org/x/sync/errgroup"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
)

const (
	podKillActionMsg = "delete pod"
)

// PodKillJob defines a job to do pod-kill chaos experiment.
type PodKillJob struct {
	podChaos  *v1alpha1.PodChaos
	kubeCli   kubernetes.Interface
	cli       versioned.Interface
	podLister corelisters.PodLister
}

// Run is the core logic to execute pod-kill chaos experiment.
func (p *PodKillJob) Run() {
	var err error

	record := &v1alpha1.PodChaosExperimentStatus{
		Phase:     v1alpha1.ExperimentPhaseRunning,
		StartTime: metav1.Now(),
	}

	if err := setExperimentRecord(p.cli, p.podChaos, record); err != nil {
		glog.Errorf("%s, fail to set experiment record, %v", p.logPrefix(), err)
	}

	defer func() {
		record.Phase = v1alpha1.ExperimentPhaseFinished
		record.EndTime = metav1.Now()
		if err != nil {
			glog.Errorf("%s, fail to run action, %v", p.logPrefix(), err)
			record.Phase = v1alpha1.ExperimentPhaseFailed
			record.Reason = err.Error()
		}

		if err := setExperimentRecord(p.cli, p.podChaos, record); err != nil {
			glog.Errorf("%s, fail to set experiment record, %v", p.logPrefix(), err)
		}
	}()

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
		err = p.deleteRandomPod(pods, record)
	case v1alpha1.AllPodMode:
		glog.Infof("%s, Try to do pod-kill action on all filtered pods", p.logPrefix())
		err = p.deleteAllPods(pods, record)
	case v1alpha1.FixedPodMode:
		glog.Infof("%s, Try to do pod-kill action on %s pods", p.logPrefix(), p.podChaos.Spec.Value)
		err = p.deleteFixedPods(pods, record)
	case v1alpha1.FixedPercentPodMode:
		glog.Infof("%s, Try to do pod-kill action on %s%% pods", p.logPrefix(), p.podChaos.Spec.Value)
		err = p.deleteFixedPercentagePods(pods, record)
	case v1alpha1.RandomMaxPercentPodMode:
		glog.Infof("%s, Try to do pod-kill action on max %s%% pods", p.logPrefix(), p.podChaos.Spec.Value)
		err = p.deleteMaxPercentagePods(pods, record)
	default:
		err = fmt.Errorf("pod-kill mode %s not supported", p.podChaos.Spec.Mode)
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

	if !reflect.DeepEqual(p.podChaos.Spec, pjob.podChaos.Spec) {
		return false
	}

	return true
}

// Stop stops pod-kill job.
func (p *PodKillJob) Close() error {
	return nil
}

// Clean is used to cleans the residue action.
// Because pod-kill is a transient operation and the pod will be maintained by kubernetes,
// so we don't need clean anything.
func (p *PodKillJob) Clean() error {
	return nil
}

func (p *PodKillJob) deleteAllPods(pods []v1.Pod, record *v1alpha1.PodChaosExperimentStatus) error {
	glog.Infof("%s, Try to delete %d pods", p.logPrefix(), len(pods))

	setRecordPods(record, p.podChaos.Spec.Action, podKillActionMsg, pods...)

	g := errgroup.Group{}
	for _, pod := range pods {
		pod := pod
		g.Go(func() error {
			return p.deletePod(pod)
		})
	}

	return g.Wait()
}

func (p *PodKillJob) deleteFixedPods(pods []v1.Pod, record *v1alpha1.PodChaosExperimentStatus) error {
	killNum, err := strconv.Atoi(p.podChaos.Spec.Value)
	if err != nil {
		return err
	}

	if len(pods) < killNum {
		glog.Infof("%s, Fixed number %d is less the count of the selected pods, set killNum to %d",
			p.logPrefix(), killNum, len(pods))
		killNum = len(pods)
	}

	glog.Infof("%s, Try to delete %d pods", p.logPrefix(), killNum)

	return p.concurrentDeletePods(pods, killNum, record)
}

func (p *PodKillJob) deleteFixedPercentagePods(pods []v1.Pod, record *v1alpha1.PodChaosExperimentStatus) error {
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

	glog.Infof("%s, Try to delete %d pods", p.logPrefix(), killNum)

	return p.concurrentDeletePods(pods, killNum, record)
}

func (p *PodKillJob) deleteMaxPercentagePods(pods []v1.Pod, record *v1alpha1.PodChaosExperimentStatus) error {
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

	glog.Infof("%s, Try to delete %d pods", p.logPrefix(), killNum)

	return p.concurrentDeletePods(pods, killNum, record)
}

func (p *PodKillJob) deleteRandomPod(pods []v1.Pod, record *v1alpha1.PodChaosExperimentStatus) error {
	if len(pods) == 0 {
		return nil
	}

	index := rand.Intn(len(pods))
	pod := pods[index]

	setRecordPods(record, p.podChaos.Spec.Action, podKillActionMsg, pod)

	return p.deletePod(pod)
}

func (p *PodKillJob) deletePod(pod v1.Pod) error {
	glog.Infof("%s, Try to delete pod %s/%s", p.logPrefix(), pod.Namespace, pod.Name)

	deleteOpts := p.getDeleteOptsForPod(pod)

	return p.kubeCli.CoreV1().Pods(pod.Namespace).Delete(pod.Name, deleteOpts)
}

func (p *PodKillJob) concurrentDeletePods(pods []v1.Pod, killNum int, record *v1alpha1.PodChaosExperimentStatus) error {
	if killNum <= 0 || len(pods) <= 0 {
		return nil
	}

	killIndexes := manager.RandomFixedIndexes(0, uint(len(pods)), uint(killNum))

	var filterPods []v1.Pod

	g := errgroup.Group{}
	for _, index := range killIndexes {
		index := index
		filterPods = append(filterPods, pods[index])
		g.Go(func() error {
			return p.deletePod(pods[index])
		})
	}

	setRecordPods(record, p.podChaos.Spec.Action, podKillActionMsg, filterPods...)

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
