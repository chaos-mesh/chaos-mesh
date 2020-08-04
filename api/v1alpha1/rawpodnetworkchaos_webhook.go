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
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var rawpodnetworkchaoslog = logf.Log.WithName("rawpodnetwork-resource")

type RawPodNetworkChaosHandler interface {
	Apply(context.Context, *RawPodNetworkChaos) error
}

var handler RawPodNetworkChaosHandler

func RegisterRawPodNetworkHandler(newHandler RawPodNetworkChaosHandler) {
	handler = newHandler
}

// SetupWebhookWithManager setup RawPodNetworkChaos's webhook with manager
func (in *RawPodNetworkChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-rawpodnetworkchaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=rawpodnetworkchaos,verbs=create;update,versions=v1alpha1,name=mrawpodnetworkchaos.kb.io

var _ webhook.Defaulter = &RawPodNetworkChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *RawPodNetworkChaos) Default() {
	rawpodnetworkchaoslog.Info("default", "name", in.Name)

	// Do nothing here
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-chaos-mesh-org-v1alpha1-rawpodnetworkchaos,mutating=false,failurePolicy=fail,groups=chaos-mesh.org,resources=rawpodnetworkchaos,versions=v1alpha1,name=vrawpodnetworkchaos.kb.io

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *RawPodNetworkChaos) ValidateCreate() error {
	rawpodnetworkchaoslog.Info("validate create", "name", in.Name)

	if handler != nil {
		handler.Apply(context.TODO(), in)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *RawPodNetworkChaos) ValidateUpdate(old runtime.Object) error {
	rawpodnetworkchaoslog.Info("validate update", "name", in.Name)

	if handler != nil {
		handler.Apply(context.TODO(), in)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *RawPodNetworkChaos) ValidateDelete() error {
	rawpodnetworkchaoslog.Info("validate delete", "name", in.Name)

	return nil
}
