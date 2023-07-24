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

	// TODO: We need to consider the case where we're applying an object that already exists (updating it).
	// In that case we should store the original objects in k8schaos.Status.OriginalObjects.
	k8schaos, ok := obj.(*v1alpha1.K8SChaos)
	if !ok {
		err := errors.New("chaos is not K8SChaos")
		impl.Log.Error(err, "chaos is not K8SChaos", "chaos", obj)
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

	_, err = client.Resource(mapping.Resource).Namespace(resource.GetNamespace()).Create(ctx, resource, v1.CreateOptions{})
	if err != nil {
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
		impl.Log.Error(err, "chaos is not K8SChaos", "chaos", obj)
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
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	if !isManaged(existingResource) {
		return v1alpha1.Injected, fmt.Errorf("resource is not managed by %s", managedBy)
	}

	err = resourceClient.Delete(ctx, resource.GetName(), v1.DeleteOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func isManaged(resource *unstructured.Unstructured) bool {
	labels := resource.GetLabels()
	return labels != nil && labels[managedByLabel] == managedBy
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
	NewImpl,
)
