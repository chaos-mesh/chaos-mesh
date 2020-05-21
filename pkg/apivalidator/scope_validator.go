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

package apivalidator

import (
	"github.com/go-playground/validator/v10"

	corev1 "k8s.io/api/core/v1"
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

	for k, v := range ms {
		if !checkName(k) || !checkName(v) {
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
