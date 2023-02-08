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
	"io"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseHttpPatchThenRecover(
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

	body := `{"msg":"Hello","target":"World"}`
	secret := "Bar"

	By("waiting for assertion normal behaviour")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == secret && string(b) == body {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "helper server doesn't work as expected")
	By("deploy helper server successfully")

	By("create http patch chaos CRD objects")

	patchbody := `{"target": "Chaos Mesh"}`
	patchSecret := "Foo!"
	expectedSecret := "Bar; Foo!"
	expectedBody := `{"msg":"Hello","target":"Chaos Mesh"}`

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
				Patch: &v1alpha1.PodHttpChaosPatchActions{
					Headers: [][]string{
						{SECRET, patchSecret},
					},
					Body: &v1alpha1.PodHttpChaosPatchBodyAction{
						Type:  "JSON",
						Value: patchbody,
					},
				},
			},
		},
	}
	err = cli.Create(ctx, httpChaos)
	framework.ExpectNoError(err, "create http chaos error")

	By("waiting for assertion HTTP patch")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == expectedSecret && string(b) == expectedBody {
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
		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == secret && string(b) == body {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover http chaos")
}

func TestcaseHttpPatchPauseAndUnPause(
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

	body := `{"msg":"Hello","target":"World"}`
	secret := "Bar"

	By("waiting for assertion normal behaviour")
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == secret && string(b) == body {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "helper server doesn't work as expected")
	By("deploy helper server successfully")

	By("create http patch chaos CRD objects")
	patchbody := `{"target": "Chaos Mesh"}`
	patchSecret := "Foo!"
	expectedSecret := "Bar; Foo!"
	expectedBody := `{"msg":"Hello","target":"Chaos Mesh"}`

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
				Patch: &v1alpha1.PodHttpChaosPatchActions{
					Headers: [][]string{
						{SECRET, patchSecret},
					},
					Body: &v1alpha1.PodHttpChaosPatchBodyAction{
						Type:  "JSON",
						Value: patchbody,
					},
				},
			},
		},
	}
	err = cli.Create(ctx, httpChaos)
	framework.ExpectNoError(err, "create http chaos error")

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

		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == expectedSecret && string(b) == expectedBody {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "http chaos doesn't work as expected")

	By("pause http patch chaos experiment")
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

	// wait 1 min to check whether io patch still exists
	err = wait.PollImmediate(1*time.Second, 1*time.Minute, func() (bool, error) {
		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == secret && string(b) == body {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "fail to recover http chaos")

	By("resume http patch chaos experiment")
	// resume experiment
	err = util.UnPauseChaos(ctx, cli, httpChaos)
	framework.ExpectNoError(err, "resume chaos error")

	By("assert that http patch is effective again")
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
		resp, err := getPodHttp(c, port, secret, body)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		s := strings.Join(resp.Header[SECRET], "; ")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		klog.Infof("Status(%d), Secret(%s), Body(%s)", resp.StatusCode, s, string(b))

		if s == expectedSecret && string(b) == expectedBody {
			return true, nil
		}
		return false, nil
	})
	framework.ExpectNoError(err, "HTTP chaos doesn't work as expected")

	By("cleanup")
	// cleanup
	cli.Delete(ctx, httpChaos)
}
