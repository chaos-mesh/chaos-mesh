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

// validateContainerNames validates the ContainerNames
func (in *PodChaosSpec) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == ContainerKillAction {
		if len(in.ContainerSelector.ContainerNames) == 0 {
			err := fmt.Errorf("the name of container should not be empty on %s action", in.Action)
			allErrs = append(allErrs, field.Invalid(path.Child("containerNames"), in.ContainerNames, err.Error()))
		}
	}
	return allErrs
}
