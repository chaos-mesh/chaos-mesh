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

package ipset

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
)

func Test_generateIPSetName(t *testing.T) {
	g := NewWithT(t)
	postfix := "alongpostfix"

	t.Run("name with postfix", func(_ *testing.T) {
		chaosName := "test"

		networkChaos := &v1alpha1.NetworkChaos{
			Name: chaosName,
		}

		name := GenerateIPSetName(networkChaos, postfix)

		g.Expect(name).Should(Equal(chaosName + "_" + postfix))
	})

	t.Run("length equal 27", func(_ *testing.T) {
		networkChaos := &v1alpha1.NetworkChaos{
			Name: "test-metav1object",
		}

		name := GenerateIPSetName(networkChaos, postfix)

		g.Expect(len(name)).Should(Equal(27))
	})
}

func TestBuildIPSets_IPv6Only(t *testing.T) {
	g := NewWithT(t)

	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "test-chaos"},
	}

	pods := []corev1.Pod{
		{
			Status: corev1.PodStatus{
				PodIPs: []corev1.PodIP{{IP: "fd00::1"}, {IP: "fd00::2"}},
			},
		},
	}

	v4Sets, v6Sets := BuildIPSets(pods, nil, chaos, "tgt", "test-source")

	// No IPv4 addresses: v4 sets should not be created
	g.Expect(v4Sets).Should(BeNil())

	// IPv6 sets should have 1 entry
	g.Expect(v6Sets).Should(HaveLen(1))
	g.Expect(v6Sets[0].IPSetType).Should(Equal(v1alpha1.NetIPSetV6))
	g.Expect(v6Sets[0].Cidrs).Should(ContainElements("fd00::1/128", "fd00::2/128"))

	// IPv4-only pods yield nil v6 sets
	v4OnlyPods := []corev1.Pod{{Status: corev1.PodStatus{PodIPs: []corev1.PodIP{{IP: "10.0.0.1"}}}}}
	v4Only, nilV6 := BuildIPSets(v4OnlyPods, nil, chaos, "tgt", "test-source")
	g.Expect(v4Only).Should(HaveLen(1))
	g.Expect(nilV6).Should(BeNil())
}

func TestBuildIPSets_PodIPFallback(t *testing.T) {
	g := NewWithT(t)

	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "test-chaos"},
	}

	pods := []corev1.Pod{
		{Status: corev1.PodStatus{PodIP: "10.0.0.1"}},
		{Status: corev1.PodStatus{PodIP: "fd00::1"}},
		{Status: corev1.PodStatus{
			PodIP:  "10.0.0.99",
			PodIPs: []corev1.PodIP{{IP: "10.0.0.2"}, {IP: "fd00::2"}},
		}},
	}

	v4Sets, v6Sets := BuildIPSets(pods, nil, chaos, "tgt", "test-source")
	g.Expect(v4Sets).Should(HaveLen(1))
	g.Expect(v4Sets[0].Cidrs).Should(Equal([]string{"10.0.0.1/32", "10.0.0.2/32"}))
	g.Expect(v6Sets).Should(HaveLen(1))
	g.Expect(v6Sets[0].Cidrs).Should(Equal([]string{"fd00::1/128", "fd00::2/128"}))
}

func TestFamilyFromDNSOnly(t *testing.T) {
	g := NewWithT(t)

	v4Pod := []corev1.Pod{{Status: corev1.PodStatus{PodIP: "10.0.0.1"}}}
	v4Explicit, err := netutils.ResolveCidr("192.0.2.1")
	g.Expect(err).NotTo(HaveOccurred())
	v6Explicit, err := netutils.ResolveCidr("2001:db8::1")
	g.Expect(err).NotTo(HaveOccurred())
	dnsDual := []netutils.Target{
		{Network: v4Explicit[0].Network, FromDNS: true},
		{Network: v6Explicit[0].Network, FromDNS: true},
	}

	v4DNS, v6DNS := FamilyFromDNSOnly(nil, dnsDual)
	g.Expect(v4DNS).To(BeTrue())
	g.Expect(v6DNS).To(BeTrue())

	v4DNS, v6DNS = FamilyFromDNSOnly(v4Pod, dnsDual)
	g.Expect(v4DNS).To(BeFalse())
	g.Expect(v6DNS).To(BeTrue())

	v4DNS, v6DNS = FamilyFromDNSOnly(nil, append(dnsDual, v6Explicit...))
	g.Expect(v4DNS).To(BeTrue())
	g.Expect(v6DNS).To(BeFalse())
}

func TestBuildIPSets_DualStack(t *testing.T) {
	g := NewWithT(t)

	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "test-chaos"},
	}

	pods := []corev1.Pod{
		{
			Status: corev1.PodStatus{
				PodIPs: []corev1.PodIP{{IP: "10.0.0.1"}, {IP: "fd00::1"}},
			},
		},
	}

	v4Sets, v6Sets := BuildIPSets(pods, nil, chaos, "tgt", "test-source")

	// Both families should have entries
	g.Expect(v4Sets[0].Cidrs).Should(ContainElements("10.0.0.1/32"))
	g.Expect(v6Sets[0].Cidrs).Should(ContainElements("fd00::1/128"))
	g.Expect(v4Sets[0].IPSetType).Should(Equal(v1alpha1.NetIPSet))
	g.Expect(v6Sets[0].IPSetType).Should(Equal(v1alpha1.NetIPSetV6))
}

func TestBuildIPSets_IPv4Mapped(t *testing.T) {
	g := NewWithT(t)
	external, err := netutils.ResolveCidr("[::ffff:192.0.2.9]:443")
	g.Expect(err).NotTo(HaveOccurred())
	pods := []corev1.Pod{{Status: corev1.PodStatus{
		PodIPs: []corev1.PodIP{{IP: "::ffff:c000:020a"}},
	}}}

	v4, v6 := BuildIPSets(pods, external, &v1alpha1.NetworkChaos{}, "tgt", "src")
	g.Expect(v6).To(BeEmpty())
	g.Expect(v4).To(HaveLen(2))
	g.Expect(v4[0].IPSetType).To(Equal(v1alpha1.NetIPSet))
	g.Expect(v4[0].Cidrs).To(Equal([]string{"192.0.2.10/32"}))
	g.Expect(v4[1].IPSetType).To(Equal(v1alpha1.NetPortIPSet))
	g.Expect(v4[1].CidrAndPorts).To(Equal([]v1alpha1.CidrAndPort{{Cidr: "192.0.2.9/32", Port: 443}}))
}

func TestBuildSetIPSet_IPv6(t *testing.T) {
	g := NewWithT(t)

	chaos := &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "test-chaos"},
	}

	targets, err := netutils.ResolveCidr("fd00::/64")
	g.Expect(err).NotTo(HaveOccurred())
	_, v6Sets := BuildIPSets(nil, targets, chaos, "tgt", "src")

	v6Agg := BuildSetIPSet(v6Sets, chaos, "6_tgt", "src")
	g.Expect(v6Agg.IPSetType).Should(Equal(v1alpha1.SetIPSet))
	g.Expect(v6Agg.SetNames).Should(HaveLen(1))
}
