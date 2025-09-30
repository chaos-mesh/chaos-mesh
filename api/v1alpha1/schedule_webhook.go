// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var schedulelog = logf.Log.WithName("schedule-resource")

var _ webhook.CustomDefaulter = &Schedule{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *Schedule) Default(_ context.Context, _ runtime.Object) error {
	schedulelog.Info("default", "name", in.Name)
	in.Spec.ConcurrencyPolicy.Default()
	return nil
}

func (in *ConcurrencyPolicy) Default() {
	if *in == "" {
		*in = ForbidConcurrent
	}
}

var _ webhook.CustomValidator = &Schedule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *Schedule) ValidateCreate(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	schedule, ok := obj.(*Schedule)
	if !ok {
		return nil, errors.Errorf("expected type *Schedule, got %T", obj)
	}

	schedulelog.Info("validate create", "name", schedule.Name)

	return in.Validate(schedule)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *Schedule) ValidateUpdate(_ context.Context, _, newObj runtime.Object) (admission.Warnings, error) {
	newSchedule, ok := newObj.(*Schedule)
	if !ok {
		return nil, errors.Errorf("expected type *Schedule, got %T", newObj)
	}

	schedulelog.Info("validate update", "name", newSchedule.Name)

	return in.Validate(newSchedule)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *Schedule) ValidateDelete(_ context.Context, obj runtime.Object) (admission.Warnings, error) {
	schedule, ok := obj.(*Schedule)
	if !ok {
		return nil, errors.Errorf("expected type *Schedule, got %T", obj)
	}

	schedulelog.Info("validate delete", "name", schedule.Name)

	return nil, nil
}

// Validate validates chaos object
func (in *Schedule) Validate(obj *Schedule) ([]string, error) {
	allErrs := in.Spec.Validate(&obj.Spec)

	if len(allErrs) > 0 {
		return nil, errors.New(allErrs.ToAggregate().Error())
	}

	return nil, nil
}

func (in *ScheduleSpec) Validate(spec *ScheduleSpec) field.ErrorList {
	specField := field.NewPath("spec")
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, spec.validateSchedule(specField.Child("schedule"))...)
	allErrs = append(allErrs, spec.validateChaos(specField)...)

	return allErrs
}

// validateSchedule validates the cron
func (in *ScheduleSpec) validateSchedule(schedule *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	_, err := StandardCronParser.Parse(in.Schedule)
	if err != nil {
		allErrs = append(
			allErrs,
			field.Invalid(
				schedule,
				in.Schedule,
				fmt.Sprintf("parse schedule field error:%s", err),
			),
		)
	}

	return allErrs
}

// validateChaos validates the chaos
func (in *ScheduleSpec) validateChaos(chaos *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if in.Type != ScheduleTypeWorkflow {
		allErrs = append(allErrs, in.EmbedChaos.Validate(chaos, string(in.Type))...)
	}

	return allErrs
}
