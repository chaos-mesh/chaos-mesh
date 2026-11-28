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

package azurechaos_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/azurechaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/azurechaos/diskdetach"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/azurechaos/vmrestart"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/azurechaos/vmstop"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/azure"
)

func TestComputeLifecycle(t *testing.T) {
	g := NewWithT(t)
	const prefix = "/subscriptions/subscription/resourceGroups/group/providers/Microsoft.Compute"
	const vmPath = prefix + "/virtualMachines/vm"
	const diskPath = prefix + "/disks/data"
	type request struct {
		method, path string
		body         compute.VirtualMachine
	}
	var mu sync.Mutex
	var calls []request
	var denied string
	vmDisks := `"dataDisks":[{"lun":1,"name":"data"},{"lun":2,"name":"other"}]`
	setComputeState := func(failure, disks string) {
		mu.Lock()
		defer mu.Unlock()
		denied, vmDisks = failure, disks
	}
	takeCalls := func() []request {
		mu.Lock()
		defer mu.Unlock()
		got := calls
		calls = nil
		return got
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		got := request{method: r.Method, path: r.URL.Path}
		if r.Method == http.MethodPut {
			if err := json.NewDecoder(r.Body).Decode(&got.body); err != nil {
				t.Errorf("decode VM update: %v", err)
			}
		}
		calls = append(calls, got)
		w.Header().Set("Content-Type", "application/json")
		if r.Method+" "+r.URL.Path == denied {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"error":{"code":"Forbidden","message":"permission denied"}}`)
			return
		}
		switch r.Method + " " + r.URL.Path {
		case "GET " + vmPath:
			fmt.Fprintf(w, `{"location":"fixture","properties":{"storageProfile":{%s}}}`, vmDisks)
		case "GET " + diskPath:
			fmt.Fprintf(w, `{"id":%q}`, diskPath)
		default:
			fmt.Fprint(w, "{}")
		}
	}))
	t.Cleanup(server.Close)
	cleanup := mock.With("MockAzureComputeClient", func(base *compute.BaseClient) {
		base.BaseURI = server.URL
		base.Authorizer = autorest.NullAuthorizer{}
		base.Sender = server.Client()
	})
	t.Cleanup(func() { NewWithT(t).Expect(cleanup()).To(Succeed()) })

	ctx := t.Context()
	log := logr.Discard()
	kube := fake.NewClientBuilder().WithObjects(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "fixture"},
		Data:       map[string][]byte{"client_id": []byte("client"), "client_secret": []byte("fixture"), "tenant_id": []byte("tenant")},
	}).Build()
	stop, restart, detach := vmstop.NewImpl(kube, log), vmrestart.NewImpl(kube, log), diskdetach.NewImpl(kube, log)
	impl := azurechaos.NewImpl(azurechaos.Impl{VMStop: stop, VMRestart: restart, DiskDetach: detach}).Impl
	selector := v1alpha1.AzureSelector{
		SubscriptionID: "subscription", ResourceGroupName: "group", VMName: "vm",
		DiskName: new("data"), LUN: new(1), SecretName: new("fixture"),
	}
	targets, err := azure.New().Select(ctx, &selector)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(targets).To(HaveLen(1))
	records := []*v1alpha1.Record{{Id: targets[0].Id()}}
	invalid := []*v1alpha1.Record{{Id: "invalid"}}
	experiment := func(action v1alpha1.AzureChaosAction) *v1alpha1.AzureChaos {
		return &v1alpha1.AzureChaos{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
			Spec:       v1alpha1.AzureChaosSpec{Action: action, AzureSelector: selector},
		}
	}

	for _, tc := range []struct {
		action          v1alpha1.AzureChaosAction
		inject, recover string
	}{
		{v1alpha1.AzureVmStop, vmPath + "/powerOff", vmPath + "/start"},
		{v1alpha1.AzureVmRestart, vmPath + "/restart", ""},
	} {
		t.Run(string(tc.action), func(t *testing.T) {
			g := NewWithT(t)
			chaos := experiment(tc.action)
			setComputeState("", "")
			takeCalls()
			g.Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
			g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
			want := []request{{method: http.MethodPost, path: tc.inject}}
			if tc.recover != "" {
				want = append(want, request{method: http.MethodPost, path: tc.recover})
			}
			g.Expect(takeCalls()).To(Equal(want))

			setComputeState("POST "+tc.inject, "")
			phase, err := impl.Apply(ctx, 0, records, chaos)
			g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			phase, err = impl.Apply(ctx, 0, invalid, chaos)
			g.Expect(err).To(HaveOccurred())
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			if tc.recover != "" {
				setComputeState("POST "+tc.recover, "")
				phase, err = impl.Recover(ctx, 0, records, chaos)
				g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
				g.Expect(phase).To(Equal(v1alpha1.Injected))
				phase, err = impl.Recover(ctx, 0, invalid, chaos)
				g.Expect(err).To(HaveOccurred())
				g.Expect(phase).To(Equal(v1alpha1.NotInjected))
			}
		})
	}

	t.Run("detaches the selected disk and recovers when Azure omits the disk list", func(t *testing.T) {
		g := NewWithT(t)
		chaos := experiment(v1alpha1.AzureDiskDetach)
		initial := `"dataDisks":[{"lun":1,"name":"data"},{"lun":2,"name":"other"}]`
		setComputeState("", initial)
		takeCalls()
		g.Expect(impl.Apply(ctx, 0, records, chaos)).To(Equal(v1alpha1.Injected))
		got := takeCalls()
		g.Expect(got).To(HaveLen(2))
		g.Expect(got[0]).To(Equal(request{method: http.MethodGet, path: vmPath}))
		g.Expect(got[1].method).To(Equal(http.MethodPut))
		g.Expect(got[1].path).To(Equal(vmPath))
		g.Expect(*got[1].body.StorageProfile.DataDisks).To(Equal([]compute.DataDisk{{Lun: new(int32(2)), Name: new("other")}}))

		setComputeState("", "")
		g.Expect(impl.Recover(ctx, 0, records, chaos)).To(Equal(v1alpha1.NotInjected))
		got = takeCalls()
		g.Expect(got).To(HaveLen(3))
		g.Expect(got[0]).To(Equal(request{method: http.MethodGet, path: diskPath}))
		g.Expect(got[1]).To(Equal(request{method: http.MethodGet, path: vmPath}))
		g.Expect(got[2].method).To(Equal(http.MethodPut))
		g.Expect(got[2].path).To(Equal(vmPath))
		g.Expect(*got[2].body.StorageProfile.DataDisks).To(Equal([]compute.DataDisk{
			{Lun: new(int32(1)), Name: new("data"), CreateOption: compute.DiskCreateOptionTypesAttach, ManagedDisk: &compute.ManagedDiskParameters{ID: new(diskPath)}},
		}))

		phase, err := impl.Apply(ctx, 0, records, chaos)
		g.Expect(err).To(MatchError(ContainSubstring("is not attached")))
		g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		for _, failure := range []string{"GET " + vmPath, "PUT " + vmPath} {
			setComputeState(failure, initial)
			phase, err = impl.Apply(ctx, 0, records, chaos)
			g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		}
		for _, failure := range []string{"GET " + diskPath, "GET " + vmPath, "PUT " + vmPath} {
			setComputeState(failure, "")
			phase, err = impl.Recover(ctx, 0, records, chaos)
			g.Expect(err).To(MatchError(ContainSubstring("permission denied")))
			g.Expect(phase).To(Equal(v1alpha1.Injected))
		}
		setComputeState("", "")
		phase, err = impl.Apply(ctx, 0, invalid, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		takeCalls()
		phase, err = impl.Recover(ctx, 0, invalid, chaos)
		g.Expect(err).To(HaveOccurred())
		g.Expect(phase).To(Equal(v1alpha1.Injected))
		g.Expect(takeCalls()).To(BeEmpty())
	})

	t.Run("credential failures precede Compute requests", func(t *testing.T) {
		g := NewWithT(t)
		takeCalls()
		for _, direct := range []impltypes.ChaosImpl{stop, restart, detach} {
			phase, err := direct.Apply(ctx, 0, records, &v1alpha1.PodChaos{})
			g.Expect(err).To(HaveOccurred())
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		}
		for _, action := range []v1alpha1.AzureChaosAction{v1alpha1.AzureVmStop, v1alpha1.AzureVmRestart, v1alpha1.AzureDiskDetach} {
			chaos := experiment(action)
			for _, secret := range []*string{nil, new("missing")} {
				chaos.Spec.SecretName = secret
				phase, err := impl.Apply(ctx, 0, records, chaos)
				g.Expect(err).To(HaveOccurred())
				g.Expect(phase).To(Equal(v1alpha1.NotInjected))
				if action != v1alpha1.AzureVmRestart {
					phase, err = impl.Recover(ctx, 0, records, chaos)
					g.Expect(err).To(HaveOccurred())
					g.Expect(phase).To(Equal(v1alpha1.Injected))
				}
			}
		}
		g.Expect(takeCalls()).To(BeEmpty())
	})
}
