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

package validation

import (
	"github.com/pingcap/chaos-mesh/api/v1alpha1"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	invalidConfigurationMsg = "invalid configuration"
)

var log = ctrl.Log.WithName("validate-webhook")

// ValidateChaos handles the validation for chaos api
func ValidateChaos(res *admissionv1beta1.AdmissionRequest, kind string) *admissionv1beta1.AdmissionResponse {
	var (
		err       error
		permitted bool
		msg       string
		validator v1alpha1.Validator
	)

	validator, err = v1alpha1.ParseValidator(res.Object.Raw, kind)
	if err == nil {
		permitted, msg, err = validator.Validate()
	}

	if err != nil {
		log.Error(err, "chaos is invalided", "kind", kind, "chaos", string(res.Object.Raw))
		return &admissionv1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	return &admissionv1beta1.AdmissionResponse{
		Allowed: permitted,
		Result: &metav1.Status{
			Message: msg,
		},
	}
}
