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

package v1alpha1

import (
	"fmt"
	"strconv"
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

// +kubebuilder:webhook:path=/mutate-pingcap-com-v1alpha1-iochaos,mutating=true,failurePolicy=fail,groups=pingcap.com,resources=iochaos,verbs=create;update,versions=v1alpha1,name=miochaos.kb.io

var _ webhook.Defaulter = &IoChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *IoChaos) Default() {
	iochaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-pingcap-com-v1alpha1-iochaos,mutating=false,failurePolicy=fail,groups=pingcap.com,resources=iochaos,versions=v1alpha1,name=viochaos.kb.io

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
	allErrs = append(allErrs, in.ValidateValue(specField)...)
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
	return ValidateScheduler(in.Spec.Duration, in.Spec.Scheduler, spec)
}

// ValidateValue validates the value
func (in *IoChaos) ValidateValue(spec *field.Path) field.ErrorList {
	return ValidateValue(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}

func (in *IoChaosSpec) validateDelay(delay *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Action == IODelayAction || in.Action == IOMixedAction {
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
	if in.Action == IOErrnoAction || in.Action == IOMixedAction {
		if in.Errno != "" {
			_, err := strconv.Atoi(in.Errno)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(errno, in.Errno,
					fmt.Sprintf("parse errno field error:%s for action:%s", err, in.Action)))
			}
		}
	}
	return allErrs
}

func (in *IoChaosSpec) validatePercent(percentField *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.Percent != "" {
		percent, err := strconv.Atoi(in.Percent)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(percentField, in.Percent,
				fmt.Sprintf("parse percent field error:%s", err)))
		}

		if percent <= 0 || percent > 100 {
			allErrs = append(allErrs, field.Invalid(percentField, in.Percent,
				fmt.Sprintf("percent value of %d is invalid, Must be (0,100]", percent)))
		}
	}
	return allErrs
}
