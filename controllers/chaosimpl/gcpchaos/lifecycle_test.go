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

package gcpchaos_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"google.golang.org/api/option"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/gcpchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/gcpchaos/diskloss"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/gcpchaos/nodereset"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/gcpchaos/nodestop"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/gcp"
)

func TestComputeLifecycle(t *testing.T) {
	g := NewWithT(t)
	const instancePath = "/projects/project-fixture/zones/zone-fixture/instances/instance-fixture"
	const disk = `{"deviceName":"disk-a","source":"projects/project-fixture/zones/zone-fixture/disks/disk-a","mode":"READ_WRITE"}`
	var mu sync.Mutex
	denied := ""
	var calls []string
	setDenied := func(path string) {
		mu.Lock()
		defer mu.Unlock()
		denied = path
	}
	takeCalls := func() []string {
		mu.Lock()
		defer mu.Unlock()
		got := calls
		calls = nil
		return got
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, r.Method+" "+r.URL.Path)
		forbidden := r.URL.Path == denied
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if forbidden {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":{"message":"permission denied"}}`)
			return
		}
		switch r.URL.Path {
		case instancePath:
			fmt.Fprintf(w, `{"disks":[{"deviceName":"other"},%s]}`, disk)
		case instancePath + "/detachDisk":
			if got := r.URL.Query().Get("deviceName"); got != "disk-a" {
				t.Errorf("detach requested disk %q", got)
			}
			fmt.Fprint(w, "{}")
		case instancePath + "/attachDisk":
			var got json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Errorf("decode attached disk: %v", err)
			}
			if ok, err := MatchJSON(disk).Match(got); err != nil || !ok {
				t.Errorf("unexpected attached disk: %s (%v)", got, err)
			}
			fmt.Fprint(w, "{}")
		default:
			fmt.Fprint(w, "{}")
		}
	}))
	t.Cleanup(server.Close)
	cleanup := mock.With("MockGCPComputeOptions", []option.ClientOption{option.WithEndpoint(server.URL + "/"), option.WithoutAuthentication()})
	t.Cleanup(func() { NewWithT(t).Expect(cleanup()).To(Succeed()) })

	ctx := t.Context()
	log := logr.Discard()
	kube := fake.NewClientBuilder().WithObjects(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "bad-credentials"},
		Data:       map[string][]byte{"service_account": []byte("not base64!")},
	}).Build()
	vmStop, vmReset, diskLoss := nodestop.NewImpl(kube, log), nodereset.NewImpl(kube, log), diskloss.NewImpl(kube, log)
	impl := gcpchaos.NewImpl(gcpchaos.Impl{NodeStop: vmStop, NodeReset: vmReset, DiskLoss: diskLoss}).Impl
	selector := v1alpha1.GCPSelector{Project: "project-fixture", Zone: "zone-fixture", Instance: "instance-fixture", DeviceNames: []string{"disk-a"}}
	targets, err := gcp.New().Select(ctx, &selector)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(targets).To(HaveLen(1))
	records := []*v1alpha1.Record{{Id: targets[0].Id()}}
	invalid := []*v1alpha1.Record{{Id: "invalid"}}

	for _, tc := range []struct {
		action  v1alpha1.GCPChaosAction
		inject  string
		recover string
	}{
		{v1alpha1.NodeStop, "/stop", "/start"},
		{v1alpha1.NodeReset, "/reset", ""},
	} {
		t.Run(string(tc.action), func(t *testing.T) {
			g := NewWithT(t)
			chaos := &v1alpha1.GCPChaos{Spec: v1alpha1.GCPChaosSpec{Action: tc.action}}
			takeCalls()
			g.Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
			g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
			expected := []string{"POST " + instancePath + tc.inject}
			if tc.recover != "" {
				expected = append(expected, "POST "+instancePath+tc.recover)
				setDenied(instancePath + tc.recover)
				phase, err := impl.Recover(ctx, 0, records, chaos)
				g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
				g.Expect(phase).To(Equal(v1alpha1.Injected))
				expected = append(expected, "POST "+instancePath+tc.recover)
				phase, err = impl.Recover(ctx, 0, invalid, chaos)
				g.Expect(err).To(HaveOccurred())
				g.Expect(phase).To(Equal(v1alpha1.Injected))
			}
			g.Expect(takeCalls()).To(Equal(expected))
			setDenied(instancePath + tc.inject)
			phase, err := impl.Apply(ctx, 0, records, chaos)
			g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			setDenied("")
			phase, err = impl.Apply(ctx, 0, invalid, chaos)
			g.Expect(err).To(HaveOccurred())
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		})
	}

	t.Run("disk loss retains the disk description for recovery", func(t *testing.T) {
		g := NewWithT(t)
		chaos := &v1alpha1.GCPChaos{Spec: v1alpha1.GCPChaosSpec{Action: v1alpha1.DiskLoss}}
		takeCalls()
		g.Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
		g.Expect(chaos.Status.AttachedDisksStrings).To(HaveLen(1))
		g.Expect(chaos.Status.AttachedDisksStrings[0]).To(MatchJSON(disk))
		g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
		g.Expect(takeCalls()).To(Equal([]string{"GET " + instancePath, "POST " + instancePath + "/detachDisk", "POST " + instancePath + "/attachDisk"}))
		for _, failure := range []string{instancePath, instancePath + "/detachDisk", instancePath + "/attachDisk"} {
			setDenied(failure)
			if failure == instancePath+"/attachDisk" {
				phase, err := impl.Recover(ctx, 0, records, chaos)
				g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
				g.Expect(phase).To(Equal(v1alpha1.Injected))
			} else {
				phase, err := impl.Apply(ctx, 0, records, chaos)
				g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
				g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			}
		}
		setDenied("")
		absent := selector
		absent.DeviceNames = []string{"missing"}
		absentTargets, err := gcp.New().Select(ctx, &absent)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(absentTargets).To(HaveLen(1))
		phase, err := impl.Apply(ctx, 0, []*v1alpha1.Record{{Id: absentTargets[0].Id()}}, chaos)
		g.Expect(err).To(MatchError(ContainSubstring("missing")))
		g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		phase, err = impl.Apply(ctx, 0, invalid, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		phase, err = impl.Recover(ctx, 0, invalid, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		chaos.Status.AttachedDisksStrings = []string{"invalid"}
		phase, err = impl.Recover(ctx, 0, records, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.Injected))
	})

	t.Run("credentials and object errors are reported before contacting Compute", func(t *testing.T) {
		g := NewWithT(t)
		takeCalls()
		for _, direct := range []impltypes.ChaosImpl{vmStop, vmReset, diskLoss} {
			phase, err := direct.Apply(ctx, 0, records, &v1alpha1.PodChaos{})
			g.Expect(err).To(HaveOccurred())
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		}
		for _, name := range []string{"missing", "bad-credentials"} {
			for _, action := range []v1alpha1.GCPChaosAction{v1alpha1.NodeStop, v1alpha1.NodeReset, v1alpha1.DiskLoss} {
				chaos := &v1alpha1.GCPChaos{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
					Spec:       v1alpha1.GCPChaosSpec{Action: action, SecretName: &name},
				}
				phase, err := impl.Apply(ctx, 0, records, chaos)
				g.Expect(err).To(HaveOccurred())
				g.Expect(phase).To(Equal(v1alpha1.NotInjected))
				if action != v1alpha1.NodeReset {
					phase, err = impl.Recover(ctx, 0, records, chaos)
					g.Expect(err).To(HaveOccurred())
					g.Expect(phase).To(Equal(v1alpha1.Injected))
				}
			}
		}
		g.Expect(takeCalls()).To(BeEmpty())
	})
}
