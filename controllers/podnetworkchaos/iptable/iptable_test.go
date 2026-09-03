// Copyright 2026 Chaos Mesh Authors.
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

package iptable

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

func TestGenerateName(t *testing.T) {
	seen := map[string]bool{}
	for _, namespace := range []string{"team-a", "team-b"} {
		for _, name := range []string{"a", "isolate", strings.Repeat("a", 253)} {
			chaos := &v1alpha1.NetworkChaos{ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      name,
			}}
			for _, direction := range []pb.Chain_Direction{pb.Chain_INPUT, pb.Chain_OUTPUT} {
				chain := GenerateName(direction, chaos)
				if len(chain) > 27 || !strings.HasPrefix(chain, direction.String()+"/") {
					t.Fatalf("invalid chain name %q for %s", chain, direction)
				}
				if seen[chain] {
					t.Fatalf("duplicate chain name %q for %s/%s", chain, namespace, name)
				}
				seen[chain] = true
				recreated := chaos.DeepCopy()
				recreated.UID = "replacement-uid"
				if GenerateName(direction, recreated) != chain {
					t.Fatal("chain name changed for the same namespace/name")
				}
			}
		}
	}
}
