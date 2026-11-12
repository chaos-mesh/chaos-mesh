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

package podnetworkchaosmanager

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// ErrNodeNotFound means the node to operate may have been deleted
var ErrNodeNotFound = errors.New("node not found")

// NodeNetworkManager saves the modifications on the NodeNetworkChaos of a node
type NodeNetworkManager struct {
	Source string

	Log logr.Logger
	client.Client
	client.Reader

	NodeName string
	T        *PodNetworkTransaction
}

// BuildNode creates a NodeNetworkManager which keeps the existing rules of the source
func (b *Builder) BuildNode(source string, nodeName string) *NodeNetworkManager {
	return &NodeNetworkManager{
		Source:   source,
		Log:      b.Log,
		Client:   b.Client,
		Reader:   b.Reader,
		NodeName: nodeName,
		T:        &PodNetworkTransaction{},
	}
}

// WithInitNode creates a NodeNetworkManager which clears the existing rules of the source
func (b *Builder) WithInitNode(source string, nodeName string) *NodeNetworkManager {
	m := b.BuildNode(source, nodeName)
	m.T.Clear(source)
	return m
}

// Commit updates the NodeNetworkChaos of the node and returns its generation
func (m *NodeNetworkManager) Commit(ctx context.Context) (int64, error) {
	m.Log.Info("running modification on node", "node", m.NodeName, "modification", m.T)
	key := types.NamespacedName{Name: m.NodeName}
	updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		chaos := &v1alpha1.NodeNetworkChaos{}

		err := m.Client.Get(ctx, key, chaos)
		if err != nil {
			if !k8sError.IsNotFound(err) {
				m.Log.Error(err, "error while getting nodenetworkchaos")
				return err
			}

			return m.createNodeNetworkChaos(ctx)
		}

		if err := m.apply(chaos); err != nil {
			return err
		}

		return m.Client.Update(ctx, chaos)
	})
	if updateError != nil {
		return 0, updateError
	}

	chaos := &v1alpha1.NodeNetworkChaos{}
	if err := m.Reader.Get(ctx, key, chaos); err != nil {
		m.Log.Error(err, "error while getting the latest generation number")
		return 0, err
	}
	return chaos.GetGeneration(), nil
}

func (m *NodeNetworkManager) createNodeNetworkChaos(ctx context.Context) error {
	node := v1.Node{}
	if err := m.Client.Get(ctx, types.NamespacedName{Name: m.NodeName}, &node); err != nil {
		if k8sError.IsNotFound(err) {
			return ErrNodeNotFound
		}
		return err
	}

	chaos := &v1alpha1.NodeNetworkChaos{}
	chaos.Name = m.NodeName
	chaos.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node.Name,
			UID:        node.UID,
		},
	}
	if err := m.apply(chaos); err != nil {
		return err
	}

	return m.Client.Create(ctx, chaos)
}

func (m *NodeNetworkManager) apply(chaos *v1alpha1.NodeNetworkChaos) error {
	staging := &v1alpha1.PodNetworkChaos{
		Spec: v1alpha1.PodNetworkChaosSpec{
			IPSets:   chaos.Spec.IPSets,
			Iptables: chaos.Spec.Iptables,
		},
	}
	if err := m.T.Apply(staging); err != nil {
		m.Log.Error(err, "error while applying transactions", "transaction", m.T)
		return err
	}
	if len(staging.Spec.TrafficControls) > 0 {
		return errors.New("traffic control is not supported on nodenetworkchaos")
	}

	chaos.Spec.IPSets = staging.Spec.IPSets
	chaos.Spec.Iptables = staging.Spec.Iptables
	return nil
}
