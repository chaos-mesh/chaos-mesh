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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var httpchaoslog = logf.Log.WithName("httpchaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-httpchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=httpchaos,verbs=create;update,versions=v1alpha1,name=mhttpchaos.kb.io

var _ webhook.Defaulter = &HTTPChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *HTTPChaos) Default() {
	httpchaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
	in.Spec.Default()
}

func (in *HTTPChaosSpec) Default() {

}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-httpchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=httpchaos,versions=v1alpha1,name=vhttpchaos.kb.io

var _ webhook.Validator = &HTTPChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateCreate() error {
	httpchaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateUpdate(old runtime.Object) error {
	httpchaoslog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*HTTPChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *HTTPChaos) ValidateDelete() error {
	httpchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *HTTPChaos) Validate() error {

	allErrs := in.Spec.Validate()
	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}

	return nil
}

func (in *HTTPChaosSpec) Validate() field.ErrorList {
	specField := field.NewPath("spec")

	allErrs := validatePodSelector(in.PodSelector.Value, in.PodSelector.Mode, specField.Child("value"))
	allErrs = append(allErrs, validateDuration(in, specField)...)
	return allErrs

}
