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

package namespaceannoationfilter

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const injectAnnotationKey = "chaos-mesh.org/inject"

func IsAllowedNamespaces(ctx context.Context, c client.Client, namespace string) (bool, error) {
	ns := &v1.Namespace{}

	err := c.Get(ctx, types.NamespacedName{Name: namespace}, ns)
	if err != nil {
		return false, err
	}

	if ns.Annotations[injectAnnotationKey] == "enabled" {
		return true, nil
	}

	return false, nil
}
