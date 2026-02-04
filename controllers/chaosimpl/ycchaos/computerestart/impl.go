// Copyright 2021 Chaos Mesh Authors.
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

package computerestart

import (
	"context"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/ycchaos/common"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client

	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	ycchaos, selected, err := common.ParseYCChaosAndSelector(obj, records, index, impl.Log)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	sdk, err := common.GetYandexCloudSDK(ctx, impl.Client, ycchaos)
	if err != nil {
		impl.Log.Error(err, "fail to get Yandex Cloud SDK")
		return v1alpha1.NotInjected, err
	}

	op, err := sdk.Compute().Instance().Restart(ctx, &compute.RestartInstanceRequest{
		InstanceId: selected.ComputeInstance,
	})
	if err != nil {
		impl.Log.Error(err, "fail to restart the compute instance")
		return v1alpha1.NotInjected, err
	}

	err = common.WaitForOperation(ctx, sdk, op, impl.Log, "restart")
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	common.LogOperationSuccess(impl.Log, "restart", selected.ComputeInstance)
	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	// For restart action, recovery is not needed as it's a one-time action
	impl.Log.Info("compute restart is a one-time action, no recovery needed")
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log:    log.WithName("computerestart"),
	}
}
