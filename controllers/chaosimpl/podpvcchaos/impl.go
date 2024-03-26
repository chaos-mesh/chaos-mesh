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

package podpvcchaos

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/podpvc"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("pod pvc chaos Apply", "namespace", obj.GetNamespace(), "name", obj.GetName())

	chaos := obj.(*v1alpha1.PodPVCChaos)

	var target podpvc.PodPVCTarget
	if err := json.Unmarshal([]byte(records[index].Id), &target); err != nil {
		return v1alpha1.NotInjected, err
	}

	var pvc v1.PersistentVolumeClaim
	err := impl.Get(ctx, target.PVC, &pvc)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return v1alpha1.Injected, nil
		}
		return v1alpha1.NotInjected, err
	}

	err = impl.Delete(ctx, &pvc, &client.DeleteOptions{})
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	var pod v1.Pod
	err = impl.Get(ctx, target.Pod, &pod)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return v1alpha1.Injected, nil
		}
		return v1alpha1.NotInjected, err
	}

	err = impl.Delete(ctx, &pod, &client.DeleteOptions{
		GracePeriodSeconds: &chaos.Spec.GracePeriod, // PeriodSeconds has to be set specifically
	})
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "podpvcchaos",
		Object: &v1alpha1.PodPVCChaos{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("podpvchaos"),
		},
		ObjectList: &v1alpha1.PodPVCChaosList{},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
