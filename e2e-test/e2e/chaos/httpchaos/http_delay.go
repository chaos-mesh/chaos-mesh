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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseHttpDelayDurationForATimeThenRecover(
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
	By("create http delay chaos CRD objects")
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
			Delay: time.Second,
		},
	}
	err = cli.Create(ctx, httpChaos)
	framework.ExpectNoError(err, "create http chaos error")
	By("waiting for assertion HTTP delay")
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodHttpDelay(c, port)
		second := dur.Seconds()
		klog.Infof("get http delay %fs", second)
		// IO Delay >= 1s
		if second >= 1 {
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
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		dur, _ := getPodHttpDelay(c, port)
		second := dur.Seconds()
		klog.Infof("get http delay %fs", second)
		// IO Delay shouldn't longer than 1s
		if second >= 1 {
			return false, nil
		}
		return true, nil
	})
	framework.ExpectNoError(err, "fail to recover http chaos")
}
