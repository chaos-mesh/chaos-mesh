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

package webhook

import (
	"context"
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/pingcap/chaos-mesh/pkg/webhook/validate"
)

//// +kubebuilder:webhook:path=/validate-v1alpha1-chaos,validating=true,failurePolicy=fail,groups="pingcap.com",resources=iochaos;podchaos;networkchaos;timechaos,verbs=create;update,versions=v1,name=chaos.validate

var validatelog = ctrl.Log.WithName("validate-webhook")

// ChaosValidator
type ChaosValidator struct {
}

// Handle handle the requests from validation admission requests
func (v *ChaosValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	name := req.Name
	namespace := req.Namespace
	kind := req.Kind
	validatelog.Info(fmt.Sprintf("receive validation req for obj[%s/%s/%s]", kind.Kind, namespace, name))
	return admission.Response{
		AdmissionResponse: *validate.ValidateChaos(&req.AdmissionRequest, kind.Kind),
	}
}
