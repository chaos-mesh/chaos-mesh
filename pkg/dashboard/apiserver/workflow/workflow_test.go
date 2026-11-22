// Copyright 2022 Chaos Mesh Authors.
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

package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/curl"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type workflowClientPool struct {
	clientpool.Clients
	kube client.Client
}

func (p workflowClientPool) Client(string) (client.Client, error) { return p.kube, nil }

type workflowLookup struct {
	core.WorkflowStore
	records map[string]*core.WorkflowEntity
}

func (s workflowLookup) FindByUID(_ context.Context, uid string) (*core.WorkflowEntity, error) {
	if record := s.records[uid]; record != nil {
		return record, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func sendWorkflowRequest(t *testing.T, router http.Handler, method, path string, body []byte, wantStatus int) []byte {
	t.Helper()
	req := httptest.NewRequest(method, "/api/workflows"+path, bytes.NewReader(body))
	out := httptest.NewRecorder()
	router.ServeHTTP(out, req)
	NewWithT(t).Expect(out.Code).To(Equal(wantStatus), out.Body.String())
	return out.Body.Bytes()
}

func TestWorkflowRoutesKeepArchivedUIDSeparateFromReplacement(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	scheme := runtime.NewScheme()
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	archived := &v1alpha1.Workflow{Namespace: "team", Name: "repeat", UID: "old", Spec: v1alpha1.WorkflowSpec{Entry: "archived-entry"}}
	live := archived.DeepCopy()
	live.UID, live.Spec.Entry = "live", "live-entry"
	live.Status.StartTime = new(metav1.NewTime(time.Unix(200, 0)))
	older := &v1alpha1.Workflow{Namespace: "team", Name: "earlier", UID: "earlier"}
	older.Status.StartTime = new(metav1.NewTime(time.Unix(100, 0)))
	foreign := live.DeepCopy()
	foreign.Namespace, foreign.UID = "foreign", "foreign"
	const longTemplate = "a-template-name-long-enough-for-kubernetes-to-truncate-it"
	child := &v1alpha1.WorkflowNode{
		Namespace: "team", Name: "truncated-suffix",
		Labels: map[string]string{v1alpha1.LabelWorkflow: live.Name},
		Spec:   v1alpha1.WorkflowNodeSpec{Type: v1alpha1.TypeSuspend, TemplateName: longTemplate},
	}
	node := &v1alpha1.WorkflowNode{
		Namespace: "team", Name: "live-node",
		Labels: map[string]string{v1alpha1.LabelWorkflow: live.Name},
		Spec:   v1alpha1.WorkflowNodeSpec{Type: v1alpha1.TypeSerial, Children: []string{longTemplate, "live-entry"}},
		Status: v1alpha1.WorkflowNodeStatus{FinishedChildren: []corev1.LocalObjectReference{{Name: child.Name}, {Name: "obsolete"}, {Name: "old-entry-z"}}},
	}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(live, older, foreign, node, child).Build()
	previous := clientpool.K8sClients
	clientpool.K8sClients = workflowClientPool{kube: kube}
	t.Cleanup(func() { clientpool.K8sClients = previous })
	records := make(map[string]*core.WorkflowEntity)
	for i, workflow := range []*v1alpha1.Workflow{archived, live} {
		record, err := core.WorkflowCR2WorkflowEntity(workflow)
		g.Expect(err).NotTo(HaveOccurred())
		record.ID = uint(i + 1)
		records[record.UID] = record
	}
	router := gin.New()
	conf := &config.ChaosDashboardConfig{TargetNamespace: "team"}
	Register(router.Group("/api"), NewService(conf, workflowLookup{records: records}, logr.Discard()))

	request := func(method, path string, body []byte, code int) []byte {
		t.Helper()
		return sendWorkflowRequest(t, router, method, path, body, code)
	}
	var listed []core.WorkflowMeta
	g.Expect(json.Unmarshal(request(http.MethodGet, "", nil, http.StatusOK), &listed)).To(Succeed())
	g.Expect(listed).To(HaveLen(2))
	g.Expect(listed[0].UID).To(Equal("live"))
	g.Expect(listed[0].ID).To(Equal(uint(2)))
	g.Expect(listed[1].UID).To(Equal("earlier"))
	g.Expect(listed[1].ID).To(BeZero())
	conf.ClusterScoped = true
	g.Expect(json.Unmarshal(request(http.MethodGet, "", nil, http.StatusOK), &listed)).To(Succeed())
	g.Expect(listed).To(HaveLen(3))

	var detail core.WorkflowDetail
	g.Expect(json.Unmarshal(request(http.MethodGet, "/old", nil, http.StatusOK), &detail)).To(Succeed())
	g.Expect(detail.Topology.Nodes).To(BeEmpty())
	g.Expect(detail.KubeObject.Spec).To(HaveKeyWithValue("entry", "archived-entry"))
	g.Expect(json.Unmarshal(request(http.MethodGet, "/live", nil, http.StatusOK), &detail)).To(Succeed())
	g.Expect(detail.Topology.Nodes).To(HaveLen(2))
	g.Expect(detail.Topology.Nodes).To(ContainElement(HaveField("Serial", Equal([]core.NodeNameWithTemplate{
		{Template: longTemplate, Name: child.Name}, {Template: "live-entry"},
	}))))
	g.Expect(detail.KubeObject.Spec).To(HaveKeyWithValue("entry", "live-entry"))
	request(http.MethodGet, "/missing", nil, http.StatusNotFound)

	update := []byte(`{"metadata":{"name":"repeat","namespace":"team"},"spec":{"entry":"updated"}}`)
	request(http.MethodPut, "/old", update, http.StatusNotFound)
	request(http.MethodDelete, "/old", nil, http.StatusNotFound)
	var stored v1alpha1.Workflow
	key := client.ObjectKeyFromObject(live)
	g.Expect(kube.Get(ctx, key, &stored)).To(Succeed())
	g.Expect(stored.UID).To(Equal(types.UID("live")))
	g.Expect(stored.Spec.Entry).To(Equal("live-entry"))

	request(http.MethodPut, "/live", []byte(`{"metadata":{"name":"repeat","namespace":"foreign"}}`), http.StatusBadRequest)
	request(http.MethodPut, "/live", []byte(`{"metadata":{"name":"different","namespace":"team"}}`), http.StatusBadRequest)
	request(http.MethodPut, "/live", update, http.StatusOK)
	g.Expect(kube.Get(ctx, key, &stored)).To(Succeed())
	g.Expect(stored.UID).To(Equal(types.UID("live")))
	g.Expect(stored.Spec.Entry).To(Equal("updated"))
	request(http.MethodDelete, "/live", nil, http.StatusOK)
	request(http.MethodDelete, "/live", nil, http.StatusNotFound)
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(foreign), &stored)).To(Succeed())

	request(http.MethodPost, "", []byte(`{`), http.StatusBadRequest)
	request(http.MethodPost, "", []byte(`{"metadata":{"name":"created","namespace":"team"},"spec":{"entry":"initial"}}`), http.StatusOK)
	g.Expect(kube.Get(ctx, client.ObjectKey{Namespace: "team", Name: "created"}, &stored)).To(Succeed())
	g.Expect(stored.Spec.Entry).To(Equal("initial"))
}

func TestWorkflowHTTPTaskRoutesRoundTrip(t *testing.T) {
	g := NewWithT(t)
	router := gin.New()
	Register(router.Group("/api"), NewService(nil, nil, logr.Discard()))
	request := func(path string, body []byte, code int) []byte {
		t.Helper()
		return sendWorkflowRequest(t, router, http.MethodPost, path, body, code)
	}
	want := curl.RequestForm{
		Name: "probe",
		CommandFlags: curl.CommandFlags{
			Method: "POST", URL: "https://example.invalid/path?foo=bar",
			Header: curl.Header{"X-Trace": {"one", "two"}}, Body: `{"name":"value"}`, JSONContent: true, FollowLocation: true,
		},
	}
	body, err := json.Marshal(want)
	g.Expect(err).NotTo(HaveOccurred())
	rendered := request("/render-task/http", body, http.StatusOK)
	g.Expect(request("/validate-task/http", rendered, http.StatusOK)).To(MatchJSON("true"))
	var got curl.RequestForm
	g.Expect(json.Unmarshal(request("/parse-task/http", rendered, http.StatusOK), &got)).To(Succeed())
	g.Expect(got).To(Equal(want))
	g.Expect(request("/validate-task/http", []byte("{}"), http.StatusOK)).To(MatchJSON("false"))
	request("/parse-task/http", []byte("{}"), http.StatusInternalServerError)
	for _, path := range []string{"/render-task/http", "/validate-task/http", "/parse-task/http"} {
		request(path, []byte("{"), http.StatusBadRequest)
	}
}
