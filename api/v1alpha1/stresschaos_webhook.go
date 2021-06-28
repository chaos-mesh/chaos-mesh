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
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/docker/go-units"
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
	in.Spec.Default()
}

func (in *StressChaosSpec) Default() {

}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-stresschaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=stresschaos,versions=v1alpha1,name=vstresschaos.kb.io

var _ webhook.Validator = &StressChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateCreate() error {
	stressChaosLog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *StressChaos) ValidateUpdate(old runtime.Object) error {
	stressChaosLog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*StressChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
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
	errs := in.Spec.Validate()
	if len(errs) > 0 {
		return fmt.Errorf(errs.ToAggregate().Error())
	}
	return nil
}

// Validate validates the scheduler and duration
func (in *StressChaosSpec) Validate() field.ErrorList {
	errs := field.ErrorList{}
	specField := field.NewPath("spec")
	var allErrs field.ErrorList
	if len(in.StressngStressors) == 0 && in.Stressors == nil {
		allErrs = append(errs, field.Invalid(specField, in, "missing stressors"))
	} else if in.Stressors != nil {
		allErrs = append(errs, in.Stressors.Validate(specField)...)
	}
	allErrs = append(allErrs, validateDuration(in, specField)...)
	return allErrs
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
	if err := in.tryParseBytes(); err != nil {
		errs = append(errs, field.Invalid(current, in,
			fmt.Sprintf("incorrect bytes format: %s", err)))
	}
	return errs
}

func (in *MemoryStressor) tryParseBytes() error {
	length := len(in.Size)
	if length == 0 {
		return nil
	}
	if in.Size[length-1] == '%' {
		percent, err := strconv.Atoi(in.Size[:length-1])
		if err != nil {
			return err
		}
		if percent > 100 || percent < 0 {
			return errors.New("illegal proportion")
		}
	} else {
		size, err := units.FromHumanSize(in.Size)
		if err != nil {
			return err
		}
		in.Size = fmt.Sprintf("%db", size)
	}
	return nil
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
