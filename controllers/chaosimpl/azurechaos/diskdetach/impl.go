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

package diskdetach

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
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
		impl.Log.Error(err, "fail to unmarshal the selector")
		return v1alpha1.NotInjected, err
	}

	vm, err := vmClient.Get(ctx, azurechaos.Spec.ResourceGroupName, azurechaos.Spec.VMName, compute.InstanceView)
	if err != nil {
		impl.Log.Error(err, "fail to get the specified vm")
		return v1alpha1.NotInjected, err
	}

	var disks []compute.DataDisk
	if vm.StorageProfile != nil && vm.StorageProfile.DataDisks != nil {
		disks = *vm.StorageProfile.DataDisks
	}
	diskIndex := -1
	for index, disk := range disks {
		if disk.Lun != nil && disk.Name != nil && *disk.Lun == int32(*azurechaos.Spec.LUN) && *disk.Name == *azurechaos.Spec.DiskName {
			diskIndex = index
			break
		}
	}
	if diskIndex == -1 {
		err := fmt.Errorf("disk %q at LUN %d is not attached to VM %q", *azurechaos.Spec.DiskName, *azurechaos.Spec.LUN, azurechaos.Spec.VMName)
		impl.Log.Error(err, "fail to get the disk")
		return v1alpha1.NotInjected, err
	}

	newDiskList := append(disks[:diskIndex], disks[diskIndex+1:]...)
	vm.StorageProfile.DataDisks = &newDiskList

	_, err = vmClient.CreateOrUpdate(ctx, azurechaos.Spec.ResourceGroupName, azurechaos.Spec.VMName, vm)
	if err != nil {
		impl.Log.Error(err, "fail to detach the disk from the vm")
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
	disksClient, err := utils.GetDiskClient(ctx, impl.Client, azurechaos)
	if err != nil {
		impl.Log.Error(err, "fail to get the disk client")
		return v1alpha1.Injected, err
	}

	var selected v1alpha1.AzureSelector
	err = json.Unmarshal([]byte(records[index].Id), &selected)
	if err != nil {
		impl.Log.Error(err, "fail to unmarshal the selector")
		return v1alpha1.Injected, err
	}

	disk, err := disksClient.Get(ctx, azurechaos.Spec.ResourceGroupName, *azurechaos.Spec.DiskName)
	if err != nil {
		impl.Log.Error(err, "fail to get the disk detail")
		return v1alpha1.Injected, err
	}
	vm, err := vmClient.Get(ctx, azurechaos.Spec.ResourceGroupName, azurechaos.Spec.VMName, compute.InstanceView)
	if err != nil {
		impl.Log.Error(err, "fail to get the specified vm")
		return v1alpha1.Injected, err
	}

	if vm.StorageProfile == nil {
		return v1alpha1.Injected, errors.New("VM has no storage profile")
	}
	var newDiskList []compute.DataDisk
	if vm.StorageProfile.DataDisks != nil {
		newDiskList = *vm.StorageProfile.DataDisks
	}
	newDiskList = append(newDiskList, compute.DataDisk{
		Lun:          new(int32(*azurechaos.Spec.LUN)),
		Name:         azurechaos.Spec.DiskName,
		CreateOption: compute.DiskCreateOptionTypesAttach,
		ManagedDisk: &compute.ManagedDiskParameters{
			ID: disk.ID,
		},
	})
	vm.StorageProfile.DataDisks = &newDiskList

	_, err = vmClient.CreateOrUpdate(ctx, azurechaos.Spec.ResourceGroupName, azurechaos.Spec.VMName, vm)
	if err != nil {
		impl.Log.Error(err, "fail to attach the disk")
		return v1alpha1.Injected, err
	}
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log:    log.WithName("diskdetach"),
	}
}
