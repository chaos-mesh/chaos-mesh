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
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	// DefaultJitter defines default value for jitter
	DefaultJitter = "0ms"

	// DefaultCorrelation defines default value for correlation
	DefaultCorrelation = "0"
)

// log is for logging in this package.
var networkchaoslog = logf.Log.WithName("networkchaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-networkchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=networkchaos,verbs=create;update,versions=v1alpha1,name=mnetworkchaos.kb.io

var _ webhook.Defaulter = &NetworkChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *NetworkChaos) Default() {
	networkchaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
	// the target's namespace selector
	if in.Spec.Target != nil {
		in.Spec.Target.Selector.DefaultNamespace(in.GetNamespace())
	}

	in.Spec.Default()
}

func (in *NetworkChaosSpec) Default() {
	// set default direction
	if in.Direction == "" {
		in.Direction = To
	}

	in.DefaultDelay()
	in.DefaultLoss()
	in.DefaultDuplicate()
	in.DefaultCorrupt()
}

// DefaultDelay set the default value if Jitter or Correlation is not set
func (in *NetworkChaosSpec) DefaultDelay() {
	if in.Delay != nil {
		if in.Delay.Jitter == "" {
			in.Delay.Jitter = DefaultJitter
		}
		if in.Delay.Correlation == "" {
			in.Delay.Correlation = DefaultCorrelation
		}

		if in.Delay.Reorder != nil {
			if in.Delay.Reorder.Correlation == "" {
				in.Delay.Reorder.Correlation = DefaultCorrelation
			}
		}
	}
}

func (in *NetworkChaosSpec) DefaultLoss() {
	if in.Loss != nil {
		if in.Loss.Correlation == "" {
			in.Loss.Correlation = DefaultCorrelation
		}
	}
}

func (in *NetworkChaosSpec) DefaultDuplicate() {
	if in.Duplicate != nil {
		if in.Duplicate.Correlation == "" {
			in.Duplicate.Correlation = DefaultCorrelation
		}
	}
}

func (in *NetworkChaosSpec) DefaultCorrupt() {
	if in.Corrupt != nil {
		if in.Corrupt.Correlation == "" {
			in.Corrupt.Correlation = DefaultCorrelation
		}
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-networkchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=networkchaos,versions=v1alpha1,name=vnetworkchaos.kb.io

var _ webhook.Validator = &NetworkChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *NetworkChaos) ValidateCreate() error {
	networkchaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *NetworkChaos) ValidateUpdate(old runtime.Object) error {
	networkchaoslog.Info("validate update", "name", in.Name)
	if !reflect.DeepEqual(in.Spec, old.(*NetworkChaos).Spec) {
		return ErrCanNotUpdateChaos
	}
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *NetworkChaos) ValidateDelete() error {
	networkchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *NetworkChaos) Validate() error {

	allErrs := in.Spec.Validate()

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}
	return nil
}

func (in *NetworkChaosSpec) Validate() field.ErrorList {
	specField := field.NewPath("spec")
	var allErrs field.ErrorList

	allErrs = append(allErrs, validateDuration(in, specField)...)
	allErrs = append(allErrs, in.validateTargets(specField.Child("target"))...)
	if in.Delay != nil {
		allErrs = append(allErrs, in.Delay.validateDelay(specField.Child("delay"))...)
	}
	if in.Loss != nil {
		allErrs = append(allErrs, in.Loss.validateLoss(specField.Child("loss"))...)
	}
	if in.Duplicate != nil {
		allErrs = append(allErrs, in.Duplicate.validateDuplicate(specField.Child("duplicate"))...)
	}
	if in.Corrupt != nil {
		allErrs = append(allErrs, in.Corrupt.validateCorrupt(specField.Child("corrupt"))...)
	}
	if in.Bandwidth != nil {
		allErrs = append(allErrs, in.Bandwidth.validateBandwidth(specField.Child("bandwidth"))...)
	}
	if in.Target != nil {
		allErrs = append(allErrs, in.validateTargetPodSelector(specField.Child("target"))...)
	}

	return allErrs
}

// validateDelay validates the delay
func (in *DelaySpec) validateDelay(delay *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, err := time.ParseDuration(in.Latency)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(delay.Child("latency"), in.Latency,
				fmt.Sprintf("parse latency field error:%s", err)))
	}
	_, err = time.ParseDuration(in.Jitter)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(delay.Child("jitter"), in.Jitter,
				fmt.Sprintf("parse jitter field error:%s", err)))
	}

	_, err = strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(delay.Child("correlation"), in.Correlation,
				fmt.Sprintf("parse correlation field error:%s", err)))
	}

	if in.Reorder != nil {
		allErrs = append(allErrs, in.Reorder.validateReorder(delay.Child("reorder"))...)
	}
	return allErrs
}

// validateReorder validates the reorder
func (in *ReorderSpec) validateReorder(reorder *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, err := strconv.ParseFloat(in.Reorder, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(reorder.Child("reorder"), in.Reorder,
				fmt.Sprintf("parse reorder field error:%s", err)))
	}

	_, err = strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(reorder.Child("correlation"), in.Correlation,
				fmt.Sprintf("parse correlation field error:%s", err)))
	}
	return allErrs
}

// validateLoss validates the loss
func (in *LossSpec) validateLoss(loss *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	_, err := strconv.ParseFloat(in.Loss, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(loss.Child("loss"), in.Loss,
				fmt.Sprintf("parse loss field error:%s", err)))
	}

	_, err = strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(loss.Child("correlation"), in.Correlation,
				fmt.Sprintf("parse correlation field error:%s", err)))
	}

	return allErrs
}

// validateDuplicate validates the duplicate
func (in *DuplicateSpec) validateDuplicate(duplicate *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, err := strconv.ParseFloat(in.Duplicate, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(duplicate.Child("duplicate"), in.Duplicate,
				fmt.Sprintf("parse duplicate field error:%s", err)))
	}

	_, err = strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(duplicate.Child("correlation"), in.Correlation,
				fmt.Sprintf("parse correlation field error:%s", err)))
	}
	return allErrs
}

// validateCorrupt validates the corrupt
func (in *CorruptSpec) validateCorrupt(corrupt *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, err := strconv.ParseFloat(in.Corrupt, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(corrupt.Child("corrupt"), in.Corrupt,
				fmt.Sprintf("parse corrupt field error:%s", err)))
	}

	_, err = strconv.ParseFloat(in.Correlation, 32)
	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(corrupt.Child("correlation"), in.Correlation,
				fmt.Sprintf("parse correlation field error:%s", err)))
	}
	return allErrs
}

// validateBandwidth validates the bandwidth
func (in *BandwidthSpec) validateBandwidth(bandwidth *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	_, err := ConvertUnitToBytes(in.Rate)

	if err != nil {
		allErrs = append(allErrs,
			field.Invalid(bandwidth.Child("rate"), in.Rate,
				fmt.Sprintf("parse rate field error:%s", err)))
	}
	return allErrs
}

func ConvertUnitToBytes(nu string) (uint64, error) {
	// normalize input
	s := strings.ToLower(strings.TrimSpace(nu))

	for i, u := range []string{"tbps", "gbps", "mbps", "kbps", "bps"} {
		if strings.HasSuffix(s, u) {
			ts := strings.TrimSuffix(s, u)
			s := strings.TrimSpace(ts)

			n, err := strconv.ParseUint(s, 10, 64)

			if err != nil {
				return 0, err
			}

			// convert unit to bytes
			for j := 4 - i; j > 0; j-- {
				n = n * 1024
			}

			return n, nil
		}
	}

	return 0, errors.New("invalid unit")
}

// validateTarget validates the target
func (in *NetworkChaosSpec) validateTargetPodSelector(target *field.Path) field.ErrorList {
	modes := []PodMode{OnePodMode, AllPodMode, FixedPodMode, FixedPercentPodMode, RandomMaxPercentPodMode}

	for _, mode := range modes {
		if in.Target.Mode == mode {
			return validatePodSelector(in.Target.Value, in.Target.Mode, target.Child("value"))
		}
	}

	return field.ErrorList{field.Invalid(target.Child("mode"), in.Target.Mode,
		fmt.Sprintf("mode %s not supported", in.Target.Mode))}
}

// ValidateTargets validates externalTargets and Targets
func (in *NetworkChaosSpec) validateTargets(target *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if in.Action == PartitionAction {
		return nil
	}

	if (in.Direction == From || in.Direction == Both) &&
		in.ExternalTargets != nil && in.Action != PartitionAction {
		allErrs = append(allErrs,
			field.Invalid(target.Child("direction"), in.Direction,
				"external targets cannot be used with `from` and `both` direction in netem action yet"))
	}

	if (in.Direction == From || in.Direction == Both) && in.Target == nil {
		if in.Action != PartitionAction {
			allErrs = append(allErrs,
				field.Invalid(target.Child("direction"), in.Direction,
					"`from` and `both` direction cannot be used when targets is empty in netem action"))
		} else if in.ExternalTargets == nil {
			allErrs = append(allErrs,
				field.Invalid(target.Child("direction"), in.Direction,
					"`from` and `both` direction cannot be used when targets and external targets are both empty"))
		}
	}

	// TODO: validate externalTargets are in ip or domain form

	return allErrs
}
