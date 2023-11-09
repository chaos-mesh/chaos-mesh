// Copyright 2023 Chaos Mesh Authors.
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
)

func (in *CloudStackAPIConfig) Default(root interface{}, structField *reflect.StructField) {
	setDefaultsFromTags(in)
}

func (in *CloudStackVMChaosSpec) Default(root interface{}, structField *reflect.StructField) {
	setDefaultsFromTags(in)
}

func setDefaultsFromTags(in interface{}) {
	fields := reflect.TypeOf(in).Elem()
	values := reflect.ValueOf(in).Elem()

	for i := 0; i < fields.NumField(); i++ {
		field := fields.Field(i)

		if tagValue, ok := field.Tag.Lookup("default"); ok {
			switch values.Field(i).Kind() {
			case reflect.String:
				values.Field(i).SetString(tagValue)
			case reflect.Bool:
				values.Field(i).SetBool(tagValue == "true")
			}
		}
	}
}
