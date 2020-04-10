// Copyright 2020 PingCAP, Inc.
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

	"github.com/robfig/cron/v3"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// ValidateSchedulerError defines the error message for ValidateScheduler
	ValidateSchedulerError = "schedule and duration should be omitted or defined at the same time"

	// ValidatePodchaosSchedulerError defines the error message for ValidateScheduler of Podchaos
	ValidatePodchaosSchedulerError = "schedule should be omitted"

	// ValidateValueParseError defines the error message for value parse error
	ValidateValueParseError = "parse value field error:%s"
)

// ValidateScheduler validates the scheduler
func ValidateScheduler(obj InnerSchedulerObject, spec *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	schedulerField := spec.Child("scheduler")
	durationField := spec.Child("duration")

	duration, err := obj.GetDuration()
	if err != nil {
		allErrs = append(allErrs, field.Invalid(durationField,
			duration,
			fmt.Sprintf("parse duration field error:%s", err)))
	}

	scheduler := obj.GetScheduler()
	if duration != nil && scheduler != nil {
		_, err := cron.ParseStandard(scheduler.Cron)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(schedulerField.Child("cron"),
				scheduler.Cron,
				fmt.Sprintf("parse cron field error:%s", err)))
		}

	} else if (duration == nil && scheduler != nil) || (duration != nil && scheduler == nil) {
		allErrs = append(allErrs, field.Invalid(schedulerField, scheduler, ValidateSchedulerError))
	}
	return allErrs
}

// ValidatePodMode validates the value with podmode
func ValidatePodMode(value string, mode PodMode, valueField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	switch mode {
	case FixedPodMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf(ValidateValueParseError, err)))
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
		}

		if percentage <= 0 || percentage > 100 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value of %d is invalid, Must be (0,100] with mode:%s",
					percentage, mode)))
		}
	}
	return allErrs
}
