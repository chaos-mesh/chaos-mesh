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

package genericwebhook

import (
	"reflect"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

type Defaulter interface {
	Default(root interface{}, field reflect.StructField)
}

func Default(obj interface{}) {
	root := obj
	walker := NewFieldWalker(obj, func(path *field.Path, obj interface{}, field reflect.StructField) bool {
		if defaulter, ok := obj.(Defaulter); ok {
			defaulter.Default(root, field)

			return true
		}

		return true
	})
	walker.Walk()
}
