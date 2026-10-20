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

package physicalmachinechaos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/desiredphase"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/finalizers"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/records"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type processEndpoint struct {
	server       *httptest.Server
	posts        atomic.Int32
	deletes      atomic.Int32
	failures     atomic.Int32
	deleteStatus int
}

func newProcessEndpoint(t *testing.T, deleteStatus int) *processEndpoint {
	t.Helper()
	endpoint := &processEndpoint{deleteStatus: deleteStatus}
	endpoint.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			endpoint.posts.Add(1)
			if r.URL.Path != "/api/attack/process" {
				t.Errorf("unexpected POST path %q", r.URL.Path)
			}
			var request struct {
				UID     string `json:"uid"`
				Signal  int    `json:"signal"`
				Process string `json:"process"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode POST: %v", err)
			}
			if request.UID != "process-uid" || request.Signal == 0 || request.Process != "1234" {
				t.Errorf("unexpected POST body %+v", request)
			}
			if endpoint.failures.Add(-1) >= 0 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodDelete:
			endpoint.deletes.Add(1)
			if r.URL.Path != "/api/attack/process-uid" {
				t.Errorf("unexpected DELETE path %q", r.URL.Path)
			}
			w.WriteHeader(endpoint.deleteStatus)
		default:
			t.Errorf("unexpected HTTP method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(endpoint.server.Close)
	return endpoint
}

func processChaos(address string, selector bool) *v1alpha1.PhysicalMachineChaos {
	obj := &v1alpha1.PhysicalMachineChaos{
		ObjectMeta: metav1.ObjectMeta{Name: "process", Namespace: "default", CreationTimestamp: metav1.Now(), Finalizers: []string{finalizers.RecordFinalizer}},
		Spec: v1alpha1.PhysicalMachineChaosSpec{
			Action:                  v1alpha1.PMProcessAction,
			PhysicalMachineSelector: v1alpha1.PhysicalMachineSelector{Address: []string{address}, Mode: v1alpha1.AllMode},
			ExpInfo:                 v1alpha1.ExpInfo{UID: "process-uid", Process: &v1alpha1.ProcessSpec{Process: "1234", Signal: 9}},
		},
	}
	id := address
	if selector {
		obj.Spec.Address = nil
		id = "default/machine"
	}
	obj.Status.Experiment.Records = []*v1alpha1.Record{{Id: id, Phase: v1alpha1.NotInjected}}
	return obj
}

func processClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	security, maxEvents := config.ControllerCfg.ChaosdSecurityMode, config.ControllerCfg.MaxEvents
	config.ControllerCfg.ChaosdSecurityMode = false
	config.ControllerCfg.MaxEvents = 100
	t.Cleanup(func() { config.ControllerCfg.ChaosdSecurityMode = security; config.ControllerCfg.MaxEvents = maxEvents })
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func fetchProcess(t *testing.T, c client.Client) *v1alpha1.PhysicalMachineChaos {
	t.Helper()
	obj := &v1alpha1.PhysicalMachineChaos{}
	require.NoError(t, c.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "process"}, obj))
	return obj
}

func reconcileProcessRecords(t *testing.T, c client.Client) ctrl.Result {
	t.Helper()
	r := &records.Reconciler{Client: c, Reader: c, Object: &v1alpha1.PhysicalMachineChaos{}, Recorder: recorder.NewDebugRecorder(), Log: logr.Discard(), Impl: &Impl{Client: c, Log: logr.Discard()}}
	result, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKey{Namespace: "default", Name: "process"}})
	require.NoError(t, err)
	return result
}

func reconcileProcess(t *testing.T, c client.Client) ctrl.Result {
	t.Helper()
	r := &desiredphase.Reconciler{Client: c, Object: &v1alpha1.PhysicalMachineChaos{}, Recorder: recorder.NewDebugRecorder(), Log: logr.Discard()}
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKey{Namespace: "default", Name: "process"}})
	require.NoError(t, err)
	return reconcileProcessRecords(t, c)
}

func cleanProcessFinalizer(t *testing.T, c client.Client) {
	t.Helper()
	r := &finalizers.CleanReconciler{ReconcilerMeta: finalizers.ReconcilerMeta{Client: c, Object: &v1alpha1.PhysicalMachineChaos{}, Recorder: recorder.NewDebugRecorder(), Log: logr.Discard()}}
	_, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: client.ObjectKey{Namespace: "default", Name: "process"}})
	require.NoError(t, err)
}

func TestProcessOneShotLifecycle(t *testing.T) {
	for _, useSelector := range []bool{false, true} {
		name := "address"
		if useSelector {
			name = "selector"
		}
		t.Run(name, func(t *testing.T) {
			endpoint := newProcessEndpoint(t, http.StatusInternalServerError)
			obj := processChaos(endpoint.server.URL, useSelector)
			machine := &v1alpha1.PhysicalMachine{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "machine"}, Spec: v1alpha1.PhysicalMachineSpec{Address: endpoint.server.URL}}
			c := processClient(t, obj, machine)
			require.False(t, controller.IsChaosFinished(obj, time.Now()))
			require.False(t, reconcileProcess(t, c).Requeue)
			obj = fetchProcess(t, c)
			require.Equal(t, v1alpha1.Injected, obj.Status.Experiment.Records[0].Phase)
			require.Equal(t, 1, obj.Status.Experiment.Records[0].InjectedCount)
			for _, state := range []string{"repeat", "pause", "resume", "expired"} {
				obj = fetchProcess(t, c)
				switch state {
				case "pause":
					obj.Annotations = map[string]string{v1alpha1.PauseAnnotationKey: "true"}
				case "resume":
					obj.Annotations = nil
				case "expired":
					duration := "1s"
					obj.Spec.Duration = &duration
					obj.CreationTimestamp = metav1.NewTime(time.Now().Add(-time.Hour))
				}
				require.NoError(t, c.Update(context.Background(), obj))
				require.False(t, reconcileProcess(t, c).Requeue, state)
				obj = fetchProcess(t, c)
				require.Equal(t, v1alpha1.RunningPhase, obj.Status.Experiment.DesiredPhase, state)
				require.Equal(t, v1alpha1.Injected, obj.Status.Experiment.Records[0].Phase, state)
				require.EqualValues(t, 1, endpoint.posts.Load(), state)
				require.Zero(t, endpoint.deletes.Load(), state)
			}
			obj.Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
			require.NoError(t, c.Update(context.Background(), obj))
			require.False(t, reconcileProcessRecords(t, c).Requeue)
			obj = fetchProcess(t, c)
			require.Equal(t, v1alpha1.NotInjected, obj.Status.Experiment.Records[0].Phase)
			require.Equal(t, 1, obj.Status.Experiment.Records[0].InjectedCount)
			require.Zero(t, endpoint.deletes.Load())
			if useSelector {
				require.NoError(t, c.Delete(context.Background(), machine))
			}
			require.False(t, reconcileProcess(t, c).Requeue)
			obj = fetchProcess(t, c)
			require.Equal(t, v1alpha1.Injected, obj.Status.Experiment.Records[0].Phase)
			require.True(t, controller.IsChaosFinished(obj, time.Now()))
			require.False(t, reconcileProcess(t, c).Requeue)
			require.EqualValues(t, 1, endpoint.posts.Load())
			require.Zero(t, endpoint.deletes.Load())
		})
	}
}

func TestProcessOneShotHistoricalNotInjected(t *testing.T) {
	for _, useSelector := range []bool{false, true} {
		name := "address"
		if useSelector {
			name = "missing machine"
		}
		t.Run(name, func(t *testing.T) {
			endpoint := newProcessEndpoint(t, http.StatusInternalServerError)
			obj := processChaos(endpoint.server.URL, useSelector)
			obj.Status.Experiment.Records[0].InjectedCount = 1
			c := processClient(t, obj)
			require.False(t, reconcileProcess(t, c).Requeue)
			require.Zero(t, endpoint.posts.Load())
			obj = fetchProcess(t, c)
			require.Equal(t, v1alpha1.Injected, obj.Status.Experiment.Records[0].Phase)
			require.True(t, controller.IsChaosFinished(obj, time.Now()))
			require.False(t, reconcileProcess(t, c).Requeue)
			require.Zero(t, endpoint.posts.Load())
			require.Zero(t, endpoint.deletes.Load())
		})
	}
}

func TestProcessOneShotPartialFailure(t *testing.T) {
	first := newProcessEndpoint(t, http.StatusInternalServerError)
	second := newProcessEndpoint(t, http.StatusInternalServerError)
	second.failures.Store(1)
	obj := processChaos(first.server.URL, false)
	obj.Spec.Address = append(obj.Spec.Address, second.server.URL)
	obj.Status.Experiment.Records = append(obj.Status.Experiment.Records, &v1alpha1.Record{Id: second.server.URL, Phase: v1alpha1.NotInjected})
	c := processClient(t, obj)
	require.True(t, reconcileProcess(t, c).Requeue)
	obj = fetchProcess(t, c)
	require.Equal(t, 1, obj.Status.Experiment.Records[0].InjectedCount)
	require.Equal(t, 0, obj.Status.Experiment.Records[1].InjectedCount)
	require.Equal(t, v1alpha1.NotInjected, obj.Status.Experiment.Records[1].Phase)
	require.False(t, controller.IsChaosFinished(obj, time.Now()))
	require.False(t, reconcileProcess(t, c).Requeue)
	obj = fetchProcess(t, c)
	require.Equal(t, v1alpha1.Injected, obj.Status.Experiment.Records[1].Phase)
	require.True(t, controller.IsChaosFinished(obj, time.Now()))
	require.False(t, reconcileProcess(t, c).Requeue)
	require.EqualValues(t, 1, first.posts.Load())
	require.EqualValues(t, 2, second.posts.Load())
	require.Zero(t, first.deletes.Load())
	require.Zero(t, second.deletes.Load())
}

func TestProcessOneShotDeletion(t *testing.T) {
	for _, useSelector := range []bool{false, true} {
		for _, phase := range []v1alpha1.Phase{v1alpha1.Injected, "Injected/recovering", v1alpha1.NotInjected} {
			name := "address/" + string(phase)
			if useSelector {
				name = "missing machine/" + string(phase)
			}
			t.Run(name, func(t *testing.T) {
				endpoint := newProcessEndpoint(t, http.StatusInternalServerError)
				obj := processChaos(endpoint.server.URL, useSelector)
				obj.Status.Experiment.DesiredPhase = v1alpha1.RunningPhase
				obj.Status.Experiment.Records[0].Phase = phase
				obj.Status.Experiment.Records[0].InjectedCount = 1
				c := processClient(t, obj)
				require.NoError(t, c.Delete(context.Background(), obj))
				require.False(t, reconcileProcess(t, c).Requeue)
				obj = fetchProcess(t, c)
				require.Equal(t, v1alpha1.NotInjected, obj.Status.Experiment.Records[0].Phase)
				cleanProcessFinalizer(t, c)
				err := c.Get(context.Background(), client.ObjectKeyFromObject(obj), &v1alpha1.PhysicalMachineChaos{})
				require.True(t, apierrors.IsNotFound(err), "finalizer must allow deletion: %v", err)
				require.Zero(t, endpoint.posts.Load())
				require.Zero(t, endpoint.deletes.Load())
			})
		}
	}
}

func TestProcessRecoverableLifecycle(t *testing.T) {
	for _, process := range []struct {
		name    string
		signal  int
		command string
	}{{"stop", 19, ""}, {"command", 9, "start-service"}} {
		for _, response := range []struct {
			name             string
			status           int
			transportFailure bool
			wantRetry        bool
		}{{"ok", http.StatusOK, false, false}, {"not found", http.StatusNotFound, false, false}, {"server failure", http.StatusInternalServerError, false, true}, {"transport failure", 0, true, true}} {
			t.Run(process.name+"/"+response.name, func(t *testing.T) {
				endpoint := newProcessEndpoint(t, response.status)
				obj := processChaos(endpoint.server.URL, false)
				obj.Spec.Process.Signal = process.signal
				obj.Spec.Process.RecoverCmd = process.command
				obj.Status.Experiment.Records[0].InjectedCount = 1
				c := processClient(t, obj)
				require.False(t, reconcileProcess(t, c).Requeue)
				require.EqualValues(t, 1, endpoint.posts.Load(), "recoverable experiments still apply after a previous recovery")
				obj = fetchProcess(t, c)
				require.NoError(t, c.Delete(context.Background(), obj))
				if response.transportFailure {
					endpoint.server.Close()
				}
				require.Equal(t, response.wantRetry, reconcileProcess(t, c).Requeue)
				cleanProcessFinalizer(t, c)
				if response.wantRetry {
					obj = fetchProcess(t, c)
					require.Equal(t, v1alpha1.Injected, obj.Status.Experiment.Records[0].Phase)
					require.Contains(t, obj.Finalizers, finalizers.RecordFinalizer)
					require.Zero(t, obj.Status.Experiment.Records[0].RecoveredCount)
					events := obj.Status.Experiment.Records[0].Events
					require.Equal(t, v1alpha1.TypeFailed, events[len(events)-1].Type)
					require.Equal(t, v1alpha1.Recover, events[len(events)-1].Operation)
					require.NotEmpty(t, events[len(events)-1].Message)
				} else {
					err := c.Get(context.Background(), client.ObjectKeyFromObject(obj), &v1alpha1.PhysicalMachineChaos{})
					require.True(t, apierrors.IsNotFound(err))
				}
				if !response.transportFailure {
					require.EqualValues(t, 1, endpoint.deletes.Load())
				}
			})
		}
	}
}
