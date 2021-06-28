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
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// ValidateValueParseError defines the error message for value parse error
	ValidateValueParseError = "parse value field error:%s"
)

// +kubebuilder:object:generate=false
type CommonSpec interface {
	GetDuration() (*time.Duration, error)
	Validate() field.ErrorList
	Default()
}

func validateDuration(spec CommonSpec, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	durationField := path.Child("duration")
	_, err := spec.GetDuration()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(durationField, nil,
			fmt.Sprintf("parse duration field error:%s", err)))
	}

	return allErrs
}

// validatePodSelector validates the value with podmode
func validatePodSelector(value string, mode PodMode, valueField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch mode {
	case FixedPodMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf(ValidateValueParseError, err)))
			break
		}

		if num <= 0 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value must be greater than 0 with mode:%s", FixedPodMode)))
		}

	case RandomMaxPercentPodMode, FixedPercentPodMode:
		percentage, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf(ValidateValueParseError, err)))
			break
		}

		if percentage <= 0 || percentage > 100 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value of %d is invalid, Must be (0,100] with mode:%s",
					percentage, mode)))
		}
	}

	return allErrs
}
