// Copyright 2021 Chaos Mesh Authors.
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

package diskloss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	compute "google.golang.org/api/compute/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/gcpchaos/utils"
)

type Impl struct {
	client.Client

	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, chaos v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		impl.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return v1alpha1.NotInjected, err
	}
	computeService, err := utils.GetComputeService(ctx, impl.Client, gcpchaos)
	if err != nil {
		impl.Log.Error(err, "fail to get the compute service")
		return v1alpha1.NotInjected, err
	}
	var selected v1alpha1.GcpSelector
	json.Unmarshal([]byte(records[index].Id), &selected)
	instance, err := computeService.Instances.Get(selected.Project, selected.Zone, selected.Instance).Do()
	if err != nil {
		impl.Log.Error(err, "fail to get the instance")
		return v1alpha1.NotInjected, err
	}
	var (
		bytes      []byte
		notFound   []string
		marshalErr []string
	)
	for _, specDeviceName := range *selected.DeviceNames {
		haveDisk := false
		for _, disk := range instance.Disks {
			if disk.DeviceName == specDeviceName {
				haveDisk = true
				bytes, err = json.Marshal(disk)
				if err != nil {
					marshalErr = append(marshalErr, err.Error())
				}
				gcpchaos.Status.AttachedDisksStrings = append(gcpchaos.Status.AttachedDisksStrings, string(bytes))
				break
			}
		}
		if !haveDisk {
			notFound = append(notFound, specDeviceName)
		}
	}
	if len(notFound) != 0 {
		err = fmt.Errorf("instance (%s) does not have the disk (%s)", selected.Instance, notFound)
		impl.Log.Error(err, "the instance does not have the disk")
		return v1alpha1.NotInjected, err
	}
	if len(marshalErr) != 0 {
		err = fmt.Errorf("instance (%s), marshal disk info error (%s)", selected.Instance, marshalErr)
		impl.Log.Error(err, "marshal disk info error")
		return v1alpha1.NotInjected, err
	}

	for _, specDeviceName := range *selected.DeviceNames {
		_, err = computeService.Instances.DetachDisk(selected.Project, selected.Zone, selected.Instance, specDeviceName).Do()
		if err != nil {
			impl.Log.Error(err, "fail to detach the disk")
			return v1alpha1.NotInjected, err
		}
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, chaos v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		impl.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return v1alpha1.Injected, err
	}
	computeService, err := utils.GetComputeService(ctx, impl.Client, gcpchaos)
	if err != nil {
		impl.Log.Error(err, "fail to get the compute service")
		return v1alpha1.Injected, err
	}
	var disk compute.AttachedDisk
	var selected v1alpha1.GcpSelector
	json.Unmarshal([]byte(records[index].Id), &selected)
	for _, attachedDiskString := range gcpchaos.Status.AttachedDisksStrings {
		err = json.Unmarshal([]byte(attachedDiskString), &disk)
		if err != nil {
			impl.Log.Error(err, "fail to unmarshal the disk info")
			return v1alpha1.Injected, err
		}
		_, err = computeService.Instances.AttachDisk(selected.Project, selected.Zone, selected.Instance, &disk).Do()
		if err != nil {
			impl.Log.Error(err, "fail to attach the disk to the instance")
			return v1alpha1.Injected, err
		}
	}
	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log:    log.WithName("diskloss"),
	}
}
