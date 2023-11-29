// Copyright 2023 Chaos Mesh Authors.
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

package k8schaos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger
}

const (
	managedByLabel = "app.kubernetes.io/managed-by"
	managedBy      = "chaos-mesh"
)

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("k8schaos Apply", "namespace", obj.GetNamespace(), "name", obj.GetName())

	k8schaos, ok := obj.(*v1alpha1.K8SChaos)
	if !ok {
		err := errors.New("chaos is not K8SChaos")
		return v1alpha1.NotInjected, err
	}

	client, err := impl.dynamicClient()
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("dynamic client new: %w", err)
	}

	resource, err := impl.resourceForIndex(k8schaos.Spec.APIObjects.Value, index)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	labels := resource.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[managedByLabel] = managedBy
	resource.SetLabels(labels)

	gvk := resource.GroupVersionKind()

	mapping, err := impl.Client.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("get rest mapping: %w", err)
	}

	originalValue, err := client.Resource(mapping.Resource).Namespace(resource.GetNamespace()).Get(ctx, resource.GetName(), v1.GetOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return v1alpha1.NotInjected, err
	}

	if k8schaos.Spec.Update && originalValue != nil {
		var resourceVersion string
		var found bool
		resourceVersion, found, err = unstructured.NestedString(originalValue.Object, "metadata", "resourceVersion")
		if err != nil {
			return v1alpha1.NotInjected, fmt.Errorf("resourceVersion is not a string: %w", err)
		}
		if found {
			resource.SetResourceVersion(resourceVersion)
		}

		unstructured.RemoveNestedField(originalValue.Object, "metadata", "resourceVersion")
		unstructured.RemoveNestedField(originalValue.Object, "metadata", "uid")
		serialized, err := json.Marshal(originalValue)
		if err != nil {
			return v1alpha1.NotInjected, fmt.Errorf("failed to serialize original resource value: %w", err)
		}

		k8schaos.Status.OriginalObjectValue = string(serialized)

		impl.Log.Info("k8schaos: updating existing resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "method", "PUT", "resourceVersion", resourceVersion)
		_, err = client.Resource(mapping.Resource).Namespace(resource.GetNamespace()).Update(ctx, resource, v1.UpdateOptions{})

		if err != nil {
			impl.Log.Error(err, "k8schaos: failed to update resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
				"target-namespace", resource.GetNamespace(), "target-name", resource.GetName())
			return v1alpha1.NotInjected, err
		}

		return v1alpha1.Injected, nil
	} else {
		if k8schaos.Spec.Update {
			impl.Log.Info("k8schaos: warning: chaos has update=true but resource not found - creating a new resource instead", "namespace",
				obj.GetNamespace(), "name", obj.GetName(),
				"target-namespace", resource.GetNamespace(), "target-name", resource.GetName())
		}

		impl.Log.Info("k8schaos: creating new resources", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "method", "POST")
		_, err = client.Resource(mapping.Resource).Namespace(resource.GetNamespace()).Create(ctx, resource, v1.CreateOptions{})
	}

	if err != nil {
		impl.Log.Error(err, "k8schaos: failed to create resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName())
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("k8schaos Recover", "namespace", obj.GetNamespace(), "name", obj.GetName())

	// TODO: We need to consider the case where we've modified the original object and should instead restore it from
	// k8schaos.Status.OriginalObjects.
	k8schaos, ok := obj.(*v1alpha1.K8SChaos)
	if !ok {
		err := errors.New("chaos is not K8SChaos")
		return v1alpha1.Injected, err
	}

	client, err := impl.dynamicClient()
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("dynamic client new: %w", err)
	}

	resource, err := impl.resourceForIndex(k8schaos.Spec.APIObjects.Value, index)
	if err != nil {
		return v1alpha1.Injected, err
	}

	gvk := resource.GroupVersionKind()

	mapping, err := impl.Client.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("get rest mapping: %w", err)
	}

	resourceClient := client.Resource(mapping.Resource).Namespace(resource.GetNamespace())
	existingResource, err := resourceClient.Get(ctx, resource.GetName(), v1.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			impl.Log.Error(err, "k8schaos: resource not found", "namespace", obj.GetNamespace(), "name", obj.GetName(),
				"target-namespace", resource.GetNamespace(), "target-name", resource.GetName())
			return v1alpha1.NotInjected, nil
		}
		impl.Log.Error(err, "k8schaos: failed to load resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName())
		return v1alpha1.Injected, err
	}

	if resMgr := getResourceManager(existingResource); resMgr != managedBy {
		impl.Log.Error(nil, "k8schaos: resource not managed by chaos mesh", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "managed-by", resMgr)
		return v1alpha1.Injected, fmt.Errorf("resource is not managed by %s: %s: \"%s\"", managedBy, managedByLabel, resMgr)
	}

	if k8schaos.Status.OriginalObjectValue != "" {
		var recoveryValue unstructured.Unstructured
		if err := json.Unmarshal([]byte(k8schaos.Status.OriginalObjectValue), &recoveryValue); err != nil {
			return v1alpha1.Injected, fmt.Errorf("failed to load value to roll back to from status: %w", err)
		}

		var resourceVersion string
		var found bool
		resourceVersion, found, err = unstructured.NestedString(existingResource.Object, "metadata", "resourceVersion")
		if err != nil {
			return v1alpha1.Injected, fmt.Errorf("resourceVersion is not a string: %w", err)
		}
		if found {
			recoveryValue.SetResourceVersion(resourceVersion)
		}

		impl.Log.Info("k8schaos: rolling back resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "method", "PUT")
		_, err = client.Resource(mapping.Resource).Namespace(resource.GetNamespace()).Update(ctx, &recoveryValue, v1.UpdateOptions{})
		if err != nil {
			impl.Log.Info("k8schaos: failed to roll back resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
				"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "method", "PUT")
			return v1alpha1.Injected, err
		}

		return v1alpha1.NotInjected, nil
	}

	if k8schaos.Spec.Update {
		impl.Log.Info("k8schaos: warning: chaos has update=true but no resource is stored in status - resource will be deleted", "namespace",
			obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName())
	}

	impl.Log.Info("k8schaos: deleting resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
		"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "method", "DELETE")
	err = resourceClient.Delete(ctx, resource.GetName(), v1.DeleteOptions{})

	if err != nil && !apiErrors.IsNotFound(err) {
		impl.Log.Info("k8schaos: failed to delete resource", "namespace", obj.GetNamespace(), "name", obj.GetName(),
			"target-namespace", resource.GetNamespace(), "target-name", resource.GetName(), "method", "DELETE")
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func getResourceManager(resource *unstructured.Unstructured) string {
	labels := resource.GetLabels()
	if labels == nil {
		return ""
	}
	return labels[managedByLabel]
}

func (impl *Impl) dynamicClient() (dynamic.Interface, error) {
	config, err := ctrl.GetConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

func (impl *Impl) resourceForIndex(yamlStr string, index int) (*unstructured.Unstructured, error) {
	decoder := yaml.NewDecoder(strings.NewReader(yamlStr))

	for i := 0; ; i++ {
		unstr := unstructured.Unstructured{}

		err := decoder.Decode(&unstr.Object)
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("no resource for index: %d", index)
			}

			return nil, err
		}

		if i == index {
			return &unstr, nil
		}
	}
}

func NewImpl(c client.Client, log logr.Logger) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "k8schaos",
		Object: &v1alpha1.K8SChaos{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("k8schaos"),
		},
		ObjectList: &v1alpha1.K8SChaosList{},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
