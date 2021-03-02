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

package diskloss

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

const (
	GcpFinalizer = "gcp-Finalizer"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		e.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return err
	}
	var computeService *compute.Service
	var err error
	if gcpchaos.Spec.SecretName != nil {
		secret := &v1.Secret{}
		err = e.Client.Get(ctx, types.NamespacedName{
			Name:      *gcpchaos.Spec.SecretName,
			Namespace: gcpchaos.Namespace,
		}, secret)
		if err != nil {
			e.Log.Error(err, "fail to get cloud secret")
			return err
		}

		decodeBytes, err := base64.StdEncoding.DecodeString(string(secret.Data["service_account"]))
		if err != nil {
			e.Log.Error(err, "fail to decode service_account")
			return err
		}
		computeService, err = compute.NewService(ctx, option.WithCredentialsJSON(decodeBytes))
		if err != nil {
			e.Log.Error(err, "fail to create the google compute service")
			return err
		}
	} else {
		computeService, err = compute.NewService(ctx)
		if err != nil {
			e.Log.Error(err, "fail to create the google compute service")
			return err
		}
	}

	haveDisk := false
	instance, err := computeService.Instances.Get(gcpchaos.Spec.Project, gcpchaos.Spec.Zone, gcpchaos.Spec.Instance).Do()
	if err != nil {
		e.Log.Error(err, "fail to get the instance")
		return err
	}
	var bytes []byte
	for _, disk := range instance.Disks {
		if disk.DeviceName == *gcpchaos.Spec.DeviceName {
			haveDisk = true
			bytes, err = json.Marshal(disk)
			if err != nil {
				e.Log.Error(err, "fail to marshal the disk info")
				return err
			}
			break
		}
	}
	if haveDisk == false {
		err = fmt.Errorf("instance (%s) does not have the disk (%s)", gcpchaos.Spec.Instance, *gcpchaos.Spec.DeviceName)
		e.Log.Error(err, "the instance does not have the disk")
		return err
	}

	gcpchaos.Finalizers = []string{GcpFinalizer}
	_, err = computeService.Instances.DetachDisk(gcpchaos.Spec.Project, gcpchaos.Spec.Zone, gcpchaos.Spec.Instance, *gcpchaos.Spec.DeviceName).Do()
	gcpchaos.Spec.AttachedDiskString = string(bytes)
	if err != nil {
		gcpchaos.Finalizers = make([]string, 0)
		e.Log.Error(err, "fail to detach the disk")
		return err
	}

	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	gcpchaos, ok := chaos.(*v1alpha1.GcpChaos)
	if !ok {
		err := errors.New("chaos is not gcpchaos")
		e.Log.Error(err, "chaos is not GcpChaos", "chaos", chaos)
		return err
	}
	gcpchaos.Finalizers = make([]string, 0)
	var computeService *compute.Service
	var err error
	if gcpchaos.Spec.SecretName != nil {
		secret := &v1.Secret{}
		err = e.Client.Get(ctx, types.NamespacedName{
			Name:      *gcpchaos.Spec.SecretName,
			Namespace: gcpchaos.Namespace,
		}, secret)
		if err != nil {
			e.Log.Error(err, "fail to get cloud secret")
			return err
		}

		decodeBytes, err := base64.StdEncoding.DecodeString(string(secret.Data["service_account"]))
		if err != nil {
			e.Log.Error(err, "fail to decode service_account")
			return err
		}
		computeService, err = compute.NewService(ctx, option.WithCredentialsJSON(decodeBytes))
		if err != nil {
			e.Log.Error(err, "fail to create the google compute service")
			return err
		}
	} else {
		computeService, err = compute.NewService(ctx)
		if err != nil {
			e.Log.Error(err, "fail to create the google compute service")
			return err
		}
	}
	var disk compute.AttachedDisk
	err = json.Unmarshal([]byte(gcpchaos.Spec.AttachedDiskString), &disk)
	if err != nil {
		e.Log.Error(err, "fail to unmarshal the disk info")
		return err
	}
	_, err = computeService.Instances.AttachDisk(gcpchaos.Spec.Project, gcpchaos.Spec.Zone, gcpchaos.Spec.Instance, &disk).Do()
	if err != nil {
		e.Log.Error(err, "fail to attach the disk to the instance")
		return err
	}
	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.GcpChaos{}
}

func init() {
	router.Register("gcpchaos", &v1alpha1.GcpChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.GcpChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.DiskLoss
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
