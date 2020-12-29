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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util"
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

	timeChaos := createTimeChaos(ns, "9m", "@every 10m")
	err = cli.Create(ctx, timeChaos)
	framework.ExpectNoError(err, "create time chaos error")

	By("waiting for assertion")
	err = waitChaosWorking(initTime, c, port)
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
	timeChaos := createTimeChaos(ns, "9m", "@every 10m")
	err = cli.Create(ctx, timeChaos)
	framework.ExpectNoError(err, "create time chaos error")

	By("waiting for assertion")
	err = waitChaosWorking(initTime, c, port)
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
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhasePaused, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check paused chaos failed")

	// wait for 1 minutes and check timer
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectError(err, "wait time chaos paused error")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	By("resume time skew chaos experiment")
	err = util.UnPauseChaos(ctx, cli, timeChaos)
	framework.ExpectNoError(err, "resume chaos error")

	By("assert chaos experiment resumed")
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhaseRunning, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check resumed chaos failed")

	// timechaos is running again, we want to check pod
	// whether time is earlier than init time,
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectNoError(err, "time chaos failed")

	By("delete chaos CRD objects")
	cli.Delete(ctx, timeChaos)
}

func TestcaseTimeSkewPauseThenAutoResumeAtRunning(
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
	timeChaos := createTimeChaos(ns, "9m", "@every 10m")
	err = cli.Create(ctx, timeChaos)
	framework.ExpectNoError(err, "create time chaos error")

	By("waiting for assertion")
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectNoError(err, "time chaos doesn't work as expected")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "timer-time-chaos",
	}

	By("pause time skew chaos experiment for 2min")
	pauseTime := time.Now()
	// pause experiment
	err = util.PauseChaosForDuration(ctx, cli, timeChaos, "2m")
	framework.ExpectNoError(err, "pause chaos error")

	By("assert pause is effective")
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhasePaused, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check paused chaos failed")

	// wait for 1 minute and check timer
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectError(err, "wait time chaos paused error")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	By("assert chaos experiment resumed")
	time.Sleep(2*time.Minute - time.Now().Sub(pauseTime))
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhaseRunning, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check resumed chaos failed")

	// timechaos is running again, we want to check pod
	// whether time is earlier than init time,
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectNoError(err, "time chaos failed")

	By("delete chaos CRD objects")
	cli.Delete(ctx, timeChaos)
}

func TestcaseTimeSkewPauseThenAutoResumeAtWaiting(
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
	timeChaos := createTimeChaos(ns, "2m", "@every 4m")
	err = cli.Create(ctx, timeChaos)
	framework.ExpectNoError(err, "create time chaos error")
	createTime := time.Now()

	By("waiting for assertion")
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectNoError(err, "time chaos doesn't work as expected")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "timer-time-chaos",
	}

	By("pause time skew chaos experiment for 2min")
	// sleep till pause at the first minute in cron cycle
	var sleepTime time.Duration
	// to make sure resume at waiting state
	if time.Now().Sub(createTime) < time.Minute {
		sleepTime = 0
	} else {
		sleepTime = 4*time.Minute - time.Now().Sub(createTime)
	}
	time.Sleep(sleepTime)
	// pause experiment
	err = util.PauseChaosForDuration(ctx, cli, timeChaos, "2m")
	framework.ExpectNoError(err, "pause chaos error")

	By("assert pause is effective")
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhasePaused, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check paused chaos failed")
	pauseTime := time.Now()

	// wait for 1 minute and check timer
	err = waitChaosWorking(initTime, c, port)
	framework.ExpectError(err, "wait time chaos paused error")
	framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

	By("assert pause stop and still in waiting state")
	time.Sleep(2*time.Minute - time.Now().Sub(pauseTime))
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhaseWaiting, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check paused chaos failed")

	By("assert chaos experiment resumed")
	time.Sleep(4*time.Minute - time.Now().Sub(createTime))
	err = waitChaosStatus(ctx, v1alpha1.ExperimentPhaseRunning, chaosKey, initTime, cli)
	framework.ExpectNoError(err, "check resumed chaos failed")

	// timechaos is running again, we want to check pod
	// whether time is earlier than init time,
	err = waitChaosWorking(initTime, c, port)

	By("delete chaos CRD objects")
	cli.Delete(ctx, timeChaos)
}
