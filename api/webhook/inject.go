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
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/injector"
	"net/http"

	"github.com/chaos-mesh/chaos-mesh/controllers/metrics"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/inject"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	v1 "k8s.io/api/core/v1"
)

var log = ctrl.Log.WithName("inject-webhook")

// +kubebuilder:webhook:path=/inject-v1-pod,mutating=false,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=vpod.kb.io

type PodInjector struct {
	client  client.Client
	decoder *admission.Decoder
	Config  *config.Config
	Metrics *metrics.ChaosCollector
}

func (v *PodInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &v1.Pod{}

	err := v.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	var log = ctrl.Log.WithName("inject-webhook")
	log.Info("Get request from pod:", "pod", pod)
	injectResponse := *inject.Inject(&req.AdmissionRequest, v.client, v.Config, v.Metrics)
	if injectResponse.PatchType == nil {
		log.Info("0000000000000")
		return admission.Response{
			AdmissionResponse: *injector.Inject(&req.AdmissionRequest),
		}
	}
	log.Info("aaaaaaaaaaaaaaaaaa")
	return admission.Response{
		AdmissionResponse: injectResponse,
	}
}

func (v *PodInjector) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

func (v *PodInjector) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}
