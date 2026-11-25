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

package collector

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/clientpool"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	archiveapi "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/archive"
	eventapi "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/event"
	experimentapi "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/experiment"
	scheduleapi "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/schedule"
	workflowapi "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apiserver/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	dashboardmetrics "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/metrics"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
	workflowstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/workflow"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/ttlcontroller"
)

type failFirstExperimentSave struct {
	core.ExperimentStore
	fail  atomic.Bool
	calls atomic.Int32
}

func (s *failFirstExperimentSave) Save(ctx context.Context, row *core.Experiment) error {
	s.calls.Add(1)
	if s.fail.CompareAndSwap(true, false) {
		return errors.New("temporary archive write failure")
	}
	return s.ExperimentStore.Save(ctx, row)
}

func awaitDashboard(t *testing.T, observation string, check func() error) {
	t.Helper()
	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	var last error
	for {
		if last = check(); last == nil {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("%s: %v", observation, last)
		case <-tick.C:
		}
	}
}

func TestDashboardManagerArchiveAndRetention(t *testing.T) {
	assets := os.Getenv("KUBEBUILDER_ASSETS")
	if assets == "" {
		t.Skip("KUBEBUILDER_ASSETS is required for the isolated Kubernetes manager integration test")
	}
	existing := false
	environment := &envtest.Environment{
		BinaryAssetsDirectory: assets,
		UseExistingCluster:    &existing,
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}
	restConfig, err := environment.Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := environment.Stop(); err != nil {
			t.Error(err)
		}
	})
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	kube, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	namespace := "dashboard-recovery"
	if err := kube.Create(ctx, &corev1.Namespace{Name: namespace}); err != nil {
		t.Fatal(err)
	}
	db := lifecycleDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	experiments, err := experimentstore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	schedules, err := schedulestore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	workflows, err := workflowstore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	events, err := eventstore.NewStore(db)
	if err != nil {
		t.Fatal(err)
	}
	faulty := &failFirstExperimentSave{ExperimentStore: experiments}
	faulty.fail.Store(true)
	manager, err := ctrl.NewManager(restConfig, ctrl.Options{Scheme: scheme, Metrics: metricsserver.Options{BindAddress: "0"}, HealthProbeBindAddress: "0"})
	if err != nil {
		t.Fatal(err)
	}
	logger := logr.Discard()
	if err := (&ChaosCollector{Client: manager.GetClient(), Log: logger, store: faulty, eventStore: events}).Setup(manager, &v1alpha1.PodChaos{}); err != nil {
		t.Fatal(err)
	}
	if err := (&ScheduleCollector{Client: manager.GetClient(), Log: logger, store: schedules}).Setup(manager, &v1alpha1.Schedule{}); err != nil {
		t.Fatal(err)
	}
	if err := (&WorkflowCollector{kubeClient: manager.GetClient(), Log: logger, store: workflows}).Setup(manager, &v1alpha1.Workflow{}); err != nil {
		t.Fatal(err)
	}
	if err := (&EventCollector{Client: manager.GetClient(), Log: logger, eventStore: events}).Setup(manager, &corev1.Event{}); err != nil {
		t.Fatal(err)
	}
	managerCtx, stop := context.WithCancel(ctx)
	done := make(chan error, 1)
	go func() { done <- manager.Start(managerCtx) }()
	t.Cleanup(func() {
		stop()
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(10 * time.Second):
			t.Error("manager did not stop")
		}
	})
	cacheCtx, cancelCache := context.WithTimeout(ctx, 15*time.Second)
	defer cancelCache()
	if !manager.GetCache().WaitForCacheSync(cacheCtx) {
		t.Fatal("manager cache did not synchronize")
	}
	pool, err := clientpool.NewLocalClient(restConfig, scheme)
	if err != nil {
		t.Fatal(err)
	}
	previous := clientpool.K8sClients
	clientpool.K8sClients = pool
	t.Cleanup(func() { clientpool.K8sClients = previous })
	conf := &config.ChaosDashboardConfig{TargetNamespace: namespace}
	router := gin.New()
	api := router.Group("/api")
	experimentapi.Register(api, experimentapi.NewService(experiments, events, conf, scheme, logger))
	scheduleapi.Register(api, scheduleapi.NewService(schedules, events, conf, scheme, logger))
	workflowapi.Register(api, workflowapi.NewService(conf, workflows, logger))
	archiveapi.Register(api, archiveapi.NewService(experiments, schedules, events, workflows, conf))
	eventapi.Register(api, eventapi.NewService(events, workflows, conf, logger))
	requestWithContext := func(requestCtx context.Context, method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(requestCtx)
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		return response
	}
	request := func(method, path, body string) *httptest.ResponseRecorder {
		return requestWithContext(ctx, method, path, body)
	}
	resources := []struct {
		path, body string
		object     client.Object
	}{
		{"experiments", fmt.Sprintf(`{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"PodChaos","metadata":{"name":"sample","namespace":%q},"spec":{"action":"pod-kill","mode":"one","selector":{"namespaces":[%q]}}}`, namespace, namespace), &v1alpha1.PodChaos{}},
		{"schedules", fmt.Sprintf(`{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"Schedule","metadata":{"name":"sample","namespace":%q},"spec":{"schedule":"*/5 * * * *","concurrencyPolicy":"Forbid","type":"PodChaos","podChaos":{"action":"pod-kill","mode":"one","selector":{"namespaces":[%q]}}}}`, namespace, namespace), &v1alpha1.Schedule{}},
		{"workflows", fmt.Sprintf(`{"apiVersion":"chaos-mesh.org/v1alpha1","kind":"Workflow","metadata":{"name":"sample","namespace":%q},"spec":{"entry":"wait","templates":[{"name":"wait","templateType":"Suspend","deadline":"1s"}]}}`, namespace), &v1alpha1.Workflow{}},
	}
	canceledCtx, cancelRequest := context.WithCancel(ctx)
	cancelRequest()
	for _, resource := range resources {
		response := requestWithContext(canceledCtx, http.MethodPost, "/api/"+resource.path, resource.body)
		if response.Code == http.StatusOK {
			t.Fatalf("canceled create %s returned success: %s", resource.path, response.Body)
		}
		if err := kube.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "sample"}, resource.object); !apierrors.IsNotFound(err) {
			t.Fatalf("canceled create %s reached Kubernetes: %v", resource.path, err)
		}
	}
	invalidSchedule := strings.Replace(resources[1].body, `"Forbid"`, `"invalid"`, 1)
	invalidResponse := request(http.MethodPost, "/api/schedules", invalidSchedule)
	if invalidResponse.Code != http.StatusBadRequest || !strings.Contains(invalidResponse.Body.String(), `"type":"error.api.bad_request"`) {
		t.Fatalf("schema rejection: status %d body %s", invalidResponse.Code, invalidResponse.Body)
	}
	if err := kube.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "sample"}, &v1alpha1.Schedule{}); !apierrors.IsNotFound(err) {
		t.Fatalf("invalid schedule was persisted: %v", err)
	}
	for _, resource := range resources {
		response := request(http.MethodPost, "/api/"+resource.path, resource.body)
		if response.Code != http.StatusOK {
			t.Fatalf("create %s: status %d body %s", resource.path, response.Code, response.Body)
		}
		if err := kube.Get(ctx, types.NamespacedName{Namespace: namespace, Name: "sample"}, resource.object); err != nil {
			t.Fatal(err)
		}
	}
	chaosUID, scheduleUID, workflowUID := string(resources[0].object.GetUID()), string(resources[1].object.GetUID()), string(resources[2].object.GetUID())
	awaitDashboard(t, "manager persists resources after transient failure", func() error {
		if faulty.calls.Load() < 2 {
			return fmt.Errorf("save attempts %d", faulty.calls.Load())
		}
		if _, err := experiments.FindByUID(ctx, chaosUID); err != nil {
			return err
		}
		if _, err := schedules.FindByUID(ctx, scheduleUID); err != nil {
			return err
		}
		_, err := workflows.FindByUID(ctx, workflowUID)
		return err
	})
	resourceEvent := &corev1.Event{Name: "chaos-created", Namespace: namespace, InvolvedObject: corev1.ObjectReference{APIVersion: v1alpha1.SchemeGroupVersion.String(), Kind: v1alpha1.KindPodChaos, Namespace: namespace, Name: "sample", UID: resources[0].object.GetUID()}, Type: "Normal", Reason: "Created", Message: "initial"}
	if err := kube.Create(ctx, resourceEvent); err != nil {
		t.Fatal(err)
	}
	awaitDashboard(t, "create event reaches the HTTP endpoint", func() error {
		response := request(http.MethodGet, "/api/events?object_id="+chaosUID, "")
		if response.Code != 200 || !strings.Contains(response.Body.String(), `"message":"initial"`) {
			return fmt.Errorf("event response %d %s", response.Code, response.Body)
		}
		return nil
	})
	resourceEvent.Message = "updated"
	if err := kube.Update(ctx, resourceEvent); err != nil {
		t.Fatal(err)
	}
	barrier := resourceEvent.DeepCopy()
	barrier.ObjectMeta = metav1.ObjectMeta{Name: "after-update", Namespace: namespace}
	barrier.Message = "barrier"
	if err := kube.Create(ctx, barrier); err != nil {
		t.Fatal(err)
	}
	awaitDashboard(t, "later create event is collected", func() error {
		rows, err := events.ListByUID(ctx, chaosUID)
		if err != nil {
			return err
		}
		for _, row := range rows {
			if row.Message == "barrier" {
				return nil
			}
		}
		return fmt.Errorf("barrier missing from %d rows", len(rows))
	})
	rows, err := events.ListByUID(ctx, chaosUID)
	if err != nil || len(rows) != 2 {
		t.Fatalf("event updates were persisted: rows %v error %v", rows, err)
	}
	for _, resource := range resources {
		if err := kube.Delete(ctx, resource.object); err != nil {
			t.Fatal(err)
		}
	}
	awaitDashboard(t, "delete watches archive all resources", func() error {
		experiment, err := experiments.FindByUID(ctx, chaosUID)
		if err != nil {
			return err
		}
		if !experiment.Archived {
			return errors.New("experiment not archived")
		}
		schedule, err := schedules.FindByUID(ctx, scheduleUID)
		if err != nil {
			return err
		}
		if !schedule.Archived {
			return errors.New("schedule not archived")
		}
		workflow, err := workflows.FindByUID(ctx, workflowUID)
		if err != nil {
			return err
		}
		if !workflow.Archived {
			return errors.New("workflow not archived")
		}
		return nil
	})
	for _, path := range []string{"/api/archives/" + chaosUID, "/api/archives/schedules/" + scheduleUID, "/api/archives/workflows/" + workflowUID} {
		response := request(http.MethodGet, path, "")
		if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "sample") {
			t.Fatalf("archive %s: status %d body %s", path, response.Code, response.Body)
		}
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(dashboardmetrics.NewCollector(logger, experiments, schedules, workflows))
	metrics := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	scrape := func() string {
		response := httptest.NewRecorder()
		metrics.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		if response.Code != 200 {
			t.Fatalf("metrics status %d body %s", response.Code, response.Body)
		}
		return response.Body.String()
	}
	before := scrape()
	for _, expected := range []string{fmt.Sprintf(`chaos_dashboard_archived_experiments{namespace=%q,type="PodChaos"} 1`, namespace), fmt.Sprintf(`chaos_dashboard_archived_schedules{namespace=%q} 1`, namespace), fmt.Sprintf(`chaos_dashboard_archived_workflows{namespace=%q} 1`, namespace)} {
		if !strings.Contains(before, expected) {
			t.Fatalf("metric %q missing from %s", expected, before)
		}
	}
	cutoff := time.Now().UTC().Add(-2 * time.Hour)
	for _, model := range []any{&core.Experiment{}, &core.Schedule{}, &core.WorkflowEntity{}} {
		if err := db.Model(model).Where("namespace = ?", namespace).Update("finish_time", cutoff).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Model(&core.Event{}).Where("namespace = ?", namespace).Update("created_at", cutoff).Error; err != nil {
		t.Fatal(err)
	}
	retentionCtx, stopRetention := context.WithCancel(ctx)
	t.Cleanup(stopRetention)
	ttlcontroller.Register(retentionCtx, ttlcontroller.NewController(events, experiments, schedules, workflows, &config.TTLConfig{EventTTL: time.Hour, ExperimentTTL: time.Hour, ScheduleTTL: time.Hour, WorkflowTTL: time.Hour, ResyncPeriod: time.Hour}, logger))
	awaitDashboard(t, "scheduled retention deletes archived rows and events", func() error {
		if _, err := experiments.FindByUID(ctx, chaosUID); !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("experiment lookup %v", err)
		}
		if _, err := schedules.FindByUID(ctx, scheduleUID); !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("schedule lookup %v", err)
		}
		if _, err := workflows.FindByUID(ctx, workflowUID); !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("workflow lookup %v", err)
		}
		rows, err := events.ListByUID(ctx, chaosUID)
		if err != nil {
			return err
		}
		if len(rows) != 0 {
			return fmt.Errorf("%d events remain", len(rows))
		}
		return nil
	})
	stopRetention()
	if after := scrape(); strings.Contains(after, "chaos_dashboard_archived_") {
		t.Fatalf("stale archive metrics remain: %s", after)
	}
}
