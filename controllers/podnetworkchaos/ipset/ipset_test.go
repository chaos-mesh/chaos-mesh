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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestGenerateIPSetName(t *testing.T) {
	identities := []types.NamespacedName{
		{Namespace: "team-a", Name: "isolate"},
		{Namespace: "team-b", Name: "isolate"},
		{Namespace: "a", Name: "b-c"},
		{Namespace: "a-b", Name: "c"},
		{Namespace: "a", Name: "b"},
		{Namespace: "b", Name: "b"},
		{Namespace: strings.Repeat("n", 63), Name: strings.Repeat("a", 253)},
		{Namespace: strings.Repeat("n", 63), Name: strings.Repeat("a", 252) + "b"},
	}
	seen := map[string]string{}
	for _, identity := range identities {
		chaos := &v1alpha1.NetworkChaos{ObjectMeta: metav1.ObjectMeta{
			Namespace: identity.Namespace,
			Name:      identity.Name,
		}}
		for _, role := range []string{"src", "tgt", "nesrc", "netgt", "basrc", "batgt"} {
			for _, setType := range []string{"net_", "netport_", "set_"} {
				postfix := setType + role
				name := GenerateIPSetName(chaos, postfix)
				if len(name) > 27 || len(name+"old") > 31 {
					t.Fatalf("IP set name %q exceeds the daemon's name limits", name)
				}
				if previous, ok := seen[name]; ok {
					t.Fatalf("IP set name %q is shared by %s and %s/%s", name, previous, identity, postfix)
				}
				seen[name] = identity.String() + "/" + postfix
				// Names follow namespace/name ownership and do not depend on a UID.
				recreated := chaos.DeepCopy()
				recreated.UID = "replacement-uid"
				if GenerateIPSetName(recreated, postfix) != name {
					t.Fatal("IP set name changed for the same namespace/name")
				}
			}
		}
	}
}
