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
	"crypto/sha256"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

const (
	gatewayAPIGroup  = "gateway.networking.k8s.io"
	ownerAnnotation  = "chaos-mesh.org/envoy-gateway-chaos-owner"
	policyNamePrefix = "chaos-mesh-envoy-"
)

var (
	backendTrafficPolicyGVK = schema.GroupVersionKind{
		Group:   "gateway.envoyproxy.io",
		Version: "v1alpha1",
		Kind:    "BackendTrafficPolicy",
	}
	backendTrafficPolicyListGVK = schema.GroupVersionKind{
		Group:   "gateway.envoyproxy.io",
		Version: "v1alpha1",
		Kind:    "BackendTrafficPolicyList",
	}
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

// Impl manages Envoy Gateway fault policies.
type Impl struct {
	client.Client
	Log logr.Logger
}

// Apply creates or reconciles the owned BackendTrafficPolicy.
func (impl *Impl) Apply(ctx context.Context, _ int, _ []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.EnvoyGatewayChaos)
	route, err := impl.getTargetRoute(ctx, chaos.Spec.Target)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	policies := newBackendTrafficPolicyList()
	if err := impl.Client.List(ctx, policies, client.InNamespace(chaos.Spec.Target.Namespace)); err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "list BackendTrafficPolicies")
	}

	name := managedPolicyName(chaos.Spec.Target)
	owner := ownerIdentity(chaos)
	var managed *unstructured.Unstructured
	for index := range policies.Items {
		policy := &policies.Items[index]
		if policy.GetName() == name {
			if policy.GetAnnotations()[ownerAnnotation] != owner {
				return v1alpha1.NotInjected, errors.Errorf("BackendTrafficPolicy %s/%s is not owned by this EnvoyGatewayChaos", policy.GetNamespace(), policy.GetName())
			}
			managed = policy
			continue
		}

		matches, err := policyTargetsRoute(policy, route, chaos.Spec.Target)
		if err != nil {
			return v1alpha1.NotInjected, errors.Wrapf(err, "inspect BackendTrafficPolicy %s/%s", policy.GetNamespace(), policy.GetName())
		}
		if matches {
			return v1alpha1.NotInjected, errors.Errorf(
				"route %s/%s is already targeted by BackendTrafficPolicy %s",
				chaos.Spec.Target.Namespace,
				chaos.Spec.Target.Route,
				policy.GetName(),
			)
		}
	}

	expectedSpec := buildPolicySpec(chaos)
	if managed != nil {
		actualSpec, _, err := unstructured.NestedMap(managed.Object, "spec")
		if err != nil {
			return v1alpha1.NotInjected, errors.Wrap(err, "read managed BackendTrafficPolicy spec")
		}
		if !reflect.DeepEqual(actualSpec, expectedSpec) {
			origin := managed.DeepCopy()
			managed.Object["spec"] = runtime.DeepCopyJSONValue(expectedSpec)
			if err := impl.Client.Patch(ctx, managed, client.MergeFromWithOptions(origin, client.MergeFromWithOptimisticLock{})); err != nil {
				return v1alpha1.NotInjected, errors.Wrap(err, "restore managed BackendTrafficPolicy")
			}
		}
		return v1alpha1.Injected, nil
	}

	policy := newBackendTrafficPolicy()
	policy.SetName(name)
	policy.SetNamespace(chaos.Spec.Target.Namespace)
	policy.SetAnnotations(map[string]string{ownerAnnotation: owner})
	policy.Object["spec"] = expectedSpec
	if err := impl.Client.Create(ctx, policy); err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "create BackendTrafficPolicy")
	}

	impl.Log.Info("created Envoy Gateway fault policy", "policy", types.NamespacedName{Namespace: policy.GetNamespace(), Name: policy.GetName()})
	return v1alpha1.Injected, nil
}

// Recover removes the owned BackendTrafficPolicy.
func (impl *Impl) Recover(ctx context.Context, _ int, _ []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.EnvoyGatewayChaos)
	policy := newBackendTrafficPolicy()
	key := types.NamespacedName{
		Namespace: chaos.Spec.Target.Namespace,
		Name:      managedPolicyName(chaos.Spec.Target),
	}
	if err := impl.Client.Get(ctx, key, policy); err != nil {
		if apierrors.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, errors.Wrap(err, "get managed BackendTrafficPolicy")
	}

	owner := ownerIdentity(chaos)
	if policy.GetAnnotations()[ownerAnnotation] != owner {
		return v1alpha1.Injected, errors.Errorf("BackendTrafficPolicy %s is not owned by this EnvoyGatewayChaos", key)
	}
	if err := impl.Client.Delete(ctx, policy); err != nil && !apierrors.IsNotFound(err) {
		return v1alpha1.Injected, errors.Wrap(err, "delete managed BackendTrafficPolicy")
	}

	impl.Log.Info("removed Envoy Gateway fault policy", "policy", key)
	return v1alpha1.NotInjected, nil
}

func (impl *Impl) getTargetRoute(ctx context.Context, target v1alpha1.EnvoyGatewayTarget) (*unstructured.Unstructured, error) {
	route := &unstructured.Unstructured{}
	route.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gatewayAPIGroup,
		Version: "v1",
		Kind:    string(target.Kind),
	})
	key := types.NamespacedName{Namespace: target.Namespace, Name: target.Route}
	if err := impl.Client.Get(ctx, key, route); err != nil {
		return nil, errors.Wrapf(err, "get target %s", target.Kind)
	}
	return route, nil
}

func policyTargetsRoute(policy, route *unstructured.Unstructured, target v1alpha1.EnvoyGatewayTarget) (bool, error) {
	if ref, found, err := unstructured.NestedMap(policy.Object, "spec", "targetRef"); err != nil {
		return false, err
	} else if found && referenceMatches(ref, target) {
		return true, nil
	}

	refs, found, err := unstructured.NestedSlice(policy.Object, "spec", "targetRefs")
	if err != nil {
		return false, err
	}
	if found {
		for index, value := range refs {
			ref, ok := value.(map[string]interface{})
			if !ok {
				return false, errors.Errorf("spec.targetRefs[%d] is not an object", index)
			}
			if referenceMatches(ref, target) {
				return true, nil
			}
		}
	}

	selectors, found, err := unstructured.NestedSlice(policy.Object, "spec", "targetSelectors")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	for index, value := range selectors {
		selector, ok := value.(map[string]interface{})
		if !ok {
			return false, errors.Errorf("spec.targetSelectors[%d] is not an object", index)
		}
		matches, err := targetSelectorMatches(selector, route, target)
		if err != nil {
			return false, errors.Wrapf(err, "read spec.targetSelectors[%d]", index)
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

func referenceMatches(ref map[string]interface{}, target v1alpha1.EnvoyGatewayTarget) bool {
	group, _, _ := unstructured.NestedString(ref, "group")
	kind, _, _ := unstructured.NestedString(ref, "kind")
	name, _, _ := unstructured.NestedString(ref, "name")
	return (group == "" || group == gatewayAPIGroup) && kind == string(target.Kind) && name == target.Route
}

func targetSelectorMatches(value map[string]interface{}, route *unstructured.Unstructured, target v1alpha1.EnvoyGatewayTarget) (bool, error) {
	group, _, err := unstructured.NestedString(value, "group")
	if err != nil {
		return false, err
	}
	kind, _, err := unstructured.NestedString(value, "kind")
	if err != nil {
		return false, err
	}
	if (group != "" && group != gatewayAPIGroup) || kind != string(target.Kind) {
		return false, nil
	}

	labelSelector := metav1.LabelSelector{}
	matchLabels, found, err := unstructured.NestedStringMap(value, "matchLabels")
	if err != nil {
		return false, err
	}
	if found {
		labelSelector.MatchLabels = matchLabels
	}
	matchExpressions, found, err := unstructured.NestedSlice(value, "matchExpressions")
	if err != nil {
		return false, err
	}
	if found {
		for index, expressionValue := range matchExpressions {
			expression, ok := expressionValue.(map[string]interface{})
			if !ok {
				return false, errors.Errorf("matchExpressions[%d] is not an object", index)
			}
			key, _, err := unstructured.NestedString(expression, "key")
			if err != nil {
				return false, err
			}
			operator, _, err := unstructured.NestedString(expression, "operator")
			if err != nil {
				return false, err
			}
			values, _, err := unstructured.NestedStringSlice(expression, "values")
			if err != nil {
				return false, err
			}
			labelSelector.MatchExpressions = append(labelSelector.MatchExpressions, metav1.LabelSelectorRequirement{
				Key:      key,
				Operator: metav1.LabelSelectorOperator(operator),
				Values:   values,
			})
		}
	}

	selector, err := metav1.LabelSelectorAsSelector(&labelSelector)
	if err != nil {
		return false, err
	}
	return selector.Matches(labels.Set(route.GetLabels())), nil
}

func buildPolicySpec(chaos *v1alpha1.EnvoyGatewayChaos) map[string]interface{} {
	return map[string]interface{}{
		"targetRefs": []interface{}{
			map[string]interface{}{
				"group": gatewayAPIGroup,
				"kind":  string(chaos.Spec.Target.Kind),
				"name":  chaos.Spec.Target.Route,
			},
		},
		"faultInjection": buildFault(chaos.Spec.Fault),
	}
}

func buildFault(fault v1alpha1.EnvoyGatewayFault) map[string]interface{} {
	result := make(map[string]interface{})
	if fault.Delay != nil {
		result["delay"] = map[string]interface{}{
			"fixedDelay": fault.Delay.FixedDelay,
			"percentage": int64(fault.Delay.Percentage),
		}
	}
	if fault.Abort != nil {
		abort := map[string]interface{}{
			"percentage": int64(fault.Abort.Percentage),
		}
		if fault.Abort.HTTPStatus != nil {
			abort["httpStatus"] = int64(*fault.Abort.HTTPStatus)
		}
		if fault.Abort.GRPCStatus != nil {
			abort["grpcStatus"] = int64(*fault.Abort.GRPCStatus)
		}
		result["abort"] = abort
	}
	return result
}

func newBackendTrafficPolicy() *unstructured.Unstructured {
	policy := &unstructured.Unstructured{}
	policy.SetGroupVersionKind(backendTrafficPolicyGVK)
	return policy
}

func newBackendTrafficPolicyList() *unstructured.UnstructuredList {
	policies := &unstructured.UnstructuredList{}
	policies.SetGroupVersionKind(backendTrafficPolicyListGVK)
	return policies
}

func ownerIdentity(chaos *v1alpha1.EnvoyGatewayChaos) string {
	return fmt.Sprintf("%s/%s:%s", chaos.Namespace, chaos.Name, chaos.UID)
}

func managedPolicyName(target v1alpha1.EnvoyGatewayTarget) string {
	sum := sha256.Sum256([]byte(target.Id()))
	return fmt.Sprintf("%s%x", policyNamePrefix, sum[:6])
}

// NewImpl registers the EnvoyGatewayChaos implementation.
func NewImpl(c client.Client, log logr.Logger) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:       "envoygatewaychaos",
		Object:     &v1alpha1.EnvoyGatewayChaos{},
		ObjectList: &v1alpha1.EnvoyGatewayChaosList{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("envoygatewaychaos"),
		},
	}
}

// Module provides the EnvoyGatewayChaos implementation.
var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
