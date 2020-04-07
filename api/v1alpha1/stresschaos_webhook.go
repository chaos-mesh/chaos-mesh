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
	"errors"
	"fmt"
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
	Validate(parent *field.Path, errs field.ErrorList) field.ErrorList
}

// log is for logging in this package.
var (
	stressChaosLog = ctrl.Log.WithName("stresschaos-resource")
)

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
	root := field.NewPath("stresschaos")
	errs := field.ErrorList{}
	errs = in.Spec.Validate(root, errs)
	errs = append(errs, in.ValidatePodMode(root)...)
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
func (in *StressChaos) ValidateScheduler(root *field.Path) field.ErrorList {
	panic("implement me")
}

// Validate validates the scheduler and duration
func (in *StressChaosSpec) Validate(parent *field.Path, errs field.ErrorList) field.ErrorList {
	current := parent.Child("spec")
	errs = in.Stressors.Validate(current, errs)
	if in.Duration != nil && in.Scheduler != nil {
		return errs
	} else if in.Duration == nil && in.Scheduler == nil {
		return errs
	}
	schedulerField := current.Child("scheduler")
	return append(errs, field.Invalid(schedulerField, in.Scheduler, ValidateSchedulerError))
}

// Validate validates whether the Stressors are all well defined
func (in *Stressors) Validate(parent *field.Path, errs field.ErrorList) field.ErrorList {
	current := parent.Child("stressors")
	once := false
	if in.VmStressor != nil {
		errs = in.VmStressor.Validate(current, errs)
		once = true
	}
	if in.CPUStressor != nil {
		errs = in.CPUStressor.Validate(current, errs)
		once = true
	}
	if !once {
		errs = append(errs, field.Invalid(current, in, "missing stressors"))
	}
	return errs
}

// Validate validates whether the Stressor is well defined
func (in *Stressor) Validate(parent *field.Path, errs field.ErrorList) field.ErrorList {
	if in.Workers <= 0 {
		errs = append(errs, field.Invalid(parent, in, "workers should always be positive"))
	}
	return errs
}

// Validate validates whether the VMStressor is well defined
func (in *VMStressor) Validate(parent *field.Path, errs field.ErrorList) field.ErrorList {
	current := parent.Child("vm")
	errs = in.Stressor.Validate(current, errs)
	if err := in.tryParseBytes(); err != nil {
		errs = append(errs, field.Invalid(current, in,
			fmt.Sprintf("incorrect bytes format: %s", err)))
	}
	return errs
}

func (in *VMStressor) tryParseBytes() error {
	length := len(in.Bytes)
	if length == 0 {
		in.Bytes = "100%"
		return nil
	}
	if in.Bytes[length-1] == '%' {
		percent, err := strconv.Atoi(in.Bytes[:length-1])
		if err != nil {
			return err
		}
		if percent > 100 || percent < 0 {
			return errors.New("illegal proportion")
		}
	} else {
		size, err := units.FromHumanSize(in.Bytes)
		if err != nil {
			return err
		}
		in.Bytes = fmt.Sprintf("%db", size)
	}
	return nil
}

// Validate validates whether the CPUStressor is well defined
func (in *CPUStressor) Validate(parent *field.Path, errs field.ErrorList) field.ErrorList {
	if in.Load == nil {
		in.Load = new(int)
		*in.Load = 100
		return errs
	}
	current := parent.Child("cpu")
	errs = in.Stressor.Validate(current, errs)
	if *in.Load < 0 || *in.Load > 100 {
		errs = append(errs, field.Invalid(current, in, "illegal proportion"))
	}
	if len(in.Method) == 0 {
		in.Method = CPUMethodAll
	}
	return errs
}
