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
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseIOErrorDurationForATimeThenRecover(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	ioChaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "io-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.IOChaosSpec{
			Action:     v1alpha1.IoFaults,
			VolumePath: "/var/run/data",
			Path:       "/var/run/data/*",
			Percent:    100,
			// errno 5 is EIO -> I/O error
			Errno: 5,
			// only inject write method
			Methods: []v1alpha1.IoMethod{v1alpha1.Write},
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
	framework.ExpectNoError(err, "create io chaos")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)
		// input/output error is errno 5
		if err != nil && strings.Contains(err.Error(), "input/output error") {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	err = cli.Delete(ctx, ioChaos)
	framework.ExpectNoError(err, "failed to delete io chaos")

	klog.Infof("success to perform io chaos")
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)

		if err == nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover io chaos")

	cancel()
}

func TestcaseIOErrorDurationForATimePauseAndUnPause(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	err := util.WaitE2EHelperReady(c, port)
	framework.ExpectNoError(err, "wait e2e helper ready error")

	ioChaos := &v1alpha1.IOChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "io-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.IOChaosSpec{
			Action:     v1alpha1.IoFaults,
			VolumePath: "/var/run/data",
			Path:       "/var/run/data/*",
			Percent:    100,
			// errno 5 is EIO -> I/O error
			Errno: 5,
			// only inject write method
			Methods: []v1alpha1.IoMethod{v1alpha1.Write},
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

	klog.Info("create iochaos successfully")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)
		// input/output error is errno 5
		if err != nil && strings.Contains(err.Error(), "input/output error") {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "io-chaos",
	}

	// pause experiment
	err = util.PauseChaos(ctx, cli, ioChaos)
	framework.ExpectNoError(err, "pause chaos error")

	klog.Info("pause iochaos")

	err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.IOChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get io chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check paused chaos failed")

	// wait 1 min to check whether io delay still exists
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)

		if err == nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover io chaos")

	// resume experiment
	err = util.UnPauseChaos(ctx, cli, ioChaos)
	framework.ExpectNoError(err, "resume chaos error")

	err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.IOChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get io chaos error")
		if chaos.Status.Experiment.DesiredPhase == v1alpha1.RunningPhase {
			return true, nil
		}
		return false, err
	})
	framework.ExpectNoError(err, "check resumed chaos failed")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)
		// input/output error is errno 5
		if err != nil && strings.Contains(err.Error(), "input/output error") {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	// cleanup
	cli.Delete(ctx, ioChaos)
	cancel()
}

func TestcaseIOErrorWithSpecifiedContainer(
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
			Action:     v1alpha1.IoFaults,
			VolumePath: "/var/run/data",
			Path:       "/var/run/data/*",
			Percent:    100,
			// errno 5 is EIO -> I/O error
			Errno: 5,
			// only inject write method
			Methods: []v1alpha1.IoMethod{v1alpha1.Write},
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
	framework.ExpectNoError(err, "create io chaos")

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)
		// input/output error is errno 5
		if err != nil && strings.Contains(err.Error(), "input/output error") {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	err = cli.Delete(ctx, ioChaos)
	framework.ExpectNoError(err, "failed to delete io chaos")

	klog.Infof("success to perform io chaos")
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)

		if err == nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover io chaos")
}
