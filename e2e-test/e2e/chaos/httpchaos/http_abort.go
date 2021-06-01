// Copyright 2021 Chaos Mesh Authors.
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

package httpchaos

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseHttpAbortThenRecover(
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
	By("create http abort chaos CRD objects")

	abort := true

	httpChaos := &v1alpha1.HTTPChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.HTTPChaosSpec{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					Namespaces:     []string{ns},
					LabelSelectors: map[string]string{"app": "http"},
				},
				Mode: v1alpha1.OnePodMode,
			},
			Port:   8080,
			Target: "Request",
			PodHttpChaosActions: v1alpha1.PodHttpChaosActions{
				Abort: &abort,
			},
		},
	}
	err = cli.Create(ctx, httpChaos)
	framework.ExpectNoError(err, "create http chaos error")

	By("waiting for assertion HTTP abort")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		_, err := getPodHttpNoBody(c, port)

		// abort applied
		if err != nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "http chaos doesn't work as expected")
	By("apply http chaos successfully")

	By("delete chaos CRD objects")
	// delete chaos CRD
	err = cli.Delete(ctx, httpChaos)
	framework.ExpectNoError(err, "failed to delete http chaos")

	By("waiting for assertion recovering")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		_, err := getPodHttpNoBody(c, port)

		// abort recovered
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectNoError(err, "fail to recover http chaos")
}

func TestcaseHttpAbortPauseAndUnPause(
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
	By("create http abort chaos CRD objects")

	abort := true

	httpChaos := &v1alpha1.HTTPChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "http-chaos",
			Namespace: ns,
		},
		Spec: v1alpha1.HTTPChaosSpec{
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					Namespaces:     []string{ns},
					LabelSelectors: map[string]string{"app": "http"},
				},
				Mode: v1alpha1.OnePodMode,
			},
			Port:   8080,
			Target: "Request",
			PodHttpChaosActions: v1alpha1.PodHttpChaosActions{
				Abort: &abort,
			},
		},
	}

	err = cli.Create(ctx, httpChaos)
	framework.ExpectNoError(err, "error occurs while applying http chaos")

	chaosKey := types.NamespacedName{
		Namespace: ns,
		Name:      "http-chaos",
	}

	By("waiting for assertion http chaos")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		chaos := &v1alpha1.HTTPChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get http chaos error")

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

		_, err := getPodHttpNoBody(c, port)

		// abort applied
		if err != nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "http chaos doesn't work as expected")

	By("pause http abort chaos experiment")
	// pause experiment
	err = util.PauseChaos(ctx, cli, httpChaos)
	framework.ExpectNoError(err, "pause chaos error")

	By("waiting for assertion about pause")
	err = wait.Poll(1*time.Second, 1*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.HTTPChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get http chaos error")

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
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		_, err := getPodHttpNoBody(c, port)

		// abort recovered
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectNoError(err, "fail to recover http chaos")

	By("resume http abort chaos experiment")
	// resume experiment
	err = util.UnPauseChaos(ctx, cli, httpChaos)
	framework.ExpectNoError(err, "resume chaos error")

	By("assert that http abort is effective again")
	err = wait.Poll(1*time.Second, 1*time.Minute, func() (done bool, err error) {
		chaos := &v1alpha1.HTTPChaos{}
		err = cli.Get(ctx, chaosKey, chaos)
		framework.ExpectNoError(err, "get http chaos error")

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

	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		_, err := getPodHttpNoBody(c, port)

		// abort applied
		if err != nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "HTTP chaos doesn't work as expected")

	By("cleanup")
	// cleanup
	cli.Delete(ctx, httpChaos)
}
