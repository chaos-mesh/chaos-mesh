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
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

type FieldValidator interface {
	Validate(root interface{}, path *field.Path) field.ErrorList
}

func Validate(obj interface{}) field.ErrorList {
	errorList := field.ErrorList{}

	root := obj
	walker := NewFieldWalker(obj, func(path *field.Path, obj interface{}, field reflect.StructField) bool {
		val := reflect.ValueOf(obj)
		for {
			if !val.IsValid() {
				return true
			}
			obj = val.Interface()
			if validator, ok := obj.(FieldValidator); ok {
				errs := validator.Validate(root, path)
				errorList = append(errorList, errs...)

				return true
			}

			if val.Kind() != reflect.Ptr {
				break
			}

			val = val.Elem()
		}

		return true
	})
	walker.Walk()

	return errorList
}

func Aggregate(errors field.ErrorList) error {
	if errors == nil || len(errors) == 0 {
		return nil
	}
	return fmt.Errorf(errors.ToAggregate().Error())
}
