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
	"encoding/json"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var podiochaoslog = logf.Log.WithName("rawpodio-resource")

// +kubebuilder:object:generate=false

// PodIoChaosHandler represents the implementation of podiochaos
type PodIoChaosHandler interface {
	Apply(context.Context, *PodIoChaos) error
}

var podIoChaosHandler PodIoChaosHandler

// RegisterPodIoHandler registers handler into webhook
func RegisterPodIoHandler(newHandler PodIoChaosHandler) {
	podIoChaosHandler = newHandler
}

// SetupWebhookWithManager setup PodIoChaos's webhook with manager
func (in *PodIoChaos) SetupWebhookWithManager(mgr ctrl.Manager) error {
	mgr.GetWebhookServer().
		Register("/mutate-chaos-mesh-org-v1alpha1-podiochaos", &webhook.Admission{Handler: &PodIoChaosWebhookRunner{}})
	return nil
}

// +kubebuilder:webhook:path=/mutate-chaos-mesh-org-v1alpha1-podiochaos,mutating=true,failurePolicy=fail,groups=chaos-mesh.org,resources=podiochaos,verbs=create;update,versions=v1alpha1,name=mpodiochaos.kb.io

// +kubebuilder:object:generate=false

// PodIoChaosWebhookRunner runs webhook for podiochaos
type PodIoChaosWebhookRunner struct {
	decoder *admission.Decoder
}

// Handle will run poiochaoshandler for this resource
func (r *PodIoChaosWebhookRunner) Handle(ctx context.Context, req admission.Request) admission.Response {
	chaos := &PodIoChaos{}
	err := r.decoder.Decode(req, chaos)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if podIoChaosHandler != nil {
		err = podIoChaosHandler.Apply(ctx, chaos)
		if err != nil {
			// TODO: refine the http status code
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	// mutate the fields in pod

	marshaledPodIoChaos, err := json.Marshal(chaos)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPodIoChaos)
}

// InjectDecoder injects decoder into webhook runner
func (r *PodIoChaosWebhookRunner) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}
