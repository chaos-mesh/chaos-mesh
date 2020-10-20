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
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var podnetworkchaoslog = logf.Log.WithName("rawpodnetwork-resource")

// +kubebuilder:object:generate=false

// PodNetworkChaosHandler represents the implementation of podnetworkchaos
type PodNetworkChaosHandler interface {
	Apply(context.Context, *PodNetworkChaos) error
}

var podNetworkChaosHandler PodNetworkChaosHandler

// RegisterRawPodNetworkHandler registers handler into webhook
func RegisterRawPodNetworkHandler(newHandler PodNetworkChaosHandler) {
	podNetworkChaosHandler = newHandler
}

// SetupWebhookWithManager setup PodNetworkChaos's webhook with manager
func (in *PodNetworkChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-podnetworkchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=podnetworkchaos,verbs=create;update,versions=v1alpha1,name=mpodnetworkchaos.kb.io

var _ webhook.Defaulter = &PodNetworkChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *PodNetworkChaos) Default() {
	podnetworkchaoslog.Info("default", "name", in.Name)

	// Do nothing here
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-podnetworkchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=podnetworkchaos,versions=v1alpha1,name=vpodnetworkchaos.kb.io

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *PodNetworkChaos) ValidateCreate() error {
	// TODO: validate

	podnetworkchaoslog.Info("validate create", "name", in.Name)

	if podNetworkChaosHandler != nil {
		err := podNetworkChaosHandler.Apply(context.TODO(), in)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *PodNetworkChaos) ValidateUpdate(old runtime.Object) error {
	// TODO: validate

	podnetworkchaoslog.Info("validate update", "name", in.Name)

	if podNetworkChaosHandler != nil {
		err := podNetworkChaosHandler.Apply(context.TODO(), in)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *PodNetworkChaos) ValidateDelete() error {
	podnetworkchaoslog.Info("validate delete", "name", in.Name)

	return nil
}
