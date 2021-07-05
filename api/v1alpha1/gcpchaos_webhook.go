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

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validateDeviceName validates the DeviceName
func (in *GcpChaosAction) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch *in {
	case NodeStop, DiskLoss:
	case NodeReset:
	default:
		err := fmt.Errorf("gcpchaos have unknown action type")
		log.Error(err, "Wrong GcpChaos Action type")

		allErrs = append(allErrs, field.Invalid(path, in, err.Error()))
	}
	return allErrs
}

// validateDeviceName validates the DeviceName
func (in *GcpDeviceNames) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	obj := root.(*GcpChaos)
	if obj.Spec.Action == DiskLoss {
		if *in == nil {
			err := fmt.Errorf("at least one device name is required on %s action", obj.Spec.Action)
			allErrs = append(allErrs, field.Invalid(path, in, err.Error()))
		}
	}
	return allErrs
}
