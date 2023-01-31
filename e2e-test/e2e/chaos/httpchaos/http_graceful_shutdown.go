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

package httpchaos

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	test "github.com/chaos-mesh/chaos-mesh/e2e-test"
	e2econfig "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseHttpGracefulAbortShutdown(
	ns string,
	cli client.Client,
	c HTTPE2EClient,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	By("waiting on e2e helper ready")
	err := util.WaitHTTPE2EHelperReady(*c.C, c.IP, port)
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
					GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "http"},
					},
				},
				Mode: v1alpha1.OneMode,
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

	defer func() {
		err = cli.Delete(ctx, httpChaos)
		framework.ExpectNoError(err, "delete http chaos")
	}()

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

	By("upgrade chaos mesh")
	// Get clients
	oa, ocfg, err := test.BuildOperatorActionAndCfg(e2econfig.TestConfig)
	framework.ExpectNoError(err, "failed to create operator action")
	err = oa.RestartDaemon(ocfg)
	framework.ExpectNoError(err, "failed to restart chaos daemon")

	By("waiting for assertion chaos recovered")
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		_, err := getPodHttpNoBody(c, port)

		// abort recovered
		if err == nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "http chaos doesn't gracefully shutdown as expected")
	By("http chaos shutdown successfully")
}
