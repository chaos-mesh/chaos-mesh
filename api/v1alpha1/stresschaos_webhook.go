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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:object:generate=false

// Validateable defines how a resource is validated
type Validateable interface {
	Validate(parent *field.Path) field.ErrorList
}

// log is for logging in this package.
var (
	stressChaosLog = ctrl.Log.WithName("stresschaos-resource")
)

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-stresschaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=stresschaos,verbs=create;update,versions=v1alpha1,name=mstresschaos.kb.io

var _ webhook.Defaulter = &StressChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *StressChaos) Default() {
	stressChaosLog.Info("default", "name", in.Name)
	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-stresschaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=stresschaos,versions=v1alpha1,name=vstresschaos.kb.io

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
	root := field.NewPath("stresschaos")
	errs := in.Spec.Validate(root)
	errs = append(errs, in.ValidatePodMode(root)...)
	errs = append(errs, in.ValidateScheduler(root.Child("spec"))...)
	if len(errs) > 0 {
		return fmt.Errorf(errs.ToAggregate().Error())
	}
	return nil
}

// ValidatePodMode validates the value with podmode
func (in *StressChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	return ValidatePodMode(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}

// ValidateScheduler validates whether scheduler is well defined
func (in *StressChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	return ValidateScheduler(in, spec)
}

// Validate validates the scheduler and duration
func (in *StressChaosSpec) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	current := parent.Child("spec")
	if len(in.StressngStressors) == 0 && in.Stressors == nil {
		errs = append(errs, field.Invalid(current, in, "missing stressors"))
	} else if in.Stressors != nil {
		errs = append(errs, in.Stressors.Validate(current)...)
	}
	return errs
}

// Validate validates whether the Stressors are all well defined
func (in *Stressors) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	current := parent.Child("stressors")
	once := false
	if in.MemoryStressor != nil {
		errs = append(errs, in.MemoryStressor.Validate(current)...)
		once = true
	}
	if in.CPUStressor != nil {
		errs = append(errs, in.CPUStressor.Validate(current)...)
		once = true
	}
	if !once {
		errs = append(errs, field.Invalid(current, in, "missing stressors"))
	}
	return errs
}

// Validate validates whether the Stressor is well defined
func (in *Stressor) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if in.Workers <= 0 {
		errs = append(errs, field.Invalid(parent, in, "workers should always be positive"))
	}
	return errs
}

// Validate validates whether the MemoryStressor is well defined
func (in *MemoryStressor) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	current := parent.Child("vm")
	errs = append(errs, in.Stressor.Validate(current)...)
	return errs
}

// Validate validates whether the CPUStressor is well defined
func (in *CPUStressor) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	current := parent.Child("cpu")
	errs = append(errs, in.Stressor.Validate(current)...)
	if in.Load != nil && (*in.Load < 0 || *in.Load > 100) {
		errs = append(errs, field.Invalid(current, in, "illegal proportion"))
	}
	return errs
}
