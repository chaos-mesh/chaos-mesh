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

package podiochaosmanager

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var (
	// ErrPodNotFound means operate pod may be deleted(almostly)
	ErrPodNotFound = errors.New("pod not found")

	// ErrPodNotRunning means operate pod may be not working
	// and it's non-sense to make changes on it.
	ErrPodNotRunning = errors.New("pod not running")
)

// PodIoManager will save all the related podiochaos
type PodIoManager struct {
	Source string
	Log    logr.Logger
	client.Client

	Modifications map[types.NamespacedName]*PodIoTransaction
}

// New creates a new PodIoMap
func New(source string, logger logr.Logger, client client.Client) *PodIoManager {
	return &PodIoManager{
		Source:        source,
		Log:           logger,
		Client:        client,
		Modifications: make(map[types.NamespacedName]*PodIoTransaction),
	}
}

// WithInit will get a transaction or start a transaction with initially clear
func (m *PodIoManager) WithInit(key types.NamespacedName) *PodIoTransaction {
	t, ok := m.Modifications[key]
	if ok {
		return t
	}

	t = &PodIoTransaction{}
	t.Clear(m.Source)
	m.Modifications[key] = t
	return t
}

// CommitResponse is a tuple (Key, Err)
type CommitResponse struct {
	Key types.NamespacedName
	Err error
}

// Commit will update all modifications to the cluster
func (m *PodIoManager) Commit(ctx context.Context) []CommitResponse {
	g := errgroup.Group{}
	results := make([]CommitResponse, len(m.Modifications))
	index := 0
	for key, t := range m.Modifications {
		i := index

		key := key
		t := t
		g.Go(func() error {
			m.Log.Info("running modification on pod", "key", key, "modification", t)
			updateError := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				chaos := &v1alpha1.PodIoChaos{}

				err := m.Client.Get(ctx, key, chaos)
				if err != nil {
					if !k8sError.IsNotFound(err) {
						m.Log.Error(err, "error while getting podiochaos")
						return err
					}

					pod := v1.Pod{}
					err = m.Client.Get(ctx, key, &pod)
					if err != nil {
						if !k8sError.IsNotFound(err) {
							m.Log.Error(err, "error while finding pod")
							return err
						}

						m.Log.Info("pod not found", "key", key, "error", err.Error())
						err = ErrPodNotFound
						return err
					}

					if pod.Status.Phase != v1.PodRunning {
						m.Log.Info("pod is not running", "key", key)
						err = ErrPodNotRunning
						return err
					}

					chaos.Name = key.Name
					chaos.Namespace = key.Namespace
					chaos.OwnerReferences = []metav1.OwnerReference{
						{
							APIVersion: pod.APIVersion,
							Kind:       pod.Kind,
							Name:       pod.Name,
							UID:        pod.UID,
						},
					}
					err = m.Client.Create(ctx, chaos)

					if err != nil {
						m.Log.Error(err, "error while creating podiochaos")
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

			results[i] = CommitResponse{
				Key: key,
				Err: updateError,
			}

			return nil
		})

		index++
	}

	g.Wait()

	return results
}
