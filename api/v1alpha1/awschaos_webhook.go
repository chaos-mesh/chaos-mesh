// Copyright 2020 Chaos Mesh Authors.
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

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (in *EbsVolume) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	awsChaos := root.(*AwsChaos)
	if awsChaos.Spec.Action == DetachVolume {
		if in == nil {
			err := fmt.Errorf("the ID of EBS volume should not be empty on %s action", awsChaos.Spec.Action)
			allErrs = append(allErrs, field.Invalid(path, in, err.Error()))
		}
	}

	return allErrs
}

func (in *AwsDeviceName) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	awsChaos := root.(*AwsChaos)
	if awsChaos.Spec.Action == DetachVolume {
		if in == nil {
			err := fmt.Errorf("the name of device should not be empty on %s action", awsChaos.Spec.Action)
			allErrs = append(allErrs, field.Invalid(path, in, err.Error()))
		}
	}

	return allErrs
}

// ValidateScheduler validates the scheduler and duration
func (in *AwsChaosAction) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch *in {
	case Ec2Stop, DetachVolume:
	case Ec2Restart:
	default:
		err := fmt.Errorf("awschaos have unknown action type")
		log.Error(err, "Wrong AwsChaos Action type")

		allErrs = append(allErrs, field.Invalid(path, in, err.Error()))
	}
	return allErrs
}
