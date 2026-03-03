// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package apivalidator

import (
	"strconv"

	"github.com/go-playground/validator/v10"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// NamespaceSelectorsValid can be used to check whether namespace selectors is valid.
func NamespaceSelectorsValid(fl validator.FieldLevel) bool {
	ns, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}

	for _, n := range ns {
		if len(n) == 0 || len(n) > 63 {
			return false
		}

		if !checkName(n) {
			return false
		}
	}

	return true
}

// MapSelectorsValid can be used to check whether map selectors is valid.
func MapSelectorsValid(fl validator.FieldLevel) bool {
	if fl.Field().IsNil() {
		return true
	}

	ms, ok := fl.Field().Interface().(map[string]string)
	if !ok {
		return false
	}

	for k := range ms {
		if len(validation.IsQualifiedName(k)) != 0 {
			return false
		}
	}

	return true
}

// RequirementSelectorsValid can be used to check whether label requirement selectors is valid.
func RequirementSelectorsValid(fl validator.FieldLevel) bool {
	if fl.Field().IsNil() {
		return true
	}

	rs, ok := fl.Field().Interface().([]metav1.LabelSelectorRequirement)
	if !ok {
		return false
	}

	for _, r := range rs {
		if len(validation.IsQualifiedName(r.Key)) != 0 {
			return false
		}

		switch r.Operator {
		case metav1.LabelSelectorOpIn, metav1.LabelSelectorOpNotIn:
			if len(r.Values) == 0 {
				return false
			}
		case metav1.LabelSelectorOpExists, metav1.LabelSelectorOpDoesNotExist:
			if len(r.Values) > 0 {
				return false
			}
		default:
			// unsupport operator
			return false
		}
	}

	return true
}

// PhaseSelectorsValid can be used to check whether phase selectors is valid.
func PhaseSelectorsValid(fl validator.FieldLevel) bool {
	ph, ok := fl.Field().Interface().([]string)
	if !ok {
		return false
	}

	for _, phase := range ph {
		if !checkPhase(phase) {
			return false
		}
	}

	return true
}

// ValueValid can be used to check whether the mode value is valid.
func ValueValid(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return false
	}

	if f < 0 {
		return false
	}

	return true
}

func checkPhase(ph string) bool {
	phases := []corev1.PodPhase{
		corev1.PodRunning,
		corev1.PodFailed,
		corev1.PodPending,
		corev1.PodSucceeded,
		corev1.PodUnknown,
		corev1.PodPending,
	}

	for _, phase := range phases {
		if string(phase) == ph {
			return true
		}
	}

	return false
}

// PodsValid can be used to check whether the pod name is valid.
func PodsValid(fl validator.FieldLevel) bool {
	if fl.Field().IsNil() {
		return true
	}

	pods, ok := fl.Field().Interface().(map[string][]string)
	if !ok {
		return false
	}

	for ns, ps := range pods {
		if !checkName(ns) {
			return false
		}

		for _, p := range ps {
			if !checkName(p) {
				return false
			}
		}
	}

	return true
}

// PhysicalMachineValid can be used to check whether the physicalMachine name is valid.
func PhysicalMachineValid(fl validator.FieldLevel) bool {
	if fl.Field().IsNil() {
		return true
	}

	physicalMachines, ok := fl.Field().Interface().(map[string][]string)
	if !ok {
		return false
	}

	for ns, ps := range physicalMachines {
		if !checkName(ns) {
			return false
		}

		for _, p := range ps {
			if !checkName(p) {
				return false
			}
		}
	}

	return true
}
