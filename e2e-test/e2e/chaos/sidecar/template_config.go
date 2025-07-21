// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package sidecar

import (
	"context"
	"time"

	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
)

func TestcaseInvalidConfigMapKey(
	ns string,
	cmNamespace string,
	cmName string,
	kubeCli kubernetes.Interface,
	cli client.Client,
) {

	ctx, cancel := context.WithCancel(context.Background())
	err := createTemplateConfig(ctx, cli, cmName,
		map[string]string{
			"chaos-pd.yaml": `name: chaosfs-pd
selector:
  labelSelectors:
    "app.kubernetes.io/component": "pd"`})
	framework.ExpectNoError(err, "failed to create template config")

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/component": "controller-manager",
		}).String(),
	}
	pods, err := kubeCli.CoreV1().Pods(cmNamespace).List(context.TODO(), listOptions)
	framework.ExpectNoError(err, "get chaos mesh controller pods error")

	err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		newPods, err := kubeCli.CoreV1().Pods(cmNamespace).List(context.TODO(), listOptions)
		framework.ExpectNoError(err, "get chaos mesh controller pods error")
		if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
			return true, nil
		}
		if len(newPods.Items) > 0 && newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
			return true, nil
		}
		return false, nil
	})
	gomega.Expect(err).Should(gomega.HaveOccurred(), "wait chaos mesh not dies")
	gomega.Expect(err).To(gomega.MatchError(wait.ErrWaitTimeout))

	cancel()
}

func TestcaseInvalidConfiguration(
	ns string,
	cmNamespace string,
	cmName string,
	kubeCli kubernetes.Interface,
	cli client.Client,
) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := createTemplateConfig(ctx, cli, cmName,
		map[string]string{
			"data": `name: chaosfs-pd
selector:
  labelSelectors:
    "app.kubernetes.io/component": "pd"`})
	framework.ExpectNoError(err, "failed to create template config")

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/component": "controller-manager",
		}).String(),
	}
	pods, err := kubeCli.CoreV1().Pods(cmNamespace).List(context.TODO(), listOptions)
	framework.ExpectNoError(err, "get chaos mesh controller pods error")

	err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
		newPods, err := kubeCli.CoreV1().Pods(cmNamespace).List(context.TODO(), listOptions)
		framework.ExpectNoError(err, "get chaos mesh controller pods error")
		if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
			return true, nil
		}
		if len(newPods.Items) > 0 && newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
			return true, nil
		}
		return false, nil
	})
	gomega.Expect(err).Should(gomega.HaveOccurred(), "wait chaos mesh not dies")
	gomega.Expect(err).To(gomega.MatchError(wait.ErrWaitTimeout))
}
