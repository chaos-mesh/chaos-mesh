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

package recorder

import (
	"encoding"
	"errors"
	"reflect"
	"strconv"

	"github.com/iancoleman/strcase"
)

var InvalidType = errors.New("invalid type of fields")
var UknownType = errors.New("uknown type of fields")

var annotationPrefix = "chaos-mesh.org/"

func generateAnnotations(e ChaosEvent) (map[string]string, error) {
	annotations := make(map[string]string)

	if e == nil {
		return annotations, nil
	}

	val := reflect.ValueOf(e)
	val = reflect.Indirect(val)
	for index := 0; index < val.NumField(); index++ {
		fieldName := val.Type().Field(index).Name
		key := annotationPrefix + strcase.ToKebab(fieldName)
		field := val.Field(index)
		switch field.Kind() {
		case reflect.Invalid:
			return nil, InvalidType
		case reflect.String:
			annotations[key] = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			annotations[key] = strconv.Itoa(int(field.Int()))
		default:
			if field, ok := field.Interface().(encoding.TextMarshaler); ok {
				text, err := field.MarshalText()
				if err != nil {
					return nil, err
				}

				annotations[key] = string(text)
			} else {
				return nil, InvalidType
			}
		}
	}
	annotations[annotationPrefix+"type"] = strcase.ToKebab(val.Type().Name())

	return annotations, nil
}

// FromAnnotations will iterate over all the registered event,
// return `nil` if there is no suitable event.
func FromAnnotations(annotations map[string]string) (ChaosEvent, error) {
	typeName := annotations[annotationPrefix+"type"]
	ev := allEvents[typeName]

	if ev == nil {
		return nil, UknownType
	}

	val := reflect.ValueOf(ev)
	val = reflect.Indirect(val)
	newEmptyValue := reflect.Indirect(reflect.New(val.Type()))

	for index := 0; index < newEmptyValue.NumField(); index++ {
		fieldName := newEmptyValue.Type().Field(index).Name
		key := annotationPrefix + strcase.ToKebab(fieldName)
		field := newEmptyValue.Field(index)
		switch field.Kind() {
		case reflect.Invalid:
			return nil, InvalidType
		case reflect.String:
			field.SetString(annotations[key])
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num, err := strconv.Atoi(annotations[key])
			if err != nil {
				return nil, err
			}
			field.SetInt(int64(num))
		default:
			if unmarshaler, ok := field.Interface().(encoding.TextUnmarshaler); ok {
				err := unmarshaler.UnmarshalText([]byte(annotations[key]))
				if err != nil {
					return nil, err
				}
			} else if unmarshaler, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
				err := unmarshaler.UnmarshalText([]byte(annotations[key]))
				if err != nil {
					return nil, err
				}
			} else {
				return nil, InvalidType
			}
		}
	}
	return newEmptyValue.Interface().(ChaosEvent), nil
}
