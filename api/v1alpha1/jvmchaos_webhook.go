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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var jvmchaoslog = logf.Log.WithName("jvmchaos-resource")

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-jvmchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=jvmchaos,verbs=create;update,versions=v1alpha1,name=mjvmchaos.kb.io

var _ webhook.Defaulter = &JVMChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *JVMChaos) Default() {
	jvmchaoslog.Info("default", "name", in.Name)

	in.Spec.Selector.DefaultNamespace(in.GetNamespace())
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-jvmchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=jvmchaos,versions=v1alpha1,name=vjvmchaos.kb.io

var _ ChaosValidator = &JVMChaos{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *JVMChaos) ValidateCreate() error {
	jvmchaoslog.Info("validate create", "name", in.Name)

	return in.Validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *JVMChaos) ValidateUpdate(old runtime.Object) error {
	jvmchaoslog.Info("validate update", "name", in.Name)

	return in.Validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *JVMChaos) ValidateDelete() error {
	jvmchaoslog.Info("validate delete", "name", in.Name)

	// Nothing to do?
	return nil
}

// Validate validates chaos object
func (in *JVMChaos) Validate() error {
	// TODO: need some validate process
	return nil
}

// ValidateScheduler validates the scheduler and duration
func (in *JVMChaos) ValidateScheduler(spec *field.Path) field.ErrorList {
	return ValidateScheduler(in, spec)
}

// ValidatePodMode validates the value with podmode
func (in *JVMChaos) ValidatePodMode(spec *field.Path) field.ErrorList {
	return ValidatePodMode(in.Spec.Value, in.Spec.Mode, spec.Child("value"))
}

// SelectSpec returns the selector config for authority validate
func (in *JVMChaos) GetSelectSpec() []SelectSpec {
	return []SelectSpec{&in.Spec}
}
