// Copyright 2021 Chaos Mesh Authors.
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

package webhook

import (
	"reflect"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1/genericwebhook"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func affectedNamespaces(obj interface{}) map[string]struct{} {
	namespaces := make(map[string]struct{})

	walker := genericwebhook.NewFieldWalker(obj, func(path *field.Path, obj interface{}, field *reflect.StructField) bool {
		// These are some trivial rules to cut a lot of edges.
		if field != nil && (field.Name == "Status" || field.Name == "TypeMeta" || field.Name == "ObjectMeta") {
			return false
		}

		if selector, ok := obj.(*v1alpha1.PodSelector); ok {
			for _, ns := range selector.Selector.Namespaces {
				namespaces[ns] = struct{}{}
			}

			return true
		}
		return true
	})
	walker.Walk()

	return namespaces
}
