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

package dnschaos

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
)

func TestcaseDNSStatic(ns string, cli client.Client, targetPort, controlPort uint16, c http.Client) {
	ctx := context.Background()
	for _, port := range []uint16{targetPort, controlPort} {
		framework.ExpectNoError(util.WaitE2EHelperReady(c, port), "wait for DNS helper")
	}
	suffix := fmt.Sprintf(".%s.svc.cluster.local", ns)
	serviceDomain := "network-peer" + suffix
	ipv6Domain := "static-ipv6" + suffix
	normalDomain := "dns-control" + suffix
	excludedDomain := "kubernetes.default.svc.cluster.local"
	serviceIP, err := testDNSServer(c, targetPort, serviceDomain)
	framework.ExpectNoError(err, "resolve service before injection")
	normalIP, err := testDNSServer(c, targetPort, normalDomain)
	framework.ExpectNoError(err, "resolve unmapped service before injection")
	excludedIP, err := testDNSServer(c, targetPort, excludedDomain)
	framework.ExpectNoError(err, "resolve pattern-excluded service before injection")
	for _, port := range []uint16{targetPort, controlPort} {
		_, err := testDNSServer(c, port, ipv6Domain)
		gomega.Expect(err).To(gomega.HaveOccurred(), "unmapped IPv6 test name must not resolve")
	}

	experiment := &v1alpha1.DNSChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "dns-chaos-static", Namespace: ns},
		Spec: v1alpha1.DNSChaosSpec{
			Action:             v1alpha1.StaticAction,
			DomainNamePatterns: []string{"*" + suffix},
			DomainIpMappingList: []string{
				strings.ToUpper(serviceDomain) + ".=192.0.2.10",
				ipv6Domain + "=2001:db8::10",
				excludedDomain + "=192.0.2.20",
			},
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode: v1alpha1.AllMode,
					Selector: v1alpha1.PodSelectorSpec{
						GenericSelectorSpec: v1alpha1.GenericSelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "network-peer"},
						},
					},
				},
			},
		},
	}
	framework.ExpectNoError(cli.Create(ctx, experiment), "create static DNSChaos")
	ginkgo.DeferCleanup(func() {
		framework.ExpectNoError(client.IgnoreNotFound(cli.Delete(ctx, experiment)), "clean up static DNSChaos")
	})

	assertIP := func(port uint16, domain, expected string) {
		gomega.Eventually(func() (string, error) {
			return testDNSServer(c, port, domain)
		}, time.Minute, time.Second).Should(gomega.Equal(expected), "DNS answer for %s from helper %d", domain, port)
	}
	assertIP(targetPort, serviceDomain, "192.0.2.10")
	assertIP(targetPort, ipv6Domain, "2001:db8::10")
	assertIP(controlPort, serviceDomain, serviceIP)
	_, err = testDNSServer(c, controlPort, ipv6Domain)
	gomega.Expect(err).To(gomega.HaveOccurred(), "static IPv6 answer must not reach the control pod")
	for _, port := range []uint16{targetPort, controlPort} {
		assertIP(port, normalDomain, normalIP)
		assertIP(port, excludedDomain, excludedIP)
	}

	framework.ExpectNoError(cli.Delete(ctx, experiment), "delete static DNSChaos")
	assertIP(targetPort, serviceDomain, serviceIP)
	gomega.Eventually(func() error {
		_, err := testDNSServer(c, targetPort, ipv6Domain)
		return err
	}, time.Minute, time.Second).Should(gomega.HaveOccurred(), "static IPv6 answer must disappear after recovery")
	assertIP(targetPort, normalDomain, normalIP)
	assertIP(controlPort, serviceDomain, serviceIP)
}
