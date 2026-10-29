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

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type webhookTestManager struct {
	ctrl.Manager
	scheme *runtime.Scheme
	server webhook.Server
}

func (m webhookTestManager) GetScheme() *runtime.Scheme       { return m.scheme }
func (m webhookTestManager) GetConfig() *rest.Config          { return &rest.Config{} }
func (m webhookTestManager) GetWebhookServer() webhook.Server { return m.server }

func TestRegisteredWebhooks(t *testing.T) {
	cases := []struct {
		kind         string
		object       v1alpha1.WebhookObject
		spec         string
		invalidSpec  string
		defaultPath  string
		defaultValue string
	}{
		{"PodChaos", &v1alpha1.PodChaos{}, `{"action":"pod-kill","mode":"all","selector":{}}`, `{"action":"container-kill","mode":"all","selector":{}}`, "/spec/selector/namespaces", `["default"]`},
		{"PhysicalMachine", &v1alpha1.PhysicalMachine{}, `{"address":"http://127.0.0.1:31767"}`, `{}`, "", ""},
		{"Schedule", &v1alpha1.Schedule{}, `{"schedule":"@every 5s","type":"Workflow","workflow":{}}`, `{}`, "", ""},
		{"Workflow", &v1alpha1.Workflow{}, `{"entry":"pause","templates":[{"name":"pause","templateType":"Suspend","deadline":"1s"}]}`, `{}`, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := v1alpha1.AddToScheme(scheme); err != nil {
				t.Fatal(err)
			}
			server := webhook.NewServer(webhook.Options{})
			if err := registerWebhook(webhookTestManager{scheme: scheme, server: server}, tc.object); err != nil {
				t.Fatal(err)
			}
			object := []byte(fmt.Sprintf(`{"apiVersion":"chaos-mesh.org/v1alpha1","kind":%q,"metadata":{"name":"example","namespace":"default"},"spec":%s}`, tc.kind, tc.spec))
			request := func(endpoint string, op admissionv1.Operation, raw []byte) *admissionv1.AdmissionResponse {
				t.Helper()
				review := admissionv1.AdmissionReview{
					TypeMeta: metav1.TypeMeta{APIVersion: "admission.k8s.io/v1", Kind: "AdmissionReview"},
					Request:  &admissionv1.AdmissionRequest{UID: "test-request", Operation: op, Namespace: "default", Object: runtime.RawExtension{Raw: raw}, OldObject: runtime.RawExtension{Raw: object}},
				}
				body, err := json.Marshal(review)
				if err != nil {
					t.Fatal(err)
				}
				req := httptest.NewRequest(http.MethodPost, "/"+endpoint+"-chaos-mesh-org-v1alpha1-"+strings.ToLower(tc.kind), bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				server.WebhookMux().ServeHTTP(recorder, req)
				if recorder.Code != http.StatusOK {
					t.Fatalf("webhook HTTP status %d: %s", recorder.Code, recorder.Body.String())
				}
				var result admissionv1.AdmissionReview
				if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
					t.Fatal(err)
				}
				if result.Response == nil {
					t.Fatal("missing admission response")
				}
				return result.Response
			}
			for _, op := range []admissionv1.Operation{admissionv1.Create, admissionv1.Update, admissionv1.Delete} {
				t.Run(string(op), func(t *testing.T) {
					response := request("validate", op, object)
					if !response.Allowed {
						t.Fatalf("valid object rejected: %+v", response.Result)
					}
				})
			}
			invalid := []byte(fmt.Sprintf(`{"apiVersion":"chaos-mesh.org/v1alpha1","kind":%q,"metadata":{"name":"example"},"spec":%s}`, tc.kind, tc.invalidSpec))
			response := request("validate", admissionv1.Create, invalid)
			if response.Allowed || response.Result == nil || response.Result.Code >= http.StatusInternalServerError {
				t.Fatalf("expected validation rejection, got %+v", response)
			}
			response = request("mutate", admissionv1.Create, object)
			if !response.Allowed {
				t.Fatalf("defaulting failed: %+v", response.Result)
			}
			if tc.defaultPath != "" {
				var patches []struct {
					Path  string          `json:"path"`
					Value json.RawMessage `json:"value"`
				}
				if err := json.Unmarshal(response.Patch, &patches); err != nil {
					t.Fatal(err)
				}
				found := false
				for _, patch := range patches {
					if patch.Path == tc.defaultPath && string(patch.Value) == tc.defaultValue {
						found = true
					}
				}
				if !found {
					t.Fatalf("patch %s does not set %s to %s", response.Patch, tc.defaultPath, tc.defaultValue)
				}
			}
		})
	}
}
