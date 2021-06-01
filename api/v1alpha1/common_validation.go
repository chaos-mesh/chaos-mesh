package v1alpha1

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	// ValidateValueParseError defines the error message for value parse error
	ValidateValueParseError = "parse value field error:%s"
)

func validateDuration(schedulerObject InnerObject, spec *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	durationField := spec.Child("duration")
	_, err := schedulerObject.GetDuration()
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
