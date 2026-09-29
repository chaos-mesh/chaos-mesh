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

package chaosdaemon

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("iptables mode detection", func() {
	Context("kubeletChainsRegex", func() {
		It("should match KUBE-IPTABLES-HINT chain", func() {
			output := []byte(`*mangle
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
:KUBE-IPTABLES-HINT - [0:0]
:KUBE-KUBELET-CANARY - [0:0]
COMMIT
`)
			Expect(kubeletChainsRegex.Match(output)).To(BeTrue())
		})

		It("should match KUBE-KUBELET-CANARY chain", func() {
			output := []byte(`*mangle
:PREROUTING ACCEPT [0:0]
:KUBE-KUBELET-CANARY - [0:0]
COMMIT
`)
			Expect(kubeletChainsRegex.Match(output)).To(BeTrue())
		})

		It("should not match when no kubelet chains present", func() {
			output := []byte(`*mangle
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
COMMIT
`)
			Expect(kubeletChainsRegex.Match(output)).To(BeFalse())
		})

		It("should not match empty output", func() {
			Expect(kubeletChainsRegex.Match([]byte{})).To(BeFalse())
		})

		It("should not match chain name appearing in a rule (not a chain definition)", func() {
			output := []byte(`*mangle
:PREROUTING ACCEPT [0:0]
-A PREROUTING -j KUBE-IPTABLES-HINT
COMMIT
`)
			Expect(kubeletChainsRegex.Match(output)).To(BeFalse())
		})
	})

	Context("InitIPTablesMode", func() {
		It("should set iptablesCmd with a mode suffix", func() {
			// On a dev machine without xtables binaries, detectIPTablesMode
			// will fail all exec calls and default to "nft"
			InitIPTablesMode()
			Expect(iptablesCmd).To(SatisfyAny(
				Equal("iptables-nft"),
				Equal("iptables-legacy"),
			))
		})
	})
})
