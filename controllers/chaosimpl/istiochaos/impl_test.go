// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istiochaos

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestApplyAndRecover(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	virtualService := testVirtualService(
		testRoute("health", "/health"),
		testRoute("service", "/api"),
	)
	chaos := testIstioChaos()
	c := fake.NewClientBuilder().WithRuntimeObjects(virtualService).Build()
	impl := &Impl{Client: c, Log: logr.Discard()}

	phase, err := impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	updated := newVirtualService()
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "app", Name: "service"}, updated)).To(Succeed())
	routes, found, err := unstructured.NestedSlice(updated.Object, "spec", "http")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(found).To(BeTrue())
	g.Expect(routes).To(HaveLen(3))
	g.Expect(routeName(routes[0])).To(Equal("health"))
	g.Expect(routeName(routes[1])).To(Equal(managedRouteName(chaos)))
	g.Expect(routeName(routes[2])).To(Equal("service"))
	g.Expect(routes[1].(map[string]interface{})["match"]).To(Equal(routes[2].(map[string]interface{})["match"]))
	g.Expect(routes[1].(map[string]interface{})["route"]).To(Equal(routes[2].(map[string]interface{})["route"]))
	g.Expect(routes[1].(map[string]interface{})["fault"]).To(Equal(map[string]interface{}{
		"abort": map[string]interface{}{
			"httpStatus": int64(503),
			"percentage": map[string]interface{}{"value": int64(10)},
		},
		"delay": map[string]interface{}{
			"fixedDelay": "250ms",
			"percentage": map[string]interface{}{"value": int64(20)},
		},
	}))
	g.Expect(updated.GetAnnotations()).To(HaveKeyWithValue(ownerAnnotation, ownerIdentity(chaos)))

	phase, err = impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "app", Name: "service"}, updated)).To(Succeed())
	routes, _, err = unstructured.NestedSlice(updated.Object, "spec", "http")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(routes).To(HaveLen(3), "a repeated apply must not duplicate the managed route")

	phase, err = impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "app", Name: "service"}, updated)).To(Succeed())
	routes, _, err = unstructured.NestedSlice(updated.Object, "spec", "http")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(routes).To(HaveLen(2))
	g.Expect(routeName(routes[0])).To(Equal("health"))
	g.Expect(routeName(routes[1])).To(Equal("service"))
	g.Expect(updated.GetAnnotations()).NotTo(HaveKey(ownerAnnotation))

	phase, err = impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func TestApplyRejectsAnotherOwner(t *testing.T) {
	g := NewWithT(t)
	virtualService := testVirtualService(testRoute("service", "/api"))
	virtualService.SetAnnotations(map[string]string{ownerAnnotation: "chaos-mesh/another-fault"})
	c := fake.NewClientBuilder().WithRuntimeObjects(virtualService).Build()
	impl := &Impl{Client: c, Log: logr.Discard()}

	phase, err := impl.Apply(context.Background(), 0, nil, testIstioChaos())
	g.Expect(err).To(MatchError(ContainSubstring("already controlled")))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func TestWritesUseOptimisticLockPatches(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	c := &recordingClient{
		Client: fake.NewClientBuilder().WithRuntimeObjects(
			testVirtualService(testRoute("service", "/api")),
		).Build(),
	}
	impl := &Impl{Client: c, Log: logr.Discard()}
	chaos := testIstioChaos()

	phase, err := impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	phase, err = impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))

	g.Expect(c.patches).To(HaveLen(2))
	for _, patch := range c.patches {
		g.Expect(string(patch)).To(ContainSubstring(`"resourceVersion"`))
	}
}

func TestApplyRejectsRecreatedChaosWithSameName(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	virtualService := testVirtualService(testRoute("service", "/api"))
	original := testIstioChaos()
	c := fake.NewClientBuilder().WithRuntimeObjects(virtualService).Build()
	impl := &Impl{Client: c, Log: logr.Discard()}

	phase, err := impl.Apply(ctx, 0, nil, original)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	recreated := original.DeepCopy()
	recreated.UID = types.UID("b78e634c-dfab-4af0-99c1-158238be6841")
	phase, err = impl.Apply(ctx, 0, nil, recreated)
	g.Expect(err).To(MatchError(ContainSubstring("already controlled")))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))

	updated := newVirtualService()
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "app", Name: "service"}, updated)).To(Succeed())
	routes, _, err := unstructured.NestedSlice(updated.Object, "spec", "http")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(routes).To(HaveLen(2), "a recreated experiment must not add another managed route")
	g.Expect(routeName(routes[0])).To(Equal(managedRouteName(original)))
	g.Expect(updated.GetAnnotations()).To(HaveKeyWithValue(ownerAnnotation, ownerIdentity(original)))
}

func TestApplySelectsTheOnlyUnnamedRoute(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	virtualService := testVirtualService(testRoute("", "/api"))
	chaos := testIstioChaos()
	chaos.Spec.Target.HTTPRoute = ""
	c := fake.NewClientBuilder().WithRuntimeObjects(virtualService).Build()
	impl := &Impl{Client: c, Log: logr.Discard()}

	phase, err := impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	updated := newVirtualService()
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "app", Name: "service"}, updated)).To(Succeed())
	routes, _, err := unstructured.NestedSlice(updated.Object, "spec", "http")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(routes).To(HaveLen(2))
	g.Expect(routeName(routes[0])).To(Equal(managedRouteName(chaos)))
	g.Expect(routeName(routes[1])).To(BeEmpty())

	phase, err = impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	g.Expect(c.Get(ctx, types.NamespacedName{Namespace: "app", Name: "service"}, updated)).To(Succeed())
	routes, _, err = unstructured.NestedSlice(updated.Object, "spec", "http")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(routes).To(HaveLen(1))
	g.Expect(routeName(routes[0])).To(BeEmpty())
}

func TestApplyValidatesRouteSelection(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		routes    []interface{}
		httpRoute string
		err       string
	}{
		{name: "missing named route", routes: []interface{}{testRoute("other", "/")}, httpRoute: "service", err: "not found"},
		{name: "duplicate named route", routes: []interface{}{testRoute("service", "/one"), testRoute("service", "/two")}, httpRoute: "service", err: "multiple HTTP routes"},
		{name: "multiple routes without name", routes: []interface{}{testRoute("", "/one"), testRoute("", "/two")}, err: "httpRoute must be set"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			virtualService := testVirtualService(testCase.routes...)
			c := fake.NewClientBuilder().WithRuntimeObjects(virtualService).Build()
			impl := &Impl{Client: c, Log: logr.Discard()}
			chaos := testIstioChaos()
			chaos.Spec.Target.HTTPRoute = testCase.httpRoute

			phase, err := impl.Apply(context.Background(), 0, nil, chaos)
			g.Expect(err).To(MatchError(ContainSubstring(testCase.err)))
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		})
	}
}

func TestRecoverSucceedsWhenVirtualServiceIsGone(t *testing.T) {
	g := NewWithT(t)
	impl := &Impl{Client: fake.NewClientBuilder().Build(), Log: logr.Discard()}

	phase, err := impl.Recover(context.Background(), 0, nil, testIstioChaos())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func testIstioChaos() *v1alpha1.IstioChaos {
	return &v1alpha1.IstioChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service-fault",
			Namespace: "chaos-mesh",
			UID:       types.UID("9fe3d51f-bfaa-4a61-9139-ec0719063a72"),
		},
		Spec: v1alpha1.IstioChaosSpec{
			Target: v1alpha1.IstioTarget{
				Namespace:      "app",
				VirtualService: "service",
				HTTPRoute:      "service",
			},
			Fault: v1alpha1.IstioFault{
				Delay: &v1alpha1.IstioDelay{FixedDelay: "250ms", Percentage: 20},
				Abort: &v1alpha1.IstioAbort{HTTPStatus: 503, Percentage: 10},
			},
		},
	}
}

func testVirtualService(routes ...interface{}) *unstructured.Unstructured {
	virtualService := newVirtualService()
	virtualService.SetName("service")
	virtualService.SetNamespace("app")
	virtualService.Object["spec"] = map[string]interface{}{
		"hosts": []interface{}{"service.app.svc.cluster.local"},
		"http":  runtime.DeepCopyJSONValue(routes),
	}
	return virtualService
}

func testRoute(name, prefix string) map[string]interface{} {
	route := map[string]interface{}{
		"match": []interface{}{
			map[string]interface{}{"uri": map[string]interface{}{"prefix": prefix}},
		},
		"route": []interface{}{
			map[string]interface{}{"destination": map[string]interface{}{"host": "service"}},
		},
	}
	if name != "" {
		route["name"] = name
	}
	return route
}

func routeName(value interface{}) string {
	name, _, _ := unstructured.NestedString(value.(map[string]interface{}), "name")
	return name
}

type recordingClient struct {
	client.Client
	patches [][]byte
}

func (c *recordingClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	data, err := patch.Data(obj)
	if err != nil {
		return err
	}
	c.patches = append(c.patches, data)
	return c.Client.Patch(ctx, obj, patch, opts...)
}
