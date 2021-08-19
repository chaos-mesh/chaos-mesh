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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	//"github.com/chaos-mesh/chaos-mesh/api/v1alpha1/genericwebhook"
)

// log is for logging in this package.
var physicalmachinechaoslog = logf.Log.WithName("physicalmachinechaos-resource")

type ExpUID string

func (in *ExpUID) Default(root interface{}, field reflect.StructField) {
	if in == nil {
		return
	}

	if len(*in) == 0 {
		*in = ExpUID(uuid.New().String())
		physicalmachinechaoslog.Info("PhysicalMachineChaosSpec default", "UID", string(*in))
	}
}

type Address []string

func (in *Address) Default(root interface{}, field reflect.StructField) {
	if in == nil {
		return
	}

	if len(*in) == 0 {
		return
	}

	newAddress := []string(*in)

	// add http prefix for address
	for i := range newAddress {
		if !strings.HasPrefix(newAddress[i], "http") {
			newAddress[i] = fmt.Sprintf("http://%s", newAddress[i])
		}
	}
	*in = Address(newAddress)
}

func (in *PhysicalMachineChaosSpec) Default(root interface{}, field reflect.StructField) {
	if in == nil {
		return
	}

	if len(in.UID) == 0 {
		in.UID = uuid.New().String()
		physicalmachinechaoslog.Info("PhysicalMachineChaosSpec default", "UID", in.UID)
	}

	for i := range in.Address {
		// add http prefix for address
		if !strings.HasPrefix(in.Address[i], "http") {
			in.Address[i] = fmt.Sprintf("http://%s", in.Address[i])
		}
	}
}

func (in *PhysicalMachineChaosSpec) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	// make sure the configuration corresponding to action is not empty
	var inInterface map[string]interface{}
	inrec, err := json.Marshal(in)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(path.Child("spec"), in, err.Error()))
	}

	err = json.Unmarshal(inrec, &inInterface)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(path.Child("spec"), in, err.Error()))
	}

	if _, ok := inInterface[string(in.Action)]; !ok {
		allErrs = append(allErrs,
			field.Invalid(path.Child("spec"), in,
				"the configuration corresponding to action is empty"))
	}

	// make sure address is not empty
	if len(in.Address) == 0 {
		allErrs = append(allErrs,
			field.Invalid(path.Child("address"), in.Address, "the address is empty"))
	}
	for _, address := range in.Address {
		if len(address) == 0 {
			allErrs = append(allErrs,
				field.Invalid(path.Child("address"), in.Address, "the address is empty"))
		}
	}

	return allErrs
}
