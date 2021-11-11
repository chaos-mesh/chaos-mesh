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

package generic

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const InjectAnnotationKey = "chaos-mesh.org/inject"

type Option struct {
	ClusterScoped         bool
	TargetNamespace       string
	EnableFilterNamespace bool
}

type ListFunc func(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error

type Selector interface {
	ListFunc(client.Reader) ListFunc
	ListOption() client.ListOption
	Match(client.Object) bool
}

type SelectorChain []Selector

func (s SelectorChain) ListObjects(c client.Client, r client.Reader,
	listObj func(listFunc ListFunc, opts client.ListOptions) error) error {
	var options []client.ListOption
	listFunc := c.List

	for _, selector := range s {
		if tmpOption := selector.ListOption(); tmpOption != nil {
			options = append(options, selector.ListOption())
		}
		if tmpListFunc := selector.ListFunc(r); tmpListFunc != nil {
			listFunc = tmpListFunc
		}
	}
	opts := client.ListOptions{}
	opts.ApplyOptions(options)
	return listObj(listFunc, opts)
}

func (s SelectorChain) Match(obj client.Object) bool {
	for _, selector := range s {
		if !selector.Match(obj) {
			return false
		}
	}
	return true
}
