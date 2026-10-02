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

package collector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

func TestCreateOrUpdateExperimentPersistsFinalObject(t *testing.T) {
	finishTime := metav1.NewTime(time.Now())
	stored := &core.Experiment{
		ExperimentMeta: core.ExperimentMeta{
			UID:        "experiment-uid",
			FinishTime: &finishTime.Time,
		},
		Experiment: `{"status":{"experiment":{"desiredPhase":"run"}}}`,
	}
	staleExperiment := stored.Experiment
	store := &experimentStoreStub{found: stored}
	collector := &ChaosCollector{store: store}

	chaos := &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pod-chaos",
			Namespace:         "default",
			UID:               "experiment-uid",
			CreationTimestamp: metav1.Now(),
			DeletionTimestamp: &finishTime,
		},
		Status: v1alpha1.PodChaosStatus{
			ChaosStatus: v1alpha1.ChaosStatus{
				Experiment: v1alpha1.ExperimentStatus{DesiredPhase: v1alpha1.StoppedPhase},
			},
		},
	}
	chaos.SetGroupVersionKind(schema.GroupVersionKind{Group: v1alpha1.SchemeGroupVersion.Group, Version: v1alpha1.SchemeGroupVersion.Version, Kind: "PodChaos"})

	require.NoError(t, collector.createOrUpdateExperiment(chaos))
	require.NotNil(t, store.saved)
	require.Equal(t, stored, store.saved)
	require.NotEqual(t, staleExperiment, store.saved.Experiment)
	require.Contains(t, store.saved.Experiment, `"desiredPhase":"Stop"`)
}

type experimentStoreStub struct {
	found *core.Experiment
	saved *core.Experiment
}

func (s *experimentStoreStub) FindByUID(context.Context, string) (*core.Experiment, error) {
	return s.found, nil
}

func (s *experimentStoreStub) Set(_ context.Context, experiment *core.Experiment) error {
	s.saved = experiment
	return nil
}

func (*experimentStoreStub) ListMeta(context.Context, string, string, string, bool) ([]*core.ExperimentMeta, error) {
	return nil, nil
}

func (*experimentStoreStub) FindManagedByNamespaceName(context.Context, string, string) ([]*core.Experiment, error) {
	return nil, nil
}

func (*experimentStoreStub) FindMetaByUID(context.Context, string) (*core.ExperimentMeta, error) {
	return nil, nil
}

func (*experimentStoreStub) Archive(context.Context, string, string) error           { return nil }
func (*experimentStoreStub) Delete(context.Context, *core.Experiment) error          { return nil }
func (*experimentStoreStub) DeleteByFinishTime(context.Context, time.Duration) error { return nil }
func (*experimentStoreStub) DeleteByUIDs(context.Context, []string) error            { return nil }
func (*experimentStoreStub) DeleteIncompleteExperiments(context.Context) error       { return nil }
