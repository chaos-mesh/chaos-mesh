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
	"time"

	"github.com/robfig/cron/v3"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// ValidateSchedulerError defined the error message for ValidateScheduler
	ValidateSchedulerError = "schedule and duration should be omitted or defined at the same time"

	// ValidatePodchaosSchedulerError defined the error message for ValidateScheduler of Podchaos
	ValidatePodchaosSchedulerError = "schedule should be omitted"

	// ValidateValueParseError defined the error message for value parse error
	ValidateValueParseError = "parse value field error"
)

// ValidateScheduler validates the scheduler and duration
func ValidateScheduler(duration *string, scheduler *SchedulerSpec, spec *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	schedulerField := spec.Child("scheduler")
	durationField := spec.Child("duration")

	if duration != nil && scheduler != nil {
		_, err := cron.ParseStandard(scheduler.Cron)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(schedulerField.Child("cron"),
				scheduler.Cron,
				fmt.Sprintf("parse cron field error:%s", err)))
		}

		_, err = time.ParseDuration(*duration)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(durationField,
				*duration,
				fmt.Sprintf("parse duration field error:%s", err)))
		}

		if len(allErrs) > 0 {
			return allErrs
		}

		return nil
	} else if duration == nil && scheduler == nil {
		return nil
	}

	allErrs = append(allErrs, field.Invalid(schedulerField, scheduler, ValidateSchedulerError))
	return allErrs
}

// ValidateValue validates the value with podmode
func ValidateValue(value string, mode PodMode, valueField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	switch mode {
	case FixedPodMode:
		num, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value, ValidateValueParseError))
		} else if num <= 0 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value must be greater than 0 with mode:%s", FixedPodMode)))
		}
		break

	case RandomMaxPercentPodMode, FixedPercentPodMode:
		percentage, err := strconv.Atoi(value)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(valueField, value, ValidateValueParseError))
		} else if percentage <= 0 || percentage > 100 {
			allErrs = append(allErrs, field.Invalid(valueField, value,
				fmt.Sprintf("value of %d is invalid, Must be (0,100] with mode:%s",
					percentage, mode)))
		}
		break
	}
	return allErrs
}
