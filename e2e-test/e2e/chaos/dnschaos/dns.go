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

package networkchaos

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseDNSRandom(
	ns string,
	cli client.Client,
	port uint16,
	c http.Client,
) {
	ctx, cancel := context.WithCancel(context.Background())

	err := util.WaitE2EHelperReady(c, port)

	effectDomainNames := []string{"not-exist-host.abc", "not_exist_host.abc", "not-exist-host.def"}

	framework.ExpectNoError(err, "wait e2e helper ready error")

	// get IP of a non exists host, and will get error
	for _, domainName := range effectDomainNames {
		_, err = testDNSServer(c, port, domainName)
		framework.ExpectError(err, "test DNS server failed")
	}

	dnsChaos := &v1alpha1.DNSChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dns-chaos-random",
			Namespace: ns,
		},
		Spec: v1alpha1.DNSChaosSpec{
			Action:             v1alpha1.RandomAction,
			DomainNamePatterns: []string{"not-exist-?ost.*", "not_exist?host.abc", "not-exist-host.def"},
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode: v1alpha1.AllPodMode,
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "network-peer"},
					},
				},
			},
		},
	}

	err = cli.Create(ctx, dnsChaos.DeepCopy())
	framework.ExpectNoError(err, "create dns chaos error")

	for _, domainName := range effectDomainNames {
		err = wait.Poll(time.Second, 5*time.Second, func() (done bool, err error) {
			// get IP of a non exists host, because chaos DNS server will return a random IP,
			// so err should be nil
			_, dnsErr := testDNSServer(c, port, domainName)
			if dnsErr != nil {
				return false, nil
			}
			return true, nil
		})
		framework.ExpectNoError(err, "test DNS server failed")
	}

	err = cli.Delete(ctx, dnsChaos.DeepCopy())
	framework.ExpectNoError(err, "failed to delete dns chaos")

	cancel()
}

func TestcaseDNSError(
	ns string,
	cli client.Client,
	port uint16,
	c http.Client,
) {
	ctx, cancel := context.WithCancel(context.Background())

	err := util.WaitE2EHelperReady(c, port)

	framework.ExpectNoError(err, "wait e2e helper ready error")

	effectDomainNames := []string{"chaos-mesh.org", "github.com", "163.com"}

	// get IP of chaos-mesh.org, and will get no error
	for _, domainName := range effectDomainNames {
		_, err = testDNSServer(c, port, domainName)
		framework.ExpectNoError(err, "test DNS server failed")
	}

	dnsChaos := &v1alpha1.DNSChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dns-chaos-error",
			Namespace: ns,
		},
		Spec: v1alpha1.DNSChaosSpec{
			Action:             v1alpha1.ErrorAction,
			DomainNamePatterns: []string{"chaos-mes?.org", "github.com", "16?.co*"},
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode: v1alpha1.AllPodMode,
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "network-peer"},
					},
				},
			},
		},
	}

	err = cli.Create(ctx, dnsChaos.DeepCopy())
	framework.ExpectNoError(err, "create dns chaos error")

	for _, domainName := range effectDomainNames {
		err = wait.Poll(time.Second, 5*time.Second, func() (done bool, err error) {
			// get IP of a chaos-mesh.org, because chaos DNS server will return error,
			// so err should not be nil
			_, dnsErr := testDNSServer(c, port, domainName)
			if dnsErr == nil {
				return false, nil
			}
			return true, nil
		})
		framework.ExpectNoError(err, "test DNS server failed")
	}

	err = cli.Delete(ctx, dnsChaos.DeepCopy())
	framework.ExpectNoError(err, "failed to delete dns chaos")

	cancel()
}

func testDNSServer(c http.Client, port uint16, url string) (string, error) {
	klog.Infof("sending request to http://localhost:%d/dns?url=%s", port, url)

	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/dns?url=%s", port, url))
	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	result := string(out)
	klog.Infof("testDNSServer result: %s", result)
	if strings.Contains(result, "failed") {
		return "", fmt.Errorf("test DNS server failed")
	}

	return result, nil
}
