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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var awschaoslog = logf.Log.WithName("awschaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-awschaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=awschaos,verbs=create;update,versions=v1alpha1,name=mawschaos.kb.io

var _ webhook.Defaulter = &AwsChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *AwsChaos) Default() {
	awschaoslog.Info("default", "name", in.Name)
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-awschaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=awschaos,versions=v1alpha1,name=vawschaos.kb.io

var _ ChaosValidator = &AwsChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *AwsChaos) ValidateCreate() error {
	awschaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *AwsChaos) ValidateUpdate(old runtime.Object) error {
	awschaoslog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *AwsChaos) ValidateDelete() error {
	awschaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *AwsChaos) Validate() error {
	specField := field.NewPath("spec")
	allErrs := in.ValidateScheduler(specField)
	allErrs = append(allErrs, in.ValidatePodMode(specField)...)
	allErrs = append(allErrs, in.Spec.validateEbsVolume(specField.Child("volumeID"))...)
	allErrs = append(allErrs, in.Spec.validateDeviceName(specField.Child("deviceName"))...)

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *AwsChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	schedulerField := spec.Child("scheduler")

	switch in.Spec.Action {
	case Ec2Stop, DetachVolume:
		allErrs = append(allErrs, ValidateScheduler(in, spec)...)
	case Ec2Restart:
		// We choose to ignore the Duration property even user define it
		if in.Spec.Scheduler != nil {
			_, err := ParseCron(in.Spec.Scheduler.Cron, schedulerField.Child("cron"))
			allErrs = append(allErrs, err...)
		}
	default:
		err := fmt.Errorf("awschaos[%s/%s] have unknown action type", in.Namespace, in.Name)
		log.Error(err, "Wrong AwsChaos Action type")

		actionField := spec.Child("action")
		allErrs = append(allErrs, field.Invalid(actionField, in.Spec.Action, err.Error()))
	}
	return allErrs
}

// ValidatePodMode validates the value with podmode
func (in *AwsChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	// Because aws chaos does not need a pod mode, so return nil here.
	return nil
}

// validateEbsVolume validates the EbsVolume
func (in *AwsChaosSpec) validateEbsVolume(containerField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == DetachVolume {
		if in.EbsVolume == nil {
			err := fmt.Errorf("the ID of EBS volume should not be empty on %s action", in.Action)
			allErrs = append(allErrs, field.Invalid(containerField, in.EbsVolume, err.Error()))
		}
	}
	return allErrs
}

// validateDeviceName validates the DeviceName
func (in *AwsChaosSpec) validateDeviceName(containerField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == DetachVolume {
		if in.DeviceName == nil {
			err := fmt.Errorf("the name of device should not be empty on %s action", in.Action)
			allErrs = append(allErrs, field.Invalid(containerField, in.DeviceName, err.Error()))
		}
	}
	return allErrs
}
