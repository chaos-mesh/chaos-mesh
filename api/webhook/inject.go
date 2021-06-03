// Copyright 2019 Chaos Mesh Authors.
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

package webhook

import (
	"context"
	"net/http"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/chaos-mesh/chaos-mesh/controllers/metrics"
	controllerCfg "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/inject"
)

var log = ctrl.Log.WithName("inject-webhook")

// +kubebuilder:webhook:path=/inject-v1-pod,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

// PodInjector is pod template config injector
type PodInjector struct {
	client        client.Client
	decoder       *admission.Decoder
	Config        *config.Config
	ControllerCfg *controllerCfg.ChaosControllerConfig
	Metrics       *metrics.ChaosCollector
}

// Handle is pod injector handler
func (v *PodInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &v1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	log.Info("Get request from pod:", "pod", pod)

	return admission.Response{
		AdmissionResponse: *inject.Inject(&req.AdmissionRequest, v.client, v.Config, v.ControllerCfg, v.Metrics),
	}
}

// InjectClient is pod injector client
func (v *PodInjector) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

// InjectDecoder is pod injector decoder
func (v *PodInjector) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
