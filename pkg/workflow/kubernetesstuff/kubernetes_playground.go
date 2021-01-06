// Copyright 2020 Chaos Mesh Authors.
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

package kubernetesstuff

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type KubernetesPlayground struct {
	client client.Client
}

func (it *KubernetesPlayground) CreateNetworkChaos(networkChaos v1alpha1.NetworkChaos) error {
	// TODO: make context work
	return it.client.Create(context.TODO(), &networkChaos)
}

func (it *KubernetesPlayground) DeleteNetworkChaos(namespace, name string) error {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	target := v1alpha1.NetworkChaos{}
	// TODO: make context work
	err := it.client.Get(context.TODO(), key, &target)
	if err != nil {
		return err
	}
	// TODO: make context work
	return it.client.Delete(context.TODO(), &target)
}
