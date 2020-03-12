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
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var podchaoslog = logf.Log.WithName("podchaos-resource")

// SetupWebhookWithManager setup PodChaos's webhook with manager
func (r *PodChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-pingcap-com-v1alpha1-podchaos,mutating=true,failurePolicy=fail,groups=pingcap.com,resources=podchaos,verbs=create;update,versions=v1alpha1,name=mpodchaos.kb.io

var _ webhook.Defaulter = &PodChaos{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *PodChaos) Default() {
	podchaoslog.Info("default", "name", r.Name)

	r.Spec.Selector.DefaultNamespace(r.GetNamespace())
}
