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

package trafficcontrol

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	selectorpod "github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

func TestBidirectionalOverlap(t *testing.T) {
	for _, targetFirst := range []bool{false, true} {
		name := "source first"
		if targetFirst {
			name = "target first"
		}
		t.Run(name, func(t *testing.T) {
			for _, targetDevice := range []string{"eth0", "net1"} {
				t.Run(targetDevice, func(t *testing.T) {
					f := newTrafficControlFixture(t, []string{"a", "overlap"}, []string{"overlap", "b"}, v1alpha1.Both, targetFirst)
					f.chaos.Spec.Device, f.chaos.Spec.TargetDevice = "eth0", targetDevice
					other := v1alpha1.RawTrafficControl{Type: v1alpha1.Netem, Source: "default/other", TcParameter: f.chaos.Spec.TcParameter}
					if err := f.client.Create(f.ctx, &v1alpha1.PodNetworkChaos{
						ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "overlap"},
						Spec:       v1alpha1.PodNetworkChaosSpec{TrafficControls: []v1alpha1.RawTrafficControl{other}},
					}); err != nil {
						t.Fatal(err)
					}
					for i, record := range f.records {
						f.apply(i)
						if record.Id == "default/overlap" {
							// Either role must commit both directions immediately.
							f.assertRules("overlap", []expectedTrafficControl{
								{device: "eth0", destinations: []string{"192.0.2.2/32", "192.0.2.3/32"}},
								{device: targetDevice, destinations: []string{"192.0.2.1/32", "192.0.2.2/32"}},
							})
						}
					}
					f.assertRules("a", []expectedTrafficControl{{device: "eth0", destinations: []string{"192.0.2.2/32", "192.0.2.3/32"}}})
					f.assertRules("b", []expectedTrafficControl{{device: targetDevice, destinations: []string{"192.0.2.1/32", "192.0.2.2/32"}}})

					// A lost status update can retry either selector role. Rebuilding
					// it must neither discard the other role nor append duplicates.
					beforeRetry := f.read("overlap")
					for i, record := range f.records {
						if record.Id == "default/overlap" {
							record.Phase = v1alpha1.NotInjected
							f.apply(i)
							got := f.read("overlap")
							if !reflect.DeepEqual(got.Spec, beforeRetry.Spec) || got.Generation != beforeRetry.Generation {
								t.Fatal("retry changed the complete pod configuration")
							}
						}
					}
					for i := range f.records {
						if f.apply(i) != waitForApplySync {
							t.Fatal("reported Injected before the pod configuration was observed")
						}
					}
					f.acknowledge()
					for i := range f.records {
						if f.apply(i) != v1alpha1.Injected {
							t.Fatal("expected Injected after the complete pod configuration was observed")
						}
					}

					// Both records recover the same Source. Repeating that operation
					// must be safe and preserve another experiment on the pod.
					for i, record := range f.records {
						phase, err := f.impl.Recover(f.ctx, i, f.records, f.chaos)
						if err != nil || phase != waitForRecoverSync {
							t.Fatalf("Recover returned %s, %v", phase, err)
						}
						record.Phase = phase
					}
					for _, pod := range []string{"a", "overlap", "b"} {
						f.assertRules(pod, nil)
					}
					if got := f.read("overlap").Spec.TrafficControls; !reflect.DeepEqual(got, []v1alpha1.RawTrafficControl{other}) {
						t.Fatalf("recovery changed the other experiment: %#v", got)
					}
					f.acknowledge()
					for i := range f.records {
						phase, err := f.impl.Recover(f.ctx, i, f.records, f.chaos)
						if err != nil || phase != v1alpha1.NotInjected {
							t.Fatalf("acknowledged recovery returned %s, %v", phase, err)
						}
					}
				})
			}
		})
	}
}

func TestTrafficControlWithoutOverlap(t *testing.T) {
	for _, direction := range []v1alpha1.Direction{v1alpha1.To, v1alpha1.From, v1alpha1.Both} {
		t.Run(string(direction), func(t *testing.T) {
			f := newTrafficControlFixture(t, []string{"a"}, []string{"b"}, direction, false)
			for i := range f.records {
				f.apply(i)
			}
			if direction != v1alpha1.From {
				f.assertRules("a", []expectedTrafficControl{{destinations: []string{"192.0.2.3/32"}}})
			} else if err := f.client.Get(f.ctx, client.ObjectKey{Namespace: "default", Name: "a"}, &v1alpha1.PodNetworkChaos{}); !apierrors.IsNotFound(err) {
				t.Fatalf("inactive source should not have a PodNetworkChaos: %v", err)
			}
			if direction != v1alpha1.To {
				f.assertRules("b", []expectedTrafficControl{{destinations: []string{"192.0.2.1/32"}}})
			} else if err := f.client.Get(f.ctx, client.ObjectKey{Namespace: "default", Name: "b"}, &v1alpha1.PodNetworkChaos{}); !apierrors.IsNotFound(err) {
				t.Fatalf("inactive target should not have a PodNetworkChaos: %v", err)
			}
		})
	}
}

type expectedTrafficControl struct {
	device       string
	destinations []string
}

type trafficControlFixture struct {
	t       *testing.T
	ctx     context.Context
	client  client.Client
	impl    *Impl
	chaos   *v1alpha1.NetworkChaos
	records []*v1alpha1.Record
}

func newTrafficControlFixture(t *testing.T, sources, targets []string, direction v1alpha1.Direction, targetFirst bool) *trafficControlFixture {
	t.Helper()
	f := &trafficControlFixture{t: t, ctx: context.Background()}
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	var pods []client.Object
	for name, ip := range map[string]string{"a": "192.0.2.1", "overlap": "192.0.2.2", "b": "192.0.2.3"} {
		pods = append(pods, &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: name}, Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: ip}})
	}
	f.client = &generationClient{Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(pods...).WithStatusSubresource(&v1alpha1.PodNetworkChaos{}).Build()}
	selector := func(names []string) v1alpha1.PodSelector {
		return v1alpha1.PodSelector{Mode: v1alpha1.AllMode, Selector: v1alpha1.PodSelectorSpec{Pods: map[string][]string{"default": names}}}
	}
	target := selector(targets)
	f.chaos = &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "delay"},
		Spec: v1alpha1.NetworkChaosSpec{
			Action: v1alpha1.DelayAction, Direction: direction, PodSelector: selector(sources), Target: &target,
			TcParameter: v1alpha1.TcParameter{Delay: &v1alpha1.DelaySpec{Latency: "200ms"}},
		},
	}
	if err := f.chaos.Default(f.ctx, f.chaos); err != nil {
		t.Fatal(err)
	}
	if _, err := f.chaos.ValidateCreate(f.ctx, f.chaos); err != nil {
		t.Fatalf("invalid test experiment: %v", err)
	}
	keys := []string{".", ".Target"}
	if targetFirst {
		slices.Reverse(keys)
	}
	for _, key := range keys {
		selected, err := selectorpod.SelectAndFilterPods(f.ctx, f.client, f.client, f.chaos.GetSelectorSpecs()[key].(*v1alpha1.PodSelector), true, "", false)
		if err != nil {
			t.Fatal(err)
		}
		for _, pod := range selected {
			f.records = append(f.records, &v1alpha1.Record{Id: pod.Namespace + "/" + pod.Name, SelectorKey: key, Phase: v1alpha1.NotInjected})
		}
	}
	builder := podnetworkchaosmanager.NewBuilder(podnetworkchaosmanager.Params{Client: f.client, Reader: f.client, Scheme: scheme, Logger: logr.Discard()})
	f.impl = NewImpl(f.client, builder, logr.Discard())
	return f
}

func (f *trafficControlFixture) apply(index int) v1alpha1.Phase {
	f.t.Helper()
	phase, err := f.impl.Apply(f.ctx, index, f.records, f.chaos)
	if err != nil {
		f.t.Fatal(err)
	}
	f.records[index].Phase = phase
	return phase
}

func (f *trafficControlFixture) read(name string) *v1alpha1.PodNetworkChaos {
	f.t.Helper()
	obj := &v1alpha1.PodNetworkChaos{}
	if err := f.client.Get(f.ctx, client.ObjectKey{Namespace: "default", Name: name}, obj); err != nil {
		f.t.Fatal(err)
	}
	return obj
}

func (f *trafficControlFixture) acknowledge() {
	f.t.Helper()
	var objects v1alpha1.PodNetworkChaosList
	if err := f.client.List(f.ctx, &objects); err != nil {
		f.t.Fatal(err)
	}
	for _, obj := range objects.Items {
		obj.Status.ObservedGeneration = obj.Generation
		if err := f.client.Status().Update(f.ctx, &obj); err != nil {
			f.t.Fatal(err)
		}
	}
}

func (f *trafficControlFixture) assertRules(pod string, want []expectedTrafficControl) {
	f.t.Helper()
	obj := f.read(pod)
	sets := map[string]v1alpha1.RawIPSet{}
	for _, set := range obj.Spec.IPSets {
		if _, ok := sets[set.Name]; ok {
			f.t.Fatalf("duplicate IP set %s", set.Name)
		}
		sets[set.Name] = set
	}
	var got []expectedTrafficControl
	for _, tc := range obj.Spec.TrafficControls {
		if tc.Source != "default/delay" {
			continue
		}
		if tc.Type != v1alpha1.Netem || !reflect.DeepEqual(tc.TcParameter, f.chaos.Spec.TcParameter) {
			f.t.Fatalf("traffic control changed the requested fault: %#v", tc)
		}
		var destinations []string
		group, ok := sets[tc.IPSet]
		if !ok || group.IPSetType != v1alpha1.SetIPSet || len(group.SetNames) != 2 {
			f.t.Fatalf("invalid traffic control filter: %#v", tc)
		}
		for _, name := range group.SetNames {
			set, ok := sets[name]
			if !ok || set.Source != tc.Source {
				f.t.Fatalf("missing or unowned IP set %s", name)
			}
			destinations = append(destinations, set.Cidrs...)
		}
		slices.Sort(destinations)
		got = append(got, expectedTrafficControl{device: tc.Device, destinations: destinations})
	}
	if !reflect.DeepEqual(got, want) || len(sets) != 3*len(want) {
		f.t.Fatalf("pod %s traffic controls = %#v (%d sets), want %#v", pod, got, len(sets), want)
	}
}

// The fake client does not increment generations. Model that API-server behavior
// to exercise the controller's wait phases without an envtest cluster.
type generationClient struct {
	client.Client
}

func (c *generationClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if _, ok := obj.(*v1alpha1.PodNetworkChaos); ok {
		obj.SetGeneration(1)
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *generationClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if chaos, ok := obj.(*v1alpha1.PodNetworkChaos); ok {
		old := &v1alpha1.PodNetworkChaos{}
		if err := c.Client.Get(ctx, client.ObjectKeyFromObject(obj), old); err != nil {
			return err
		}
		// Compare the API representation: omitempty fields can be empty slices
		// in a transaction and nil after the fake client's serialization.
		oldSpec, err := json.Marshal(old.Spec)
		if err != nil {
			return err
		}
		newSpec, err := json.Marshal(chaos.Spec)
		if err != nil {
			return err
		}
		if !bytes.Equal(oldSpec, newSpec) {
			chaos.Generation = old.Generation + 1
		}
	}
	return c.Client.Update(ctx, obj, opts...)
}
