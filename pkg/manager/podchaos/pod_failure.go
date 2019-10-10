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
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned"
	"github.com/pingcap/chaos-operator/pkg/manager"
	"github.com/pingcap/chaos-operator/pkg/util"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

const (
	// fakeImage is a not-existing image.
	fakeImage = "pingcap.com/fake-chaos-operator:latest"
)

// PodFailureJob defines a job to do pod-failure chaos experiment.
// It can be used to make certain pods fail for a while.
type PodFailureJob struct {
	podChaos  *v1alpha1.PodChaos
	kubeCli   kubernetes.Interface
	cli       versioned.Interface
	podLister corelisters.PodLister

	cancel *context.CancelFunc
	wg     *sync.WaitGroup

	isRunning int32
}

// Run is the core logic to execute pod-failure chaos experiment.
func (p *PodFailureJob) Run() {
	if !atomic.CompareAndSwapInt32(&p.isRunning, 0, 1) {
		glog.Warningf("%s, ignore this experiment, because the last experiment is still running", p.logPrefix())
		return
	}

	defer func() {
		if err := p.cleanFinalizersAndRecover(); err != nil {
			glog.Errorf("%s, fail to clean finalizer, %v", p.logPrefix(), err)
		}
		atomic.CompareAndSwapInt32(&p.isRunning, 1, 0)
	}()

	var err error

	pods, err := manager.SelectPods(p.podChaos.Spec.Selector, p.podLister, p.kubeCli)
	if err != nil {
		glog.Errorf("%s, fail to get selected pods, %v", p.logPrefix(), err)
	}

	if pods == nil || len(pods) == 0 {
		glog.Errorf("%s, no pod is selected", p.logPrefix())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = &cancel
	p.wg = new(sync.WaitGroup)

	switch p.podChaos.Spec.Mode {
	case v1alpha1.OnePodMode:
		glog.Infof("%s, Try to select one pod to do pod-failure job randomly", p.logPrefix())
		err = p.failRandomPod(ctx, pods)
	case v1alpha1.AllPodMode:
		glog.Infof("%s, Try to do pod-failure action on all filtered pods", p.logPrefix())
	case v1alpha1.FixedPodMode:
		glog.Infof("%s, Try to do pod-failure action on %s pods", p.logPrefix(), p.podChaos.Spec.Value)
	case v1alpha1.FixedPercentPodMode:
		glog.Infof("%s, Try to do pod-failure action on %s%% pods", p.logPrefix(), p.podChaos.Spec.Value)
	case v1alpha1.RandomMaxPercentPodMode:
		glog.Infof("%s, Try to do pod-failure action on max %s%% pods", p.logPrefix(), p.podChaos.Spec.Value)
	default:
		err = fmt.Errorf("pod-failure mode %s not supported", p.podChaos.Spec.Mode)
	}

	if err != nil {
		glog.Errorf("%s, fail to run action, %v", p.logPrefix(), err)
	}
}

// Equal returns true when the two jobs have same PodChaos.
// It can be used to judge if the job need to update this job.
func (p *PodFailureJob) Equal(job manager.Job) bool {
	pjob, ok := job.(*PodFailureJob)
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

// Close close the pod-failure job and cleans the residue actions.
// It will check the finalizers of the PodChaos and cleans them.
func (p *PodFailureJob) Close() error {
	if p.cancel != nil {
		(*p.cancel)()
	}

	if p.wg != nil {
		p.wg.Wait()
	}

	return p.cleanFinalizersAndRecover()
}

func (p *PodFailureJob) failRandomPod(ctx context.Context, pods []v1.Pod) error {
	if len(pods) == 0 {
		return nil
	}

	if p.wg != nil {
		p.wg.Add(1)

		defer p.wg.Done()
	}

	index := rand.Intn(len(pods))
	pod := pods[index]

	duration, err := time.ParseDuration(p.podChaos.Spec.Duration)
	if err != nil {
		return err
	}

	if err := p.addPodFinalizer(pod); err != nil {
		return err
	}

	if err := p.failPod(pod); err != nil {
		return err
	}

	util.Sleep(ctx, duration)

	return nil
}

func (p *PodFailureJob) addPodFinalizer(pod v1.Pod) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pc, err := p.cli.PingcapV1alpha1().PodChaoses(p.podChaos.Namespace).Get(p.podChaos.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		key, err := cache.MetaNamespaceKeyFunc(&pod)
		if err != nil {
			return err
		}

		pc.Finalizers = append(pc.Finalizers, key)

		_, err = p.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Update(pc)

		return err
	})
}

func (p *PodFailureJob) cleanFinalizersAndRecover() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		pc, err := p.cli.PingcapV1alpha1().PodChaoses(p.podChaos.Namespace).Get(p.podChaos.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if len(pc.Finalizers) == 0 {
			return nil
		}

		for _, key := range pc.Finalizers {
			ns, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				return err
			}

			pod, err := p.kubeCli.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					return err
				}

				continue
			}

			if err := p.recoverPod(*pod); err != nil {
				return err
			}
		}

		pc.Finalizers = []string{}

		_, err = p.cli.PingcapV1alpha1().PodChaoses(pc.Namespace).Update(pc)

		return err
	})
}

// failPod updates the image of this pod with a non-existing image
// and save the previous image in annotations of this pod for recovery.
func (p *PodFailureJob) failPod(pod v1.Pod) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nPod, err := p.kubeCli.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		// TODO: check the annotations or others in case that this pod is used by other chaos
		for index := range nPod.Spec.Containers {
			originImage := nPod.Spec.Containers[index].Image
			name := nPod.Spec.Containers[index].Name

			key := GenAnnotationKeyForImage(p.podChaos, name)

			if nPod.Annotations == nil {
				nPod.Annotations = make(map[string]string)
			}

			if _, ok := nPod.Annotations[key]; ok {
				return fmt.Errorf("annotation %s exist", key)
			}

			nPod.Annotations[key] = originImage

			nPod.Spec.Containers[index].Image = fakeImage
		}

		_, err = p.kubeCli.CoreV1().Pods(pod.Namespace).Update(nPod)
		return err
	})
}

// recoverPod updates the images of pod with the previous image stored at annotation.
func (p *PodFailureJob) recoverPod(pod v1.Pod) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nPod, err := p.kubeCli.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				return nil
			}

			return err
		}

		for index := range nPod.Spec.Containers {
			name := nPod.Spec.Containers[index].Name
			annotationKey := GenAnnotationKeyForImage(p.podChaos, name)

			if nPod.Annotations == nil {
				nPod.Annotations = make(map[string]string)
			}

			_, ok := nPod.Annotations[annotationKey]
			if !ok {
				continue
			}
		}

		return p.kubeCli.CoreV1().Pods(pod.Namespace).Delete(nPod.Name, &metav1.DeleteOptions{
			GracePeriodSeconds: new(int64),
		})
	})
}

func (p *PodFailureJob) logPrefix() string {
	return fmt.Sprintf("[%s/%s] [action:%s]", p.podChaos.Namespace, p.podChaos.Name, p.podChaos.Spec.Action)
}
