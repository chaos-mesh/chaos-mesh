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

// Detection algorithm adapted from kubernetes-sigs/iptables-wrappers v3
// https://github.com/kubernetes-sigs/iptables-wrappers
// Original files: internal/iptables/detect.go, internal/iptables/rules.go
// Copyright 2023 The Kubernetes Authors — Apache-2.0 License

package chaosdaemon

import (
	"os/exec"
	"regexp"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// iptablesCmd is set by InitIPTablesMode() at daemon startup.
	// Left empty so any use before initialization fails obviously.
	iptablesCmd string

	kubeletChainsRegex = regexp.MustCompile(`(?m)^:(KUBE-IPTABLES-HINT|KUBE-KUBELET-CANARY)`)
)

// detectIPTablesMode detects whether the host is using iptables-nft or
// iptables-legacy by checking for kubelet-created chains in the mangle table.
// It checks nft first (more common on modern distros), then falls back to
// legacy. If neither backend has kubelet chains, it defaults to nft.
//
// We nsenter into PID 1's network namespace to inspect the host's iptables
// rules. The daemon runs with hostPID=true but may not have hostNetwork=true,
// so its own namespace may not have kubelet chains.
func detectIPTablesMode() string {
	// Check nft first — it's more common these days and we can check
	// efficiently by passing -t mangle.
	for _, saveCmd := range []string{"iptables-save", "ip6tables-save"} {
		out, err := exec.Command("nsenter", "-t", "1", "-n", "--", "xtables-nft-multi", saveCmd, "-t", "mangle").Output()
		if err == nil && kubeletChainsRegex.Match(out) {
			return "nft"
		}
	}

	// Check legacy. We can't pass "-t mangle" to iptables-legacy-save because
	// it would cause the kernel to create that table if it didn't already exist.
	// So we grab all the rules.
	for _, saveCmd := range []string{"iptables-save", "ip6tables-save"} {
		out, err := exec.Command("nsenter", "-t", "1", "-n", "--", "xtables-legacy-multi", saveCmd).Output()
		if err == nil && kubeletChainsRegex.Match(out) {
			return "legacy"
		}
	}

	return "nft"
}

// InitIPTablesMode detects the host's iptables backend and sets iptablesCmd
// to the correct binary name (e.g. "iptables-nft" or "iptables-legacy").
// Must be called before any iptables operations.
func InitIPTablesMode() {
	log := ctrl.Log.WithName("iptables-mode")
	mode := detectIPTablesMode()
	iptablesCmd = "iptables-" + mode
	log.Info("detected iptables backend", "mode", mode, "iptablesCmd", iptablesCmd)
}
