// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestKubernetesInputErrorsAreBadRequests(t *testing.T) {
	for _, failure := range []error{
		apierrors.NewBadRequest("malformed request"),
		apierrors.NewInvalid(schema.GroupKind{Group: "chaos-mesh.org", Kind: "Schedule"}, "sample", field.ErrorList{field.NotSupported(field.NewPath("spec", "concurrencyPolicy"), "", []string{"Allow", "Forbid"})}),
	} {
		response := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(response)
		SetAPImachineryError(ctx, failure)
		var body APIError
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if response.Code != http.StatusBadRequest || body.Code != 400 || body.Type != ErrBadRequest.FullName() || !ctx.IsAborted() {
			t.Fatalf("input error %v returned status %d body %+v", failure, response.Code, body)
		}
	}
}

func TestAdmissionRejectionErrors(t *testing.T) {
	denied := &apierrors.StatusError{ErrStatus: metav1.Status{
		Status: metav1.StatusFailure, Reason: metav1.StatusReasonForbidden, Code: 403,
		Message: `admission webhook "vnetworkchaos.kb.io" denied the request: Spec.PodSelector.value: Invalid value: "not-a-number"`,
	}}
	for _, scenario := range []struct {
		name string
		err  error
		code int
		kind string
	}{
		{"webhook validation", denied, 400, ErrBadRequest.FullName()},
		{"wrapped validation", fmt.Errorf("create experiment: %w", denied), 400, ErrBadRequest.FullName()},
		{"validation mentioning namespace", &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonForbidden, Code: 403, Message: `admission webhook "vnetworkchaos.kb.io" denied the request: invalid selector in the namespace`}}, 400, ErrBadRequest.FullName()},
		{"namespace permission", apierrors.NewForbidden(schema.GroupResource{Group: "chaos-mesh.org", Resource: "networkchaos"}, "sample", fmt.Errorf("cannot create resource in the namespace target")), 401, ErrNoNamespacePrivilege.FullName()},
		{"cluster permission", apierrors.NewForbidden(schema.GroupResource{Group: "chaos-mesh.org", Resource: "networkchaos"}, "sample", fmt.Errorf("cannot list resource at the cluster scope")), 401, ErrNoClusterPrivilege.FullName()},
		{"webhook unavailable", apierrors.NewInternalError(fmt.Errorf(`failed calling webhook "vnetworkchaos.kb.io": connection refused`)), 500, ErrInternalServer.FullName()},
		{"other forbidden", &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonForbidden, Code: 403, Message: "operation forbidden"}}, 500, ErrInternalServer.FullName()},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)
			SetAPImachineryError(ctx, scenario.err)
			var body APIError
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != scenario.code || body.Code != scenario.code || body.Type != scenario.kind || !ctx.IsAborted() {
				t.Fatalf("status %d body %+v", response.Code, body)
			}
		})
	}
}
