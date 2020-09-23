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
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var iochaoslog = logf.Log.WithName("iochaos-resource")

// SetupWebhookWithManager setup IoChaos's webhook with manager
func (in *IoChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-iochaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=iochaos,verbs=create;update,versions=v1alpha1,name=miochaos.kb.io

var _ webhook.Defaulter = &IoChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *IoChaos) Default() {
	iochaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-iochaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=iochaos,versions=v1alpha1,name=viochaos.kb.io

var _ ChaosValidator = &IoChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *IoChaos) ValidateCreate() error {
	iochaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *IoChaos) ValidateUpdate(old runtime.Object) error {
	iochaoslog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *IoChaos) ValidateDelete() error {
	iochaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *IoChaos) Validate() error {
	specField := field.NewPath("spec")
	allErrs := in.ValidateScheduler(specField)
	allErrs = append(allErrs, in.ValidatePodMode(specField)...)
	allErrs = append(allErrs, in.Spec.validateDelay(specField.Child("delay"))...)
	allErrs = append(allErrs, in.Spec.validateErrno(specField.Child("errno"))...)
	allErrs = append(allErrs, in.Spec.validatePercent(specField.Child("percent"))...)

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *IoChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	return ValidateScheduler(in, spec)
}

// ValidatePodMode validates the value with podmode
func (in *IoChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	return ValidatePodMode(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}

func (in *IoChaosSpec) validateDelay(delay *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == IoLatency {
		_, err := time.ParseDuration(in.Delay)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(delay, in.Delay,
				fmt.Sprintf("parse delay field error:%s for action:%s", err, in.Action)))
		}
	}
	return allErrs
}

func (in *IoChaosSpec) validateErrno(errno *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == IoFaults {
		if in.Errno == 0 {
			allErrs = append(allErrs, field.Invalid(errno, in.Errno,
				fmt.Sprintf("action %s: errno 0 is not supported", in.Action)))
		}
	}
	return allErrs
}

func (in *IoChaosSpec) validatePercent(percentField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Percent > 100 || in.Percent < 0 {
		allErrs = append(allErrs, field.Invalid(percentField, in.Percent,
			"percent field should be in 0-100"))
	}

	return allErrs
}
