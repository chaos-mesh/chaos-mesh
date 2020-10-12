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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var dnschaoslog = logf.Log.WithName("dnschaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-dnschaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=dnschaos,verbs=create;update,versions=v1alpha1,name=mdnschaos.kb.io

var _ webhook.Defaulter = &DNSChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *DNSChaos) Default() {
	dnschaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-dnschaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=dnschaos,versions=v1alpha1,name=vdnschaos.kb.io

var _ ChaosValidator = &DNSChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *DNSChaos) ValidateCreate() error {
	dnschaoslog.Info("validate create", "name", in.Name)
	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *DNSChaos) ValidateUpdate(old runtime.Object) error {
	dnschaoslog.Info("validate update", "name", in.Name)
	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *DNSChaos) ValidateDelete() error {
	dnschaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *DNSChaos) Validate() error {
	specField := field.NewPath("spec")
	allErrs := in.ValidateScheduler(specField)
	allErrs = append(allErrs, in.ValidatePodMode(specField)...)

	if len(allErrs) > 0 {
		return fmt.Errorf(allErrs.ToAggregate().Error())
	}

	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *DNSChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	return ValidateScheduler(in, spec)
}

// ValidatePodMode validates the value with podmode
func (in *DNSChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	return ValidatePodMode(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}
