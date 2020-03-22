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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var stressChaosLog = ctrl.Log.WithName("stresschaos-resource")

// SetupWebhookWithManager setup StressChaos's webhook with manager
func (in *StressChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-pingcap-com-v1alpha1-stresschaos,mutating=true,failurePolicy=fail,groups=pingcap.com,resources=stresschaos,verbs=create;update,versions=v1alpha1,name=mstresschaos.kb.io

var _ webhook.Defaulter = &StressChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *StressChaos) Default() {
	stressChaosLog.Info("default", "name", in.Name)
	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-pingcap-com-v1alpha1-stresschaos,mutating=false,failurePolicy=fail,groups=pingcap.com,resources=stresschaos,versions=v1alpha1,name=vstresschaos.kb.io

var _ ChaosValidator = &StressChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateCreate() error {
	stressChaosLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateUpdate(old runtime.Object) error {
	stressChaosLog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateDelete() error {
	stressChaosLog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *StressChaos) Validate() error {
	specField := field.NewPath("spec")

	errList := in.ValidateScheduler(specField)
	if len(errList) == 0 {
		errList = field.ErrorList{}
	}
	if strings.TrimSpace(in.Spec.Stressors) == "" {
		errList = append(errList, field.Invalid(
			specField.Child("stressors"), in.Spec.Stressors,
			"stressors should always be set"))
	}
	if len(errList) > 0 {
		return fmt.Errorf(errList.ToAggregate().Error())
	}
	return nil

}

// ValidateScheduler validates the scheduler and duration
func (in *StressChaos) ValidateScheduler(root *field.Path) field.ErrorList {
	if in.Spec.Duration != nil && in.Spec.Scheduler != nil {
		return nil
	} else if in.Spec.Duration == nil && in.Spec.Scheduler == nil {
		return nil
	}

	allErrs := field.ErrorList{}
	schedulerField := root.Child("scheduler")

	allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidateSchedulerError))
	return allErrs
}
