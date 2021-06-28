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
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var timechaoslog = logf.Log.WithName("timechaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-timechaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=timechaos,verbs=create;update,versions=v1alpha1,name=mtimechaos.kb.io

var _ webhook.Defaulter = &TimeChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *TimeChaos) Default() {
	timechaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
	in.Spec.Default()
}

func (in *TimeChaosSpec) Default() {
	in.DefaultClockIds()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-timechaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=timechaos,versions=v1alpha1,name=vtimechaos.kb.io

var _ webhook.Validator = &TimeChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *TimeChaos) ValidateCreate() error {
	timechaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *TimeChaos) ValidateUpdate(old runtime.Object) error {
	timechaoslog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*TimeChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *TimeChaos) ValidateDelete() error {
	timechaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *TimeChaos) Validate() error {
	allErrs := in.Spec.Validate()

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

func (in *TimeChaosSpec) Validate() field.ErrorList {
	specField := field.NewPath("spec")
	allErrs := in.validateTimeOffset(specField.Child("timeOffset"))
	allErrs = append(allErrs, validateDuration(in, specField)...)

	return allErrs
}

// validateTimeOffset validates the timeOffset
func (in *TimeChaosSpec) validateTimeOffset(timeOffset *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	_, err := time.ParseDuration(in.TimeOffset)
	if err != nil {
		allErrs = append(allErrs, field.Invalid(timeOffset,
			in.TimeOffset,
			fmt.Sprintf("parse timeOffset field error:%s", err)))
	}

	return allErrs
}
