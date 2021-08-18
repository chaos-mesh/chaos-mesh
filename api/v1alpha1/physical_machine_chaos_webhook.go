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
	"strings"

	"github.com/google/uuid"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1/genericwebhook"
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

	for i := range newAddress {
		// add http prefix for address
		if !strings.HasPrefix(newAddress[i], "http") {
			newAddress[i] = fmt.Sprintf("http://%s", newAddress[i])
		}
	}
	*in = Address(newAddress)
}

func init() {
	genericwebhook.Register("ExpUID", reflect.PtrTo(reflect.TypeOf(ExpUID(""))))
	genericwebhook.Register("Address", reflect.PtrTo(reflect.TypeOf(Address([]string{}))))
}
