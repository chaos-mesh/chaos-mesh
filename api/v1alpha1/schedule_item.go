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

package v1alpha1

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ScheduleItem struct {
	EmbedChaos `json:",inline"`
	// +optional
	Workflow *WorkflowSpec `json:"workflow,omitempty"`
}

func (in EmbedChaos) Validate(chaosType string) field.ErrorList {
	allErrs := field.ErrorList{}
	spec := reflect.ValueOf(in).FieldByName(chaosType)

	if !spec.IsValid() || spec.IsNil() {
		allErrs = append(allErrs, field.Invalid(field.NewPath(chaosType),
			in,
			fmt.Sprintf("parse schedule field error: missing chaos spec")))
		return allErrs
	}
	addr, success := spec.Interface().(CommonSpec)
	if success == false {
		logf.Log.Info(fmt.Sprintf("%s does not seem to have a validator", chaosType))
		return allErrs
	}
	addr.Default()
	allErrs = append(allErrs, addr.Validate()...)
	return allErrs
}
