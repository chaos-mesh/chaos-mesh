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

package v1alpha1

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/chaos-mesh/chaos-mesh/api/genericwebhook"
)

func TestDNSStaticMappingsValidation(t *testing.T) {
	cases := []struct {
		name     string
		mappings []string
		invalid  bool
	}{
		{name: "IPv4 and IPv6", mappings: []string{"Example.COM.=192.0.2.1", "v6.example.com=2001:db8::1"}},
		{name: "nil", invalid: true},
		{name: "empty", mappings: []string{}, invalid: true},
		{name: "missing separator", mappings: []string{"example.com"}, invalid: true},
		{name: "multiple separators", mappings: []string{"example.com=192.0.2.1=bad"}, invalid: true},
		{name: "invalid IPv4", mappings: []string{"example.com=999.0.0.1"}, invalid: true},
		{name: "missing IP", mappings: []string{"example.com="}, invalid: true},
		{name: "missing name", mappings: []string{"=192.0.2.1"}, invalid: true},
		{name: "wildcard", mappings: []string{"*.example.com=192.0.2.1"}, invalid: true},
		{name: "question mark", mappings: []string{"exam?le.com=192.0.2.1"}, invalid: true},
		{name: "empty label", mappings: []string{"example..com=192.0.2.1"}, invalid: true},
		{name: "two trailing dots", mappings: []string{"example.com..=192.0.2.1"}, invalid: true},
		{name: "whitespace", mappings: []string{" example.com=192.0.2.1"}, invalid: true},
		{name: "long label", mappings: []string{strings.Repeat("a", 64) + ".com=192.0.2.1"}, invalid: true},
		{name: "zone", mappings: []string{"example.com=fe80::1%eth0"}, invalid: true},
		{name: "normalized duplicate", mappings: []string{"Example.COM.=192.0.2.1", "example.com=192.0.2.2"}, invalid: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := DNSChaosSpec{Action: StaticAction, DomainIpMappingList: tc.mappings}
			errs := spec.Validate(nil, field.NewPath("spec"))
			if (len(errs) > 0) != tc.invalid {
				t.Fatalf("validation errors = %v, want invalid %v", errs, tc.invalid)
			}
			if len(errs) > 0 && errs[0].Field != "spec.domainIpMapping" {
				t.Fatalf("error field = %q", errs[0].Field)
			}
		})
	}
}

func TestDNSStaticMappingsNormalize(t *testing.T) {
	got, err := ParseDNSDomainIPMappings([]string{"Example.COM.=::ffff:192.0.2.1", "V6.EXAMPLE.COM=2001:0db8::1"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"example.com": "192.0.2.1", "v6.example.com": "2001:db8::1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mappings = %v, want %v", got, want)
	}
}

func TestDNSStaticMappingsGenericWebhook(t *testing.T) {
	obj := &DNSChaos{Spec: DNSChaosSpec{Action: StaticAction, ContainerSelector: ContainerSelector{PodSelector: PodSelector{Mode: AllMode}}}}
	errs := genericwebhook.Validate(obj)
	found := false
	for _, err := range errs {
		if err.Field == "Spec.domainIpMapping" {
			found = true
		}
	}
	if !found {
		t.Fatalf("generic webhook did not reject missing static mappings: %v", errs)
	}
	for _, action := range []DNSChaosAction{ErrorAction, RandomAction} {
		obj.Spec.Action = action
		obj.Spec.DomainIpMappingList = []string{"ignored legacy value"}
		if errs := obj.Spec.Validate(nil, field.NewPath("spec")); len(errs) > 0 {
			t.Fatalf("%s rejected unused mappings: %v", action, errs)
		}
	}
}
