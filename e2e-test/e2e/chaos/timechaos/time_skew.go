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

package timechaos

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseTimeSkewOnceThenRecover(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("wait e2e helper ready")
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	By("create chaos CRD objects")
	initTime, err := getPodTimeNS(c, port)
	framework.ExpectNoError(err, "failed to get pod time")

	timeChaos := &v1alpha1.TimeChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "timer-time-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.TimeChaosSpec{
			Duration:   pointer.StringPtr("9m"),
			TimeOffset: "-1h",
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "timer"},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, timeChaos)
	framework.ExpectNoError(err, "create time chaos error")

	By("waiting for assertion")
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		podTime, err := getPodTimeNS(c, port)
		framework.ExpectNoError(err, "failed to get pod time")
		if podTime.Before(*initTime) {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "time chaos doesn't work as expected")

	By("delete chaos CRD objects")
	err = cli.Delete(ctx, timeChaos)
	framework.ExpectNoError(err, "failed to delete time chaos")

	By("waiting for assertion recovering")
	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		podTime, err := getPodTimeNS(c, port)
		framework.ExpectNoError(err, "failed to get pod time")
		// since there is no timechaos now, current pod time should not be earlier
		// than the init time
		if podTime.Before(*initTime) {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectError(err, "wait no timechaos error")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())
	By("success to perform time chaos")
}

func TestcaseTimeSkewPauseThenUnpause(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("wait e2e helper ready")
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	initTime, err := getPodTimeNS(c, port)
	framework.ExpectNoError(err, "failed to get pod time")

	By("create chaos CRD objects")
	timeChaos := &v1alpha1.TimeChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "timer-time-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.TimeChaosSpec{
			Duration:   pointer.StringPtr("9m"),
			TimeOffset: "-1h",
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "timer"},
					},
					Mode: v1alpha1.OnePodMode,
				},
			},
		},
	}
	err = cli.Create(ctx, timeChaos)
	framework.ExpectNoError(err, "create time chaos error")

	By("waiting for assertion")
	err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		podTime, err := getPodTimeNS(c, port)
		framework.ExpectNoError(err, "failed to get pod time")
		if podTime.Before(*initTime) {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "time chaos doesn't work as expected")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "timer-time-chaos",
	}

	By("pause time skew chaos experiment")
	// pause experiment
	err = util.PauseChaos(ctx, cli, timeChaos)
	framework.ExpectNoError(err, "pause chaos error")

	By("assert pause is effective")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.TimeChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get time chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check paused chaos failed")

	// wait for 1 minutes and check timer
	framework.ExpectNoError(err, "get timer pod error")
	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		podTime, err := getPodTimeNS(c, port)
		framework.ExpectNoError(err, "failed to get pod time")
		if podTime.Before(*initTime) {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectError(err, "wait time chaos paused error")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	By("resume time skew chaos experiment")
	err = util.UnPauseChaos(ctx, cli, timeChaos)
	framework.ExpectNoError(err, "resume chaos error")

	By("assert chaos experiment resumed")
	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.TimeChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get time chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.RunningPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check resumed chaos failed")

	// timechaos is running again, we want to check pod
	// whether time is earlier than init time,
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		podTime, err := getPodTimeNS(c, port)
		framework.ExpectNoError(err, "failed to get pod time")
		if podTime.Before(*initTime) {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "time chaos failed")

	By("delete chaos CRD objects")
	cli.Delete(ctx, timeChaos)
}
