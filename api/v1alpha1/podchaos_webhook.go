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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var podchaoslog = logf.Log.WithName("podchaos-resource")

// SetupWebhookWithManager setup PodChaos's webhook with manager
func (in *PodChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-pingcap-com-v1alpha1-podchaos,mutating=true,failurePolicy=fail,groups=pingcap.com,resources=podchaos,verbs=create;update,versions=v1alpha1,name=mpodchaos.kb.io

var _ webhook.Defaulter = &PodChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *PodChaos) Default() {
	podchaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-pingcap-com-v1alpha1-podchaos,mutating=false,failurePolicy=fail,groups=pingcap.com,resources=podchaos,versions=v1alpha1,name=vpodchaos.kb.io

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
	allErrs = append(allErrs, in.ValidateValue(specField)...)

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
		if in.Spec.Duration != nil && in.Spec.Scheduler != nil {
			return nil
		} else if in.Spec.Duration == nil && in.Spec.Scheduler == nil {
			return nil
		}
		allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidateSchedulerError))
		break
	case PodKillAction:
		// We choose to ignore the Duration property even user define it
		if in.Spec.Scheduler == nil {
			allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidatePodchaosSchedulerError))
		}
		break
	case ContainerKillAction:
		// We choose to ignore the Duration property even user define it
		if in.Spec.Scheduler == nil {
			allErrs = append(allErrs, field.Invalid(schedulerField, in.Spec.Scheduler, ValidatePodchaosSchedulerError))
		}
		break
	default:
		err := fmt.Errorf("podchaos[%s/%s] have unknown action type", in.Namespace, in.Name)
		log.Error(err, "Wrong PodChaos Action type")

		actionField := spec.Child("action")
		allErrs = append(allErrs, field.Invalid(actionField, in.Spec.Action, err.Error()))
		break
	}
	return allErrs
}

// ValidateValue validates the value
func (in *PodChaos) ValidateValue(spec *field.Path) field.ErrorList {
	return ValidateValue(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}
