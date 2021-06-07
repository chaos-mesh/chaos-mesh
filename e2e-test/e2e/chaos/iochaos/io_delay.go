// Copyright 2020 Chaos Mesh Authors.
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

package iochaos

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseIODelayDurationForATimeThenRecover(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("waiting on e2e helper ready")
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")
	By("create IO delay chaos CRD objects")
	ioChaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "io-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.IOChaosSpec{
			Action:     v1alpha1.IoLatency,
			VolumePath: "/var/run/data",
			Path:       "/var/run/data/*",
			Delay:      "1s",
			Percent:    100,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "io"},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, ioChaos)
	framework.ExpectNoError(err, "create io chaos error")
	By("waiting for assertion IO delay")
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodIODelay(c, port)
		second := dur.Seconds()
		klog.Infof("get io delay %fs", second)
		// IO Delay >= 1s
		if second >= 1 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")
	By("apply io chaos successfully")

	By("delete chaos CRD objects")
	// delete chaos CRD
	err = cli.Delete(ctx, ioChaos)
	framework.ExpectNoError(err, "failed to delete io chaos")
	By("waiting for assertion recovering")
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodIODelay(c, port)
		second := dur.Seconds()
		klog.Infof("get io delay %fs", second)
		// IO Delay shouldn't longer than 1s
		if second >= 1 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectNoError(err, "fail to recover io chaos")
}

func TestcaseIODelayDurationForATimePauseAndUnPause(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	By("waiting for e2e helper ready")
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	By("create io chaos crd object")
	ioChaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "io-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.IOChaosSpec{
			Action:     v1alpha1.IoLatency,
			VolumePath: "/var/run/data",
			Path:       "/var/run/data/*",
			Delay:      "10ms",
			Percent:    100,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "io"},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, ioChaos)
	framework.ExpectNoError(err, "error occurs while applying io chaos")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "io-chaos",
	}

	By("waiting for assertion io chaos")
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		chaos := &v1alpha1.IOChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get io chaos error")

		for _, c := range chaos.GetStatus().Conditions {
			if c.Type == v1alpha1.ConditionAllInjected {
				if c.Status != corev1.ConditionTrue {
					return false, nil
				}
			} else if c.Type == v1alpha1.ConditionSelected {
				if c.Status != corev1.ConditionTrue {
					return false, nil
				}
			}
		}

		dur, _ := getPodIODelay(c, port)

		ms := dur.Milliseconds()
		klog.Infof("get io delay %dms", ms)
		// IO Delay >= 500ms
		if ms >= 10 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	By("pause io delay chaos experiment")
	// pause experiment
	err = util.PauseChaos(ctx, cli, ioChaos)
	framework.ExpectNoError(err, "pause chaos error")

	By("waiting for assertion about pause")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.IOChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get io chaos error")

		for _, c := range chaos.GetStatus().Conditions {
			if c.Type == v1alpha1.ConditionAllRecovered {
				if c.Status != corev1.ConditionTrue {
					return false, nil
				}
			} else if c.Type == v1alpha1.ConditionSelected {
				if c.Status != corev1.ConditionTrue {
					return false, nil
				}
			}
		}

		return true, err
	})
	framework.ExpectNoError(err, "check paused chaos failed")

	// wait 1 min to check whether io delay still exists
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodIODelay(c, port)

		ms := dur.Milliseconds()
		klog.Infof("get io delay %dms", ms)
		// IO Delay shouldn't longer than 10ms
		if ms > 10 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectNoError(err, "fail to recover io chaos")

	By("resume io delay chaos experiment")
	// resume experiment
	err = util.UnPauseChaos(ctx, cli, ioChaos)
	framework.ExpectNoError(err, "resume chaos error")

	By("assert that io delay is effective again")
	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.IOChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get io chaos error")

		for _, c := range chaos.GetStatus().Conditions {
			if c.Type == v1alpha1.ConditionAllInjected {
				if c.Status != corev1.ConditionTrue {
					return false, nil
				}
			} else if c.Type == v1alpha1.ConditionSelected {
				if c.Status != corev1.ConditionTrue {
					return false, nil
				}
			}
		}

		return true, err
	})
	framework.ExpectNoError(err, "check resumed chaos failed")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodIODelay(c, port)

		ms := dur.Milliseconds()
		klog.Infof("get io delay %dms", ms)
		// IO Delay >= 10ms
		if ms >= 10 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	By("cleanup")
	// cleanup
	cli.Delete(ctx, ioChaos)
}

func TestcaseIODelayWithSpecifiedContainer(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	containerName := "io"
	ioChaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "io-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.IOChaosSpec{
			Action:     v1alpha1.IoLatency,
			VolumePath: "/var/run/data",
			Path:       "/var/run/data/*",
			Delay:      "10ms",
			Percent:    100,
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "io"},
					},
					Mode: v1alpha1.OnePodMode,
				},
				ContainerNames: []string{containerName},
			},
		},
	}
	err = cli.Create(ctx, ioChaos)
	framework.ExpectNoError(err, "create io chaos error")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodIODelay(c, port)

		ms := dur.Milliseconds()
		klog.Infof("get io delay %dms", ms)
		// IO Delay >= 10ms
		if ms >= 10 {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")
	klog.Infof("apply io chaos successfully")

	err = cli.Delete(ctx, ioChaos)
	framework.ExpectNoError(err, "failed to delete io chaos")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodIODelay(c, port)

		ms := dur.Milliseconds()
		klog.Infof("get io delay %dms", ms)
		// IO Delay shouldn't longer than 10ms
		if ms >= 10 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectNoError(err, "fail to recover io chaos")
}

func TestcaseIODelayWithWrongSpec(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("waiting on e2e helper ready")
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")
	By("create IO delay chaos CRD objects")
	ioChaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "io-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.IOChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "io"},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
			Action:     v1alpha1.IoLatency,
			VolumePath: "/var/run/data/123",
			Path:       "/var/run/data/*",
			Delay:      "1s",
			Percent:    100,
			Duration:   pointer.StringPtr("9m"),
		},
	}
	err = cli.Create(ctx, ioChaos)
	framework.ExpectNoError(err, "create io chaos error")
	// TODO: resume the e2e test
	// err = wait.PollImmediate(5*time.Second, 30*time.Second, func() (bool, error) {
	// 	err := cli.Get(ctx, types.NamespacedName{Namespace: ioChaos.ObjectMeta.Namespace, Name: ioChaos.ObjectMeta.Name}, ioChaos)
	// 	if err != nil {
	// 		return false, err
	// 	}
	// 	errStr := ioChaos.Status.ChaosStatus.FailedMessage
	// 	klog.Infof("get chaos err: %s", errStr)
	// 	if strings.Contains(errStr, "Toda startup takes too long or an error occurs") {
	// 		return true, nil
	// 	}
	// 	return false, nil
	// })
	// framework.ExpectNoError(err, "A wrong chaos spec should raise an error")
}
