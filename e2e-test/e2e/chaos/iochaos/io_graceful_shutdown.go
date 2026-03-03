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

package iochaos

import (
	"context"
	"net/http"
	"strings"
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

func TestcaseIOErrorGracefulShutdown(
	ns string,
	cli client.Client,
	c http.Client,
	port uint16,
) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "io"},
						},
					},
					Mode: v1alpha1.OneMode,
				},
			},
		},
	}
	err = cli.Create(ctx, ioChaos)
	framework.ExpectNoError(err, "create io chaos")

	defer func() {
		err = cli.Delete(ctx, ioChaos)
		framework.ExpectNoError(err, "delete io chaos")
	}()

	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)
		// input/output error is errno 5
		if err != nil && strings.Contains(err.Error(), "input/output error") {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't work as expected")

	By("upgrade chaos mesh")
	// Get clients
	oa, ocfg, err := test.BuildOperatorActionAndCfg(e2econfig.TestConfig)
	framework.ExpectNoError(err, "failed to create operator action")
	err = oa.RestartDaemon(ocfg)
	framework.ExpectNoError(err, "failed to restart chaos daemon")

	By("waiting for assertion IO error recovery")
	err = wait.PollImmediate(5*time.Second, 2*time.Minute, func() (bool, error) {
		_, err = getPodIODelay(c, port)
		// recovered
		if err == nil {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "io chaos doesn't gracefully shutdown as expected")
	By("io chaos shutdown successfully")
}
