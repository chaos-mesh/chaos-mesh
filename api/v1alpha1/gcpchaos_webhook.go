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
var gcpchaoslog = logf.Log.WithName("gcpchaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-gcpchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=gcpchaos,verbs=create;update,versions=v1alpha1,name=mgcpchaos.kb.io

var _ webhook.Defaulter = &GCPChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *GCPChaos) Default() {
	gcpchaoslog.Info("default", "name", in.Name)
	in.Spec.Default()
}

func (in *GCPChaosSpec) Default() {
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-gcpchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=gcpchaos,versions=v1alpha1,name=vgcpchaos.kb.io

var _ webhook.Validator = &GCPChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *GCPChaos) ValidateCreate() error {
	gcpchaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *GCPChaos) ValidateUpdate(old runtime.Object) error {
	gcpchaoslog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*GCPChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *GCPChaos) ValidateDelete() error {
	gcpchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *GCPChaos) Validate() error {
	allErrs := in.Spec.Validate()

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

func (in *GCPChaosSpec) Validate() field.ErrorList {
	specField := field.NewPath("spec")
	allErrs := in.validateDeviceName(specField.Child("deviceName"))
	allErrs = append(allErrs, validateDuration(in, specField)...)
	allErrs = append(allErrs, in.validateAction(specField)...)
	return allErrs
}

// validateDeviceName validates the DeviceName
func (in *GCPChaosSpec) validateAction(spec *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	switch in.Action {
	case NodeStop, DiskLoss:
	case NodeReset:
	default:
		err := fmt.Errorf("gcpchaos have unknown action type")
		log.Error(err, "Wrong GCPChaos Action type")

		actionField := spec.Child("action")
		allErrs = append(allErrs, field.Invalid(actionField, in.Action, err.Error()))
	}
	return allErrs
}

// validateDeviceName validates the DeviceName
func (in *GCPChaosSpec) validateDeviceName(containerField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == DiskLoss {
		if in.DeviceNames == nil {
			err := fmt.Errorf("at least one device name is required on %s action", in.Action)
			allErrs = append(allErrs, field.Invalid(containerField, in.DeviceNames, err.Error()))
		}
	}
	return allErrs
}
