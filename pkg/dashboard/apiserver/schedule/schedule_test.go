// Copyright 2021 Chaos Mesh Authors.
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

package schedule

import (
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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	apiservertypes "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type scheduleClientPool struct {
	clientpool.Clients
	kube client.Client
}

func (p scheduleClientPool) Client(string) (client.Client, error) { return p.kube, nil }

type scheduleLookup struct {
	core.ScheduleStore
	schedules map[string]*core.Schedule
}

func (s scheduleLookup) FindByUID(_ context.Context, uid string) (*core.Schedule, error) {
	if schedule := s.schedules[uid]; schedule != nil {
		return schedule, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func TestScheduleAPILifecycle(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	scheme := runtime.NewScheme()
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
	makeSchedule := func(namespace, name string, created int64) *v1alpha1.Schedule {
		schedule := &v1alpha1.Schedule{Namespace: namespace, Name: name, UID: types.UID(namespace + "-" + name),
			CreationTimestamp: metav1.NewTime(time.Unix(created, 0)), Annotations: map[string]string{"owner": "test"}}
		schedule.Spec.Type = v1alpha1.ScheduleTypePodChaos
		schedule.Spec.Schedule = "@every 1h"
		return schedule
	}
	primary, secondary, foreign := makeSchedule("test", "primary", 100), makeSchedule("test", "secondary", 200), makeSchedule("foreign", "primary", 300)
	child := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "child", UID: "child-uid", Labels: map[string]string{v1alpha1.LabelManagedBy: primary.Name}}}
	foreignChild := child.DeepCopy()
	foreignChild.Namespace, foreignChild.UID = "foreign", "foreign-child-uid"
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(primary, secondary, foreign, child, foreignChild).Build()
	old := clientpool.K8sClients
	clientpool.K8sClients = scheduleClientPool{kube: kube}
	t.Cleanup(func() { clientpool.K8sClients = old })
	store := scheduleLookup{schedules: map[string]*core.Schedule{
		"one":   {ScheduleMeta: core.ScheduleMeta{Namespace: "test", Name: primary.Name, UID: string(primary.UID)}},
		"two":   {ScheduleMeta: core.ScheduleMeta{Namespace: "test", Name: secondary.Name, UID: string(secondary.UID)}},
		"stale": {ScheduleMeta: core.ScheduleMeta{Namespace: "test", Name: "missing", UID: "stale"}},
		"old":   {ScheduleMeta: core.ScheduleMeta{Namespace: "test", Name: primary.Name, UID: "old"}},
	}}
	router := gin.New()
	Register(router.Group("/api"), NewService(store, nil, &config.ChaosDashboardConfig{TargetNamespace: "test"}, scheme, logr.Discard()))
	request := func(method, path string, code int) []byte {
		t.Helper()
		out := httptest.NewRecorder()
		router.ServeHTTP(out, httptest.NewRequest(method, "/api/schedules"+path, nil))
		g.Expect(out.Code).To(Equal(code), out.Body.String())
		return out.Body.Bytes()
	}

	var listed []apiservertypes.Schedule
	g.Expect(json.Unmarshal(request(http.MethodGet, "", http.StatusOK), &listed)).To(Succeed())
	g.Expect(listed).To(HaveLen(2))
	g.Expect(listed[0].Name).To(Equal(secondary.Name))
	g.Expect(listed[1].Name).To(Equal(primary.Name))
	g.Expect(json.Unmarshal(request(http.MethodGet, "?name=primary", http.StatusOK), &listed)).To(Succeed())
	g.Expect(listed).To(HaveLen(1))
	g.Expect(listed[0].Namespace).To(Equal("test"))

	var detail apiservertypes.ScheduleDetail
	g.Expect(json.Unmarshal(request(http.MethodGet, "/one", http.StatusOK), &detail)).To(Succeed())
	g.Expect(detail.ExperimentUIDs).To(Equal([]string{"child-uid"}))
	g.Expect(detail.KubeObject.Kind).To(Equal("Schedule"))
	key := client.ObjectKeyFromObject(primary)
	request(http.MethodGet, "/old", http.StatusNotFound)
	request(http.MethodPut, "/pause/old", http.StatusNotFound)
	request(http.MethodPut, "/start/old", http.StatusNotFound)
	request(http.MethodDelete, "/old", http.StatusNotFound)
	request(http.MethodDelete, "?uids=old,two", http.StatusNotFound)
	g.Expect(kube.Get(ctx, key, primary)).To(Succeed())
	g.Expect(primary.IsPaused()).To(BeFalse())

	request(http.MethodPut, "/pause/one", http.StatusOK)
	g.Expect(kube.Get(ctx, key, primary)).To(Succeed())
	g.Expect(primary.IsPaused()).To(BeTrue())
	request(http.MethodPut, "/start/one", http.StatusOK)
	g.Expect(kube.Get(ctx, key, primary)).To(Succeed())
	g.Expect(primary.IsPaused()).To(BeFalse())
	g.Expect(primary.Annotations).To(HaveKeyWithValue("owner", "test"))
	request(http.MethodPut, "/pause/missing", http.StatusNotFound)
	request(http.MethodPut, "/start/stale", http.StatusNotFound)
	request(http.MethodGet, "/stale", http.StatusNotFound)

	request(http.MethodDelete, "?uids=one,missing,two", http.StatusNotFound)
	g.Expect(kube.Get(ctx, key, primary)).To(Satisfy(apierrors.IsNotFound))
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(secondary), secondary)).To(Succeed())
	request(http.MethodDelete, "/two", http.StatusOK)
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(secondary), secondary)).To(Satisfy(apierrors.IsNotFound))
	g.Expect(kube.Get(ctx, client.ObjectKeyFromObject(foreign), foreign)).To(Succeed())
}
