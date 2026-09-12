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

package registry

import (
	"errors"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

// mockSelector is a minimal implementation of generic.Selector for testing.
type mockSelector struct{}

func (m *mockSelector) ListFunc(_ client.Reader) generic.ListFunc { return nil }
func (m *mockSelector) ListOption() client.ListOption             { return nil }
func (m *mockSelector) Match(_ client.Object) bool                { return true }

func TestParse_EmptyRegistry(t *testing.T) {
	reg := Registry{}
	chain, err := Parse(reg, v1alpha1.GenericSelectorSpec{}, generic.Option{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 0 {
		t.Errorf("expected empty chain, got length %d", len(chain))
	}
}

func TestParse_SingleFactory(t *testing.T) {
	reg := Registry{
		"mock": func(_ v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return &mockSelector{}, nil
		},
	}
	chain, err := Parse(reg, v1alpha1.GenericSelectorSpec{}, generic.Option{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 1 {
		t.Errorf("expected 1 selector in chain, got %d", len(chain))
	}
}

func TestParse_MultipleFactories(t *testing.T) {
	reg := Registry{
		"mock1": func(_ v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return &mockSelector{}, nil
		},
		"mock2": func(_ v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return &mockSelector{}, nil
		},
	}
	chain, err := Parse(reg, v1alpha1.GenericSelectorSpec{}, generic.Option{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chain) != 2 {
		t.Errorf("expected 2 selectors in chain, got %d", len(chain))
	}
}

func TestParse_FactoryReturnsError(t *testing.T) {
	reg := Registry{
		"failing": func(_ v1alpha1.GenericSelectorSpec, _ generic.Option) (generic.Selector, error) {
			return nil, errors.New("factory failed")
		},
	}
	_, err := Parse(reg, v1alpha1.GenericSelectorSpec{}, generic.Option{})
	if err == nil {
		t.Error("expected error when factory fails, got nil")
	}
}

func TestParse_SpecAndOptionPassedToFactory(t *testing.T) {
	wantNamespace := "test-ns"
	wantClusterScoped := true

	spec := v1alpha1.GenericSelectorSpec{
		Namespaces: []string{wantNamespace},
	}
	option := generic.Option{
		ClusterScoped: wantClusterScoped,
	}

	var gotSpec v1alpha1.GenericSelectorSpec
	var gotOption generic.Option

	reg := Registry{
		"capture": func(s v1alpha1.GenericSelectorSpec, o generic.Option) (generic.Selector, error) {
			gotSpec = s
			gotOption = o
			return &mockSelector{}, nil
		},
	}

	_, err := Parse(reg, spec, option)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotSpec.Namespaces) == 0 || gotSpec.Namespaces[0] != wantNamespace {
		t.Errorf("expected namespace %q, got %v", wantNamespace, gotSpec.Namespaces)
	}
	if gotOption.ClusterScoped != wantClusterScoped {
		t.Errorf("expected ClusterScoped=%v, got %v", wantClusterScoped, gotOption.ClusterScoped)
	}
}
