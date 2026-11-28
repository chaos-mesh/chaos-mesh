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
//

package webhook

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	authnv1 "k8s.io/api/authentication/v1"
	authzv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestSarVerb(t *testing.T) {
	cases := []struct {
		operation admissionv1.Operation
		want      string
	}{
		{admissionv1.Create, "create"},
		{admissionv1.Update, "update"},
		{admissionv1.Delete, "delete"},
	}

	for _, c := range cases {
		got := sarVerb(c.operation)
		if got != c.want {
			t.Errorf("sarVerb(%q) = %q, want %q", c.operation, got, c.want)
		}
	}
}

func TestAuthValidatorReviewsTheAffectedScope(t *testing.T) {
	const reviewPath = "/apis/authorization.k8s.io/v1/subjectaccessreviews"
	for _, example := range []struct {
		name       string
		namespaces []string
		operation  admissionv1.Operation
		allow      bool
		httpError  bool
		wantCode   int32
	}{
		{"accepts each selected namespace", []string{"app", "jobs"}, admissionv1.Update, true, false, http.StatusOK},
		{"denies a namespaced create", []string{"app"}, admissionv1.Create, false, false, http.StatusForbidden},
		{"accepts a cluster-wide create", nil, admissionv1.Create, true, false, http.StatusOK},
		{"denies a cluster-wide update", nil, admissionv1.Update, false, false, http.StatusForbidden},
		{"returns namespace review failures", []string{"app"}, admissionv1.Create, false, true, http.StatusBadRequest},
		{"returns cluster review failures", nil, admissionv1.Create, false, true, http.StatusBadRequest},
	} {
		t.Run(example.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			requests := make(chan authzv1.SubjectAccessReviewSpec, 3)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var review authzv1.SubjectAccessReview
				if r.Method != http.MethodPost || r.URL.Path != reviewPath || json.NewDecoder(r.Body).Decode(&review) != nil {
					http.Error(w, "unexpected review request", http.StatusBadRequest)
					return
				}
				select {
				case requests <- review.Spec:
				default:
					http.Error(w, "unexpected repeated review", http.StatusBadRequest)
					return
				}
				if example.httpError {
					http.Error(w, "review service unavailable", http.StatusServiceUnavailable)
					return
				}
				review.Status.Allowed = example.allow
				review.TypeMeta = metav1.TypeMeta{APIVersion: "authorization.k8s.io/v1", Kind: "SubjectAccessReview"}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(&review)
			}))
			t.Cleanup(server.Close)
			scheme := runtime.NewScheme()
			g.Expect(v1alpha1.AddToScheme(scheme)).To(gomega.Succeed())
			config := &rest.Config{Host: server.URL, Timeout: 5 * time.Second, ContentConfig: rest.ContentConfig{ContentType: runtime.ContentTypeJSON}}
			validator := NewAuthValidator(true, authorizationv1.NewForConfigOrDie(config), scheme, true, "", false, logr.Discard())

			obj := &v1alpha1.PodChaos{
				TypeMeta: metav1.TypeMeta{APIVersion: v1alpha1.SchemeGroupVersion.String(), Kind: v1alpha1.KindPodChaos},
				Spec: v1alpha1.PodChaosSpec{ContainerSelector: v1alpha1.ContainerSelector{PodSelector: v1alpha1.PodSelector{
					Selector: v1alpha1.PodSelectorSpec{GenericSelectorSpec: v1alpha1.GenericSelectorSpec{Namespaces: example.namespaces}},
				}}},
			}
			raw, err := json.Marshal(obj)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			req := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Kind:      metav1.GroupVersionKind{Group: v1alpha1.SchemeGroupVersion.Group, Version: v1alpha1.SchemeGroupVersion.Version, Kind: v1alpha1.KindPodChaos},
				Operation: example.operation, Object: runtime.RawExtension{Raw: raw},
				UserInfo: authnv1.UserInfo{Username: "operator", UID: "fixture-id", Groups: []string{"team"},
					Extra: map[string]authnv1.ExtraValue{"reason": {"testing"}}},
			}}
			result := validator.Handle(t.Context(), req)
			g.Expect(result.Result).NotTo(gomega.BeNil())
			g.Expect(result.Allowed).To(gomega.Equal(example.allow), "%s", result.Result.Message)
			g.Expect(result.Result.Code).To(gomega.Equal(example.wantCode), "%s", result.Result.Message)

			wantNamespaces := example.namespaces
			if len(wantNamespaces) == 0 {
				wantNamespaces = []string{""}
			}
			g.Expect(requests).To(gomega.HaveLen(len(wantNamespaces)))
			var namespaces []string
			for range wantNamespaces {
				sent := <-requests
				g.Expect(sent.User).To(gomega.Equal("operator"))
				g.Expect(sent.UID).To(gomega.Equal("fixture-id"))
				g.Expect(sent.Groups).To(gomega.Equal([]string{"team"}))
				g.Expect(sent.Extra).To(gomega.Equal(map[string]authzv1.ExtraValue{"reason": {"testing"}}))
				g.Expect(sent.ResourceAttributes).NotTo(gomega.BeNil())
				g.Expect(sent.ResourceAttributes.Group).To(gomega.Equal("chaos-mesh.org"))
				g.Expect(sent.ResourceAttributes.Resource).To(gomega.Equal("podchaos"))
				g.Expect(sent.ResourceAttributes.Verb).To(gomega.Equal(sarVerb(example.operation)))
				namespaces = append(namespaces, sent.ResourceAttributes.Namespace)
			}
			g.Expect(namespaces).To(gomega.ConsistOf(wantNamespaces))
		})
	}
}

func TestAuthValidatorHandlesExemptAndMalformedRequests(t *testing.T) {
	g := gomega.NewWithT(t)
	scheme := runtime.NewScheme()
	g.Expect(v1alpha1.AddToScheme(scheme)).To(gomega.Succeed())
	disabled := NewAuthValidator(false, nil, scheme, true, "", false, logr.Discard())
	enabled := NewAuthValidator(true, nil, scheme, true, "", false, logr.Discard())
	request := admission.Request{}
	g.Expect(disabled.Handle(t.Context(), request).Allowed).To(gomega.BeTrue())

	request.Kind.Kind = v1alpha1.KindPodIOChaos
	g.Expect(enabled.Handle(t.Context(), request).Allowed).To(gomega.BeTrue())

	for _, kind := range []string{"UnknownKind", v1alpha1.KindPodChaos} {
		request.Kind.Kind = kind
		request.Object.Raw = []byte("invalid")
		result := enabled.Handle(t.Context(), request)
		g.Expect(result.Allowed).To(gomega.BeFalse())
		g.Expect(result.Result.Code).To(gomega.Equal(int32(http.StatusBadRequest)))
	}
}
