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

package envoygatewaychaos

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestBuildPolicySpecForGRPCRoute(t *testing.T) {
	g := NewWithT(t)
	grpcStatus := int32(14)
	chaos := testEnvoyGatewayChaos()
	chaos.Spec.Target.Kind = v1alpha1.EnvoyGatewayGRPCRoute
	chaos.Spec.Fault = v1alpha1.EnvoyGatewayFault{
		Abort: &v1alpha1.EnvoyGatewayAbort{GRPCStatus: &grpcStatus, Percentage: 25},
	}

	spec := buildPolicySpec(chaos)
	status, found, err := unstructured.NestedInt64(spec, "faultInjection", "abort", "grpcStatus")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(found).To(BeTrue())
	g.Expect(status).To(Equal(int64(14)))
	_, found, err = unstructured.NestedInt64(spec, "faultInjection", "abort", "httpStatus")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(found).To(BeFalse())
}

func TestApplyAndRecover(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	chaos := testEnvoyGatewayChaos()
	impl := testImpl(testRoute(v1alpha1.EnvoyGatewayHTTPRoute, map[string]string{"app": "api"}))

	phase, err := impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	policy := newBackendTrafficPolicy()
	key := types.NamespacedName{Namespace: "app", Name: managedPolicyName(chaos.Spec.Target)}
	g.Expect(impl.Get(ctx, key, policy)).To(Succeed())
	g.Expect(policy.GetAnnotations()).To(HaveKeyWithValue(ownerAnnotation, ownerIdentity(chaos)))
	spec, found, err := unstructured.NestedMap(policy.Object, "spec")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(found).To(BeTrue())
	g.Expect(spec).To(Equal(buildPolicySpec(chaos)))

	phase, err = impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	phase, err = impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
	g.Expect(apierrors.IsNotFound(impl.Get(ctx, key, policy))).To(BeTrue())
}

func TestApplyRestoresOwnedPolicy(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	chaos := testEnvoyGatewayChaos()
	impl := testImpl(testRoute(v1alpha1.EnvoyGatewayHTTPRoute, nil))

	phase, err := impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	policy := newBackendTrafficPolicy()
	key := types.NamespacedName{Namespace: "app", Name: managedPolicyName(chaos.Spec.Target)}
	g.Expect(impl.Get(ctx, key, policy)).To(Succeed())
	g.Expect(unstructured.SetNestedField(policy.Object, int64(500), "spec", "faultInjection", "abort", "httpStatus")).To(Succeed())
	g.Expect(impl.Update(ctx, policy)).To(Succeed())

	phase, err = impl.Apply(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.Injected))
	g.Expect(impl.Get(ctx, key, policy)).To(Succeed())
	spec, _, err := unstructured.NestedMap(policy.Object, "spec")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(spec).To(Equal(buildPolicySpec(chaos)))
}

func TestApplyRejectsPolicyConflicts(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		policy *unstructured.Unstructured
	}{
		{
			name: "direct reference",
			policy: testPolicy("existing", map[string]interface{}{
				"targetRefs": []interface{}{map[string]interface{}{
					"group": gatewayAPIGroup,
					"kind":  "HTTPRoute",
					"name":  "api",
				}},
			}),
		},
		{
			name: "deprecated direct reference",
			policy: testPolicy("existing", map[string]interface{}{
				"targetRef": map[string]interface{}{
					"group": gatewayAPIGroup,
					"kind":  "HTTPRoute",
					"name":  "api",
				},
			}),
		},
		{
			name: "label selector",
			policy: testPolicy("existing", map[string]interface{}{
				"targetSelectors": []interface{}{map[string]interface{}{
					"group": gatewayAPIGroup,
					"kind":  "HTTPRoute",
					"matchLabels": map[string]interface{}{
						"app": "api",
					},
				}},
			}),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			impl := testImpl(testRoute(v1alpha1.EnvoyGatewayHTTPRoute, map[string]string{"app": "api"}), testCase.policy)

			phase, err := impl.Apply(context.Background(), 0, nil, testEnvoyGatewayChaos())
			g.Expect(err).To(MatchError(ContainSubstring("already targeted")))
			g.Expect(phase).To(Equal(v1alpha1.NotInjected))
		})
	}
}

func TestApplyRejectsForeignManagedPolicy(t *testing.T) {
	g := NewWithT(t)
	chaos := testEnvoyGatewayChaos()
	policy := testPolicy(managedPolicyName(chaos.Spec.Target), buildPolicySpec(chaos))
	policy.SetAnnotations(map[string]string{ownerAnnotation: "other/experiment:uid"})
	impl := testImpl(testRoute(v1alpha1.EnvoyGatewayHTTPRoute, nil), policy)

	phase, err := impl.Apply(context.Background(), 0, nil, chaos)
	g.Expect(err).To(MatchError(ContainSubstring("not owned")))
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func TestApplyRequiresTargetRoute(t *testing.T) {
	g := NewWithT(t)
	impl := testImpl()

	phase, err := impl.Apply(context.Background(), 0, nil, testEnvoyGatewayChaos())
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func TestRecoverIsOwnershipSafe(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	chaos := testEnvoyGatewayChaos()
	policy := testPolicy(managedPolicyName(chaos.Spec.Target), buildPolicySpec(chaos))
	policy.SetAnnotations(map[string]string{ownerAnnotation: "other/experiment:uid"})
	impl := testImpl(policy)

	phase, err := impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).To(MatchError(ContainSubstring("not owned")))
	g.Expect(phase).To(Equal(v1alpha1.Injected))

	g.Expect(impl.Delete(ctx, policy)).To(Succeed())
	phase, err = impl.Recover(ctx, 0, nil, chaos)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(phase).To(Equal(v1alpha1.NotInjected))
}

func testImpl(objects ...client.Object) *Impl {
	runtimeObjects := make([]client.Object, len(objects))
	copy(runtimeObjects, objects)
	return &Impl{
		Client: fake.NewClientBuilder().WithObjects(runtimeObjects...).Build(),
		Log:    logr.Discard(),
	}
}

func testEnvoyGatewayChaos() *v1alpha1.EnvoyGatewayChaos {
	httpStatus := int32(503)
	return &v1alpha1.EnvoyGatewayChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-fault",
			Namespace: "chaos-mesh",
			UID:       types.UID("9fe3d51f-bfaa-4a61-9139-ec0719063a72"),
		},
		Spec: v1alpha1.EnvoyGatewayChaosSpec{
			Target: v1alpha1.EnvoyGatewayTarget{
				Namespace: "app",
				Kind:      v1alpha1.EnvoyGatewayHTTPRoute,
				Route:     "api",
			},
			Fault: v1alpha1.EnvoyGatewayFault{
				Delay: &v1alpha1.EnvoyGatewayDelay{FixedDelay: "250ms", Percentage: 20},
				Abort: &v1alpha1.EnvoyGatewayAbort{HTTPStatus: &httpStatus, Percentage: 10},
			},
		},
	}
}

func testRoute(kind v1alpha1.EnvoyGatewayRouteKind, routeLabels map[string]string) *unstructured.Unstructured {
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{Group: gatewayAPIGroup, Version: "v1", Kind: string(kind)})
	route.SetName("api")
	route.SetNamespace("app")
	route.SetLabels(routeLabels)
	return route
}

func testPolicy(name string, spec map[string]interface{}) *unstructured.Unstructured {
	policy := newBackendTrafficPolicy()
	policy.SetName(name)
	policy.SetNamespace("app")
	policy.Object["spec"] = spec
	return policy
}
