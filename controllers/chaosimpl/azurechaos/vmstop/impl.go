// Copyright 2022 Chaos Mesh Authors.
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

package vmstop

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/azurechaos/utils"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client

	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, chaos v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	azurechaos, ok := chaos.(*v1alpha1.AzureChaos)
	if !ok {
		err := errors.New("chaos is not azurechaos")
		impl.Log.Error(err, "chaos is not azureChaos", "chaos", chaos)
		return v1alpha1.NotInjected, err
	}

	vmClient, err := utils.GetVMClient(ctx, impl.Client, azurechaos)
	if err != nil {
		impl.Log.Error(err, "fail to get the vm client")
		return v1alpha1.NotInjected, err
	}

	var selected v1alpha1.AzureSelector
	err = json.Unmarshal([]byte(records[index].Id), &selected)
	if err != nil {
		impl.Log.Error(err, "selector unmarshal error")
		return v1alpha1.NotInjected, err
	}

	_, err = vmClient.PowerOff(ctx, selected.ResourceGroupName, selected.VMName, nil)
	if err != nil {
		impl.Log.Error(err, "fail to power off the vm")
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, chaos v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	azurechaos, ok := chaos.(*v1alpha1.AzureChaos)
	if !ok {
		err := errors.New("chaos is not azurechaos")
		impl.Log.Error(err, "chaos is not AzureChaos", "chaos", chaos)
		return v1alpha1.Injected, err
	}
	vmClient, err := utils.GetVMClient(ctx, impl.Client, azurechaos)
	if err != nil {
		impl.Log.Error(err, "fail to get the vm client")
		return v1alpha1.Injected, err
	}

	var selected v1alpha1.AzureSelector
	err = json.Unmarshal([]byte(records[index].Id), &selected)
	if err != nil {
		impl.Log.Error(err, "fail to unmarshal the selector")
		return v1alpha1.NotInjected, err
	}

	_, err = vmClient.Start(ctx, selected.ResourceGroupName, selected.VMName)
	if err != nil {
		impl.Log.Error(err, "fail to start the vm")
		return v1alpha1.Injected, err
	}
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log:    log.WithName("vmstop"),
	}
}
