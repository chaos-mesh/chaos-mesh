// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"reflect"
	"strconv"

	lru "github.com/hashicorp/golang-lru"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var _ client.Client = &UpdatedClient{}

type UpdatedClient struct {
	client client.Client
	scheme *runtime.Scheme

	cache *lru.Cache
}

func (c *UpdatedClient) objectKey(key client.ObjectKey, obj runtime.Object) (string, error) {
	gvk, err := apiutil.GVKForObject(obj, c.scheme)
	if err != nil {
		return "", err
	}

	return gvk.String() + "/" + key.String(), nil
}

func (c *UpdatedClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	err := c.client.Get(ctx, key, obj)
	if err != nil {
		return err
	}

	objectKey, err := c.objectKey(key, obj)
	if err != nil {
		return err
	}

	cachedObject, ok := c.cache.Get(objectKey)
	if ok {
		cachedMeta, err := meta.Accessor(cachedObject)
		if err != nil {
			return nil
		}

		objMeta, err := meta.Accessor(obj)
		if err != nil {
			return nil
		}

		cachedResourceVersion, err := strconv.Atoi(cachedMeta.GetResourceVersion())
		if err != nil {
			return nil
		}
		newResourceVersion, err := strconv.Atoi(objMeta.GetResourceVersion())
		if err != nil {
			return nil
		}
		if cachedResourceVersion >= newResourceVersion {
			reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(cachedObject).Elem())
		}
	}

	return nil
}

func (c *UpdatedClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return c.client.List(ctx, list, opts...)
}

func (c *UpdatedClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	return c.client.Create(ctx, obj, opts...)
}

func (c *UpdatedClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	return c.client.Delete(ctx, obj, opts...)
}

func (c *UpdatedClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	err := c.client.Update(ctx, obj, opts...)
	if err != nil {
		return err
	}

	objMeta, err := meta.Accessor(obj)
	if err != nil {
		return nil
	}

	objectKey, err := c.objectKey(types.NamespacedName{
		Namespace: objMeta.GetNamespace(),
		Name:      objMeta.GetName(),
	}, obj)
	if err != nil {
		return err
	}

	c.cache.Add(objectKey, obj.DeepCopyObject())
	return nil
}

func (c *UpdatedClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *UpdatedClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c *UpdatedClient) Status() client.StatusWriter {
	// TODO: add cache for status client

	return c.client.Status()
}
