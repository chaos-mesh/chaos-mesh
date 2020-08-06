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

package podnetworkmanager

import (
	"context"

	"github.com/go-logr/logr"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// PodNetworkManager will save all the related podnetworkchaos
type PodNetworkManager struct {
	Source string
	Log    logr.Logger
	client.Client

	Modifications map[types.NamespacedName]*PodNetworkTransaction
}

// New creates a new PodNetworkMap
func New(source string, logger logr.Logger, client client.Client) *PodNetworkManager {
	return &PodNetworkManager{
		Source:        source,
		Log:           logger,
		Client:        client,
		Modifications: make(map[types.NamespacedName]*PodNetworkTransaction),
	}
}

// WithInit will get a transaction or start a transaction with initially clear
func (m *PodNetworkManager) WithInit(key types.NamespacedName) *PodNetworkTransaction {
	t, ok := m.Modifications[key]
	if ok {
		return t
	}

	t = &PodNetworkTransaction{}
	t.Clear(m.Source)
	m.Modifications[key] = t
	return t
}

// Commit will update all modifications to the cluster
func (m *PodNetworkManager) Commit(ctx context.Context) error {

	// TODO: parallel update
	for key, t := range m.Modifications {
		m.Log.Info("running modification on pod", "key", key, "modification", t)
		updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			chaos := &v1alpha1.PodNetworkChaos{}

			err := m.Client.Get(ctx, key, chaos)
			if err != nil {
				if !k8sError.IsNotFound(err) {
					m.Log.Error(err, "error while getting podnetworkchaos")
					return err
				}

				chaos.Name = key.Name
				chaos.Namespace = key.Namespace
				err = m.Client.Create(ctx, chaos)

				if err != nil {
					m.Log.Error(err, "error while creating podnetworkchaos")
					return err
				}
			}

			err = t.Apply(chaos)
			if err != nil {
				m.Log.Error(err, "error while applying transactions", "transaction", t)
				return err
			}

			return m.Client.Update(ctx, chaos)
		})
		if updateError != nil {
			m.Log.Error(updateError, "error while updating")
			return updateError
		}
	}

	return nil
}
