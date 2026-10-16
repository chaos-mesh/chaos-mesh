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
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
	experimentstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/experiment"
	schedulestore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/schedule"
)

var collectorKinds = []struct{ kind, field, action, spec string }{
	{"AWSChaos", "awsChaos", "ec2-stop", `{"action":"ec2-stop","region":"us-west-2"}`},
	{"AzureChaos", "azureChaos", "vm-stop", `{"action":"vm-stop"}`},
	{"BlockChaos", "blockChaos", "delay", `{"action":"delay"}`},
	{"DNSChaos", "dnsChaos", "error", `{"action":"error"}`},
	{"GCPChaos", "gcpChaos", "node-stop", `{"action":"node-stop"}`},
	{"HTTPChaos", "httpChaos", "", `{"port":8080,"target":"Request"}`},
	{"IOChaos", "ioChaos", "latency", `{"action":"latency","delay":"1s"}`},
	{"JVMChaos", "jvmChaos", "latency", `{"action":"latency"}`},
	{"KernelChaos", "kernelChaos", "", `{"failKernRequest":{"callchain":[{"funcname":"test"}]}}`},
	{"NetworkChaos", "networkChaos", "delay", `{"action":"delay","delay":{"latency":"1ms"}}`},
	{"PhysicalMachineChaos", "physicalmachineChaos", "stress-cpu", `{"action":"stress-cpu"}`},
	{"PodChaos", "podChaos", "pod-kill", `{"action":"pod-kill"}`},
	{"StressChaos", "stressChaos", "", `{"stressors":{"cpu":{"workers":1}}}`},
	{"TimeChaos", "timeChaos", "", `{"timeOffset":"1s"}`},
}

func TestCollectorChaosKindsPersistSchemaAndEvents(t *testing.T) {
	if len(collectorKinds) != len(v1alpha1.AllKinds()) {
		t.Fatal("collector fixtures must cover each registered chaos kind")
	}
	for _, tc := range collectorKinds {
		t.Run(tc.kind, func(t *testing.T) {
			registered, ok := v1alpha1.AllKinds()[tc.kind]
			if !ok {
				t.Fatalf("unregistered chaos kind %s", tc.kind)
			}
			obj := registered.SpawnObject()
			if err := json.Unmarshal([]byte(`{"spec":`+tc.spec+`}`), obj); err != nil {
				t.Fatal(err)
			}
			obj.SetName("experiment")
			obj.SetNamespace("default")
			obj.SetUID("experiment-uid")
			start := metav1.NewTime(time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC))
			obj.SetCreationTimestamp(start)
			ev := &corev1.Event{ObjectMeta: metav1.ObjectMeta{Name: "event", Namespace: "default"}, InvolvedObject: corev1.ObjectReference{Name: obj.GetName(), Namespace: obj.GetNamespace(), UID: obj.GetUID(), Kind: tc.kind}, Reason: "Applied"}
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj, ev).WithInterceptorFuncs(interceptor.Funcs{Get: getWithTypeMeta}).Build()
			db := lifecycleDB(t)
			experiments, err := experimentstore.NewStore(db)
			if err != nil {
				t.Fatal(err)
			}
			events, err := eventstore.NewStore(db)
			if err != nil {
				t.Fatal(err)
			}
			collector := &ChaosCollector{Client: kube, Log: logr.Discard(), apiType: obj, store: experiments, eventStore: events}
			reconcileObject(t, collector, obj)
			got, err := experiments.FindByUID(context.Background(), "experiment-uid")
			if err != nil {
				t.Fatal(err)
			}
			if got.Kind != tc.kind || got.Action != tc.action || got.Name != "experiment" || got.Namespace != "default" || got.Archived || got.FinishTime != nil || !got.StartTime.Equal(start.Time) {
				t.Errorf("persisted metadata %#v", got.ExperimentMeta)
			}
			assertStoredSpec(t, got.Experiment, obj)
			eventCollector := &EventCollector{Client: kube, Log: logr.Discard(), apiType: &corev1.Event{}, eventStore: events}
			reconcileObject(t, eventCollector, ev)
			rows, err := events.List(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 || rows[0].Kind != tc.kind || rows[0].ObjectID != "experiment-uid" {
				t.Errorf("persisted events %#v", rows)
			}
		})
	}
}

func TestCollectorScheduleKindsPersistSchema(t *testing.T) {
	cases := append(append([]struct{ kind, field, action, spec string }{}, collectorKinds...), struct{ kind, field, action, spec string }{"Workflow", "workflow", "", `{"entry":"wait","templates":[{"name":"wait","templateType":"Suspend","deadline":"1s"}]}`})
	for _, tc := range cases {
		t.Run(tc.kind, func(t *testing.T) {
			obj := &v1alpha1.Schedule{ObjectMeta: metav1.ObjectMeta{Name: "schedule", Namespace: "default", UID: "schedule-uid", CreationTimestamp: metav1.NewTime(time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC))}}
			if err := json.Unmarshal([]byte(`{"spec":{"schedule":"@every 1m","type":"`+tc.kind+`","`+tc.field+`":`+tc.spec+`}}`), obj); err != nil {
				t.Fatal(err)
			}
			kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(obj).WithInterceptorFuncs(interceptor.Funcs{Get: getWithTypeMeta}).Build()
			store, err := schedulestore.NewStore(lifecycleDB(t))
			if err != nil {
				t.Fatal(err)
			}
			collector := &ScheduleCollector{Client: kube, Log: logr.Discard(), apiType: &v1alpha1.Schedule{}, store: store}
			reconcileObject(t, collector, obj)
			got, err := store.FindByUID(context.Background(), "schedule-uid")
			if err != nil {
				t.Fatal(err)
			}
			if got.Kind != "Schedule" || got.Action != tc.action || got.Archived || got.FinishTime != nil || !got.StartTime.Equal(obj.CreationTimestamp.Time) {
				t.Errorf("persisted schedule metadata %#v", got.ScheduleMeta)
			}
			assertStoredSpec(t, got.Schedule, obj)
		})
	}
}

func assertStoredSpec(t *testing.T, raw string, obj client.Object) {
	t.Helper()
	var stored, original struct {
		Spec map[string]any `json:"spec"`
	}
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &original); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored.Spec, original.Spec) {
		t.Errorf("stored spec %#v, want %#v", stored.Spec, original.Spec)
	}
}

func TestEventCollectorWatchFilter(t *testing.T) {
	filter := eventCollectorFilter()
	kinds := []string{"Schedule", "Workflow", "WorkflowNode"}
	for _, tc := range collectorKinds {
		kinds = append(kinds, tc.kind)
	}
	kinds = append(kinds, "Pod", "Deployment", "")
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			obj := &corev1.Event{InvolvedObject: corev1.ObjectReference{Kind: kind}}
			want := kind != "Pod" && kind != "Deployment" && kind != ""
			if got := filter.Create(event.CreateEvent{Object: obj}); got != want {
				t.Errorf("create filter = %v, want %v", got, want)
			}
			if filter.Update(event.UpdateEvent{ObjectOld: obj, ObjectNew: obj}) || filter.Delete(event.DeleteEvent{Object: obj}) || filter.Generic(event.GenericEvent{Object: obj}) {
				t.Error("only initial event creation should be collected")
			}
		})
	}
	if filter.Create(event.CreateEvent{}) || filter.Create(event.CreateEvent{Object: &v1alpha1.PodChaos{}}) {
		t.Error("non-event object accepted")
	}
}

func TestCollectorRejectsUnsupportedScheduleType(t *testing.T) {
	f := newRecoveryFixture(t, "schedule")
	obj := f.object.(*v1alpha1.Schedule)
	obj.Spec.Type = "Unsupported"
	if err := f.kube.Update(context.Background(), obj); err != nil {
		t.Fatal(err)
	}
	result, err := f.run(context.Background())
	requireCollectorRetry(t, result, err)
	if count := f.rowCount(t); count != 0 {
		t.Errorf("unsupported schedule persisted %d rows", count)
	}
}
