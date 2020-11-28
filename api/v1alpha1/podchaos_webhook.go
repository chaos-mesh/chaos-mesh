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
var podchaoslog = logf.Log.WithName("podchaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-podchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=podchaos,verbs=create;update,versions=v1alpha1,name=mpodchaos.kb.io

var _ webhook.Defaulter = &PodChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *PodChaos) Default() {
	podchaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-podchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=podchaos,versions=v1alpha1,name=vpodchaos.kb.io

var _ ChaosValidator = &PodChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PodChaos) ValidateCreate() error {
	podchaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PodChaos) ValidateUpdate(old runtime.Object) error {
	podchaoslog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *PodChaos) ValidateDelete() error {
	podchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *PodChaos) Validate() error {
	specField := field.NewPath("spec")
	allErrs := in.ValidateScheduler(specField)
	allErrs = append(allErrs, in.ValidatePodMode(specField)...)
	allErrs = append(allErrs, in.Spec.validateContainerName(specField.Child("containerName"))...)

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *PodChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	schedulerField := spec.Child("scheduler")

	switch in.Spec.Action {
	case PodFailureAction:
		allErrs = append(allErrs, ValidateScheduler(in, spec)...)
	case PodKillAction:
		// We choose to ignore the Duration property even user define it
		if in.Spec.Scheduler == nil {
			allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidatePodchaosSchedulerError))
		} else {
			_, err := ParseCron(in.Spec.Scheduler.Cron, schedulerField.Child("cron"))
			allErrs = append(allErrs, err...)
		}
	case ContainerKillAction:
		// We choose to ignore the Duration property even user define it
		if in.Spec.Scheduler == nil {
			allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidatePodchaosSchedulerError))
		} else {
			_, err := ParseCron(in.Spec.Scheduler.Cron, schedulerField.Child("cron"))
			allErrs = append(allErrs, err...)
		}
	default:
		err := fmt.Errorf("podchaos[%s/%s] have unknown action type", in.Namespace, in.Name)
		log.Error(err, "Wrong PodChaos Action type")

		actionField := spec.Child("action")
		allErrs = append(allErrs, field.Invalid(actionField, in.Spec.Action, err.Error()))
	}
	return allErrs
}

// ValidatePodMode validates the value with podmode
func (in *PodChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	return ValidatePodMode(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}

// validateContainerName validates the ContainerName
func (in *PodChaosSpec) validateContainerName(containerField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == ContainerKillAction {
		if in.ContainerName == "" {
			err := fmt.Errorf("the name of container should not be empty on %s action", in.Action)
			allErrs = append(allErrs, field.Invalid(containerField, in.ContainerName, err.Error()))
		}
	}
	return allErrs
}
