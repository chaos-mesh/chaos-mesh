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
//

package vmstop

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/cloudstackvm/utils"
)

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	cloudstackchaos := obj.(*v1alpha1.CloudStackVMChaos)
	spec := cloudstackchaos.Spec

	client, err := utils.GetCloudStackClient(ctx, impl.Client, cloudstackchaos)
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("creating cloudstack api client: %w", err)
	}

	var selector v1alpha1.CloudStackVMChaosSelector
	if err := json.Unmarshal([]byte(records[index].Id), &selector); err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("decoding selector: %w", err)
	}

	params := utils.SelectorToListParams(&selector)
	resp, err := client.VirtualMachine.ListVirtualMachines(params)
	if err != nil {
		return v1alpha1.NotInjected, fmt.Errorf("listing VMs: %w", err)
	}

	if len(resp.VirtualMachines) == 0 {
		return v1alpha1.NotInjected, fmt.Errorf("no VMs returned matching criteria")
	}

	for _, vm := range resp.VirtualMachines {
		impl.Log.Info("Stopping VM", "id", vm.Id, "name", vm.Name, "dry-run", spec.DryRun)

		if !spec.DryRun {
			_, err := client.VirtualMachine.StopVirtualMachine(client.VirtualMachine.NewStopVirtualMachineParams(vm.Id))
			if err != nil {
				return v1alpha1.NotInjected, fmt.Errorf("stopping vm %s: %w", vm.Name, err)
			}
		}

		impl.Log.Info("Stopped VM", "id", vm.Id, "name", vm.Name, "dry-run", spec.DryRun)
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	cloudstackchaos := obj.(*v1alpha1.CloudStackVMChaos)
	spec := cloudstackchaos.Spec

	client, err := utils.GetCloudStackClient(ctx, impl.Client, cloudstackchaos)
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("creating cloudstack api client: %w", err)
	}

	var selector v1alpha1.CloudStackVMChaosSelector
	if err := json.Unmarshal([]byte(records[index].Id), &selector); err != nil {
		return v1alpha1.Injected, fmt.Errorf("decoding selector: %w", err)
	}

	params := utils.SelectorToListParams(&selector)
	resp, err := client.VirtualMachine.ListVirtualMachines(params)
	if err != nil {
		return v1alpha1.Injected, fmt.Errorf("listing VMs: %w", err)
	}

	if len(resp.VirtualMachines) == 0 {
		return v1alpha1.Injected, fmt.Errorf("no VMs returned matching criteria")
	}

	for _, vm := range resp.VirtualMachines {
		impl.Log.Info("Starting VM", "id", vm.Id, "name", vm.Name, "dry-run", spec.DryRun)

		if !spec.DryRun {
			_, err := client.VirtualMachine.StartVirtualMachine(client.VirtualMachine.NewStartVirtualMachineParams(vm.Id))
			if err != nil {
				return v1alpha1.Injected, fmt.Errorf("stopping vm %s: %w", vm.Name, err)
			}
		}

		impl.Log.Info("Started VM", "id", vm.Id, "name", vm.Name, "dry-run", spec.DryRun)
	}

	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log:    log.WithName("vmstop"),
	}
}
