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
	"crypto/sha256"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

const (
	ownerAnnotation = "chaos-mesh.org/istio-chaos-owner"
	routeNamePrefix = "chaos-mesh-istio-"
)

var virtualServiceGVK = schema.GroupVersionKind{
	Group:   "networking.istio.io",
	Version: "v1beta1",
	Kind:    "VirtualService",
}

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, _ int, _ []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.IstioChaos)
	virtualService := newVirtualService()
	key := types.NamespacedName{
		Namespace: chaos.Spec.Target.Namespace,
		Name:      chaos.Spec.Target.VirtualService,
	}
	if err := impl.Client.Get(ctx, key, virtualService); err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "get target VirtualService")
	}

	owner := ownerIdentity(chaos)
	annotations := virtualService.GetAnnotations()
	if currentOwner := annotations[ownerAnnotation]; currentOwner != "" && currentOwner != owner {
		return v1alpha1.NotInjected, errors.Errorf("VirtualService is already controlled by IstioChaos %s", currentOwner)
	}

	routes, found, err := unstructured.NestedSlice(virtualService.Object, "spec", "http")
	if err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "read spec.http from target VirtualService")
	}
	if !found {
		return v1alpha1.NotInjected, errors.New("target VirtualService has no spec.http routes")
	}

	managedRouteName := managedRouteName(chaos)
	targetIndex := -1
	for index, routeValue := range routes {
		route, ok := routeValue.(map[string]interface{})
		if !ok {
			return v1alpha1.NotInjected, errors.Errorf("spec.http[%d] is not an object", index)
		}
		name, _, err := unstructured.NestedString(route, "name")
		if err != nil {
			return v1alpha1.NotInjected, errors.Wrapf(err, "read spec.http[%d].name", index)
		}
		if name == managedRouteName {
			if annotations[ownerAnnotation] != owner {
				return v1alpha1.NotInjected, errors.Errorf("managed route %q exists without matching ownership", managedRouteName)
			}
			return v1alpha1.Injected, nil
		}
		if chaos.Spec.Target.HTTPRoute != "" && name == chaos.Spec.Target.HTTPRoute {
			if targetIndex != -1 {
				return v1alpha1.NotInjected, errors.Errorf("multiple HTTP routes named %q", chaos.Spec.Target.HTTPRoute)
			}
			targetIndex = index
		}
	}
	if chaos.Spec.Target.HTTPRoute == "" {
		if len(routes) != 1 {
			return v1alpha1.NotInjected, errors.Errorf(
				"target VirtualService has %d HTTP routes; httpRoute must be set unless it has exactly one",
				len(routes),
			)
		}
		targetIndex = 0
	} else if targetIndex == -1 {
		return v1alpha1.NotInjected, errors.Errorf("HTTP route %q not found", chaos.Spec.Target.HTTPRoute)
	}

	managedRoute := runtime.DeepCopyJSONValue(routes[targetIndex]).(map[string]interface{})
	managedRoute["name"] = managedRouteName
	managedRoute["fault"] = buildFault(chaos.Spec.Fault)
	routes = append(routes, nil)
	copy(routes[targetIndex+1:], routes[targetIndex:])
	routes[targetIndex] = managedRoute
	if err := unstructured.SetNestedSlice(virtualService.Object, routes, "spec", "http"); err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "write spec.http to target VirtualService")
	}
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[ownerAnnotation] = owner
	virtualService.SetAnnotations(annotations)

	if err := impl.Client.Update(ctx, virtualService); err != nil {
		return v1alpha1.NotInjected, errors.Wrap(err, "inject fault into target VirtualService")
	}

	impl.Log.Info("injected Istio fault", "virtualService", key, "route", chaos.Spec.Target.HTTPRoute)
	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, _ int, _ []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.IstioChaos)
	virtualService := newVirtualService()
	key := types.NamespacedName{
		Namespace: chaos.Spec.Target.Namespace,
		Name:      chaos.Spec.Target.VirtualService,
	}
	if err := impl.Client.Get(ctx, key, virtualService); err != nil {
		if apierrors.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, errors.Wrap(err, "get target VirtualService")
	}

	owner := ownerIdentity(chaos)
	annotations := virtualService.GetAnnotations()
	if currentOwner := annotations[ownerAnnotation]; currentOwner != "" && currentOwner != owner {
		return v1alpha1.Injected, errors.Errorf("VirtualService ownership changed to IstioChaos %s", currentOwner)
	}

	routes, found, err := unstructured.NestedSlice(virtualService.Object, "spec", "http")
	if err != nil {
		return v1alpha1.Injected, errors.Wrap(err, "read spec.http from target VirtualService")
	}
	if !found {
		routes = []interface{}{}
	}

	managedName := managedRouteName(chaos)
	filtered := make([]interface{}, 0, len(routes))
	removed := false
	for index, routeValue := range routes {
		route, ok := routeValue.(map[string]interface{})
		if !ok {
			return v1alpha1.Injected, errors.Errorf("spec.http[%d] is not an object", index)
		}
		name, _, err := unstructured.NestedString(route, "name")
		if err != nil {
			return v1alpha1.Injected, errors.Wrapf(err, "read spec.http[%d].name", index)
		}
		if name == managedName {
			removed = true
			continue
		}
		filtered = append(filtered, routeValue)
	}

	ownsAnnotation := annotations[ownerAnnotation] == owner
	if !removed && !ownsAnnotation {
		return v1alpha1.NotInjected, nil
	}
	if removed {
		if err := unstructured.SetNestedSlice(virtualService.Object, filtered, "spec", "http"); err != nil {
			return v1alpha1.Injected, errors.Wrap(err, "write spec.http to target VirtualService")
		}
	}
	if ownsAnnotation {
		delete(annotations, ownerAnnotation)
		virtualService.SetAnnotations(annotations)
	}
	if err := impl.Client.Update(ctx, virtualService); err != nil {
		return v1alpha1.Injected, errors.Wrap(err, "recover target VirtualService")
	}

	impl.Log.Info("recovered Istio fault", "virtualService", key, "route", chaos.Spec.Target.HTTPRoute)
	return v1alpha1.NotInjected, nil
}

func buildFault(fault v1alpha1.IstioFault) map[string]interface{} {
	result := make(map[string]interface{})
	if fault.Delay != nil {
		result["delay"] = map[string]interface{}{
			"fixedDelay": fault.Delay.FixedDelay,
			"percentage": map[string]interface{}{"value": int64(fault.Delay.Percentage)},
		}
	}
	if fault.Abort != nil {
		result["abort"] = map[string]interface{}{
			"httpStatus": int64(fault.Abort.HTTPStatus),
			"percentage": map[string]interface{}{"value": int64(fault.Abort.Percentage)},
		}
	}
	return result
}

func newVirtualService() *unstructured.Unstructured {
	virtualService := &unstructured.Unstructured{}
	virtualService.SetGroupVersionKind(virtualServiceGVK)
	return virtualService
}

func namespacedName(chaos *v1alpha1.IstioChaos) string {
	return types.NamespacedName{Namespace: chaos.Namespace, Name: chaos.Name}.String()
}

func ownerIdentity(chaos *v1alpha1.IstioChaos) string {
	return fmt.Sprintf("%s:%s", namespacedName(chaos), chaos.UID)
}

func managedRouteName(chaos *v1alpha1.IstioChaos) string {
	identity := string(chaos.UID)
	if identity == "" {
		identity = namespacedName(chaos)
	}
	sum := sha256.Sum256([]byte(identity))
	return fmt.Sprintf("%s%x", routeNamePrefix, sum[:6])
}

func NewImpl(c client.Client, log logr.Logger) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:       "istiochaos",
		Object:     &v1alpha1.IstioChaos{},
		ObjectList: &v1alpha1.IstioChaosList{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("istiochaos"),
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
