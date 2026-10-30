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
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-logr/logr"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
	eventstore "github.com/chaos-mesh/chaos-mesh/pkg/dashboard/store/event"
)

type lostEventAcknowledgementStore struct {
	core.EventStore
	loseNextAcknowledgement bool
}

func (s *lostEventAcknowledgementStore) Create(ctx context.Context, event *core.Event) error {
	if err := s.EventStore.Create(ctx, event); err != nil {
		return err
	}
	if s.loseNextAcknowledgement {
		s.loseNextAcknowledgement = false
		return errors.New("database committed event but acknowledgement was lost")
	}
	return nil
}

func openReplayEventStore(t *testing.T, path string) (core.EventStore, func()) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	closeDB := func() {
		if err := sqlDB.Close(); err != nil {
			t.Fatal(err)
		}
	}
	store, err := eventstore.NewStore(db)
	if err != nil {
		closeDB()
		t.Fatal(err)
	}
	return store, closeDB
}

func replayEvent(uid types.UID, name string) *corev1.Event {
	return &corev1.Event{
		Name: name, Namespace: "default", UID: uid,
		CreationTimestamp: metav1.NewTime(time.Date(2026, 10, 16, 1, 0, 0, 0, time.UTC)),
		InvolvedObject:    corev1.ObjectReference{Kind: v1alpha1.KindPodChaos, Name: "target", Namespace: "default", UID: "target-uid"},
		Type:              corev1.EventTypeNormal, Reason: "Applied", Message: "Fault injected",
	}
}

func replayCollector(store core.EventStore, events ...*corev1.Event) *EventCollector {
	objects := []client.Object{&v1alpha1.PodChaos{Name: "target", Namespace: "default", UID: "target-uid"}}
	for _, event := range events {
		objects = append(objects, event.DeepCopy())
	}
	return &EventCollector{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build(),
		Log:    logr.Discard(), apiType: &corev1.Event{}, eventStore: store,
	}
}

func requireReplayEventRows(t *testing.T, store core.EventStore, want int) {
	t.Helper()
	rows, err := store.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != want {
		t.Fatalf("persisted %d event rows, want %d source events", len(rows), want)
	}
}

func TestEventCollectorReplayAfterLostCommitAcknowledgement(t *testing.T) {
	for _, recreate := range []bool{false, true} {
		name := "same_collector"
		if recreate {
			name = "recreated_store_and_collector"
		}
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "events.db")
			store, closeDB := openReplayEventStore(t, path)
			defer func() { closeDB() }()
			source := replayEvent("source-event-1", "event-1")
			wrapped := &lostEventAcknowledgementStore{EventStore: store, loseNextAcknowledgement: true}
			collector := replayCollector(wrapped, source)
			request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(source)}
			result, err := collector.Reconcile(context.Background(), request)
			requireCollectorRetry(t, result, err)
			requireReplayEventRows(t, store, 1)
			if recreate {
				closeDB()
				store, closeDB = openReplayEventStore(t, path)
				collector = replayCollector(store, source)
			}
			result, err = collector.Reconcile(context.Background(), request)
			if err != nil || result != (ctrl.Result{}) {
				t.Fatalf("replay did not converge: result=%#v, error=%v", result, err)
			}
			requireReplayEventRows(t, store, 1)
		})
	}
}

func TestEventCollectorReplayAfterStoreRecreation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.db")
	store, closeDB := openReplayEventStore(t, path)
	defer func() { closeDB() }()
	source := replayEvent("source-event-1", "event-1")
	request := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(source)}
	result, err := replayCollector(store, source).Reconcile(context.Background(), request)
	if err != nil || result != (ctrl.Result{}) {
		t.Fatalf("initial collection failed: result=%#v, error=%v", result, err)
	}
	requireReplayEventRows(t, store, 1)
	closeDB()
	store, closeDB = openReplayEventStore(t, path)
	result, err = replayCollector(store, source).Reconcile(context.Background(), request)
	if err != nil || result != (ctrl.Result{}) {
		t.Fatalf("replay did not converge: result=%#v, error=%v", result, err)
	}
	requireReplayEventRows(t, store, 1)
}

func TestEventCollectorDistinctSourceEventsWithIdenticalContent(t *testing.T) {
	store, closeDB := openReplayEventStore(t, filepath.Join(t.TempDir(), "events.db"))
	defer closeDB()
	first := replayEvent("source-event-1", "event-1")
	second := replayEvent("source-event-2", "event-2")
	collector := replayCollector(store, first, second)
	for _, source := range []*corev1.Event{first, second} {
		result, err := collector.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKeyFromObject(source)})
		if err != nil || result != (ctrl.Result{}) {
			t.Fatalf("collection failed: result=%#v, error=%v", result, err)
		}
	}
	requireReplayEventRows(t, store, 2)
}
