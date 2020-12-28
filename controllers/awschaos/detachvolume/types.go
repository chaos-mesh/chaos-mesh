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

package detachvolume

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

const (
	AwsFinalizer = "aws-Finalizer"
)

type endpoint struct {
	ctx.Context
}

func (e *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	awschaos, ok := chaos.(*v1alpha1.AwsChaos)
	if !ok {
		err := errors.New("chaos is not stresschaos")
		e.Log.Error(err, "chaos is not StressChaos", "chaos", chaos)
		return err
	}

	secret := &v1.Secret{}
	err := e.Client.Get(ctx, types.NamespacedName{
		Name:      awschaos.Spec.SecretName,
		Namespace: awschaos.Namespace,
	}, secret)
	if err != nil {
		e.Log.Error(err, "fail to get cloud secret")
		return err
	}

	ec2client := ec2.New(ec2.Options{
		Region:      awschaos.Spec.AwsRegion,
		Credentials: credentials.NewStaticCredentialsProvider(string(secret.Data["aws_access_key_id"]), string(secret.Data["aws_secret_access_key"]), ""),
	})

	awschaos.Finalizers = []string{AwsFinalizer}
	_, err = ec2client.DetachVolume(context.TODO(), &ec2.DetachVolumeInput{
		VolumeId:   &awschaos.Spec.EbsVolume,
		Device:     &awschaos.Spec.DeviceName,
		Force:      true,
		InstanceId: &awschaos.Spec.Ec2Instance,
	})

	if err != nil {
		e.Log.Error(err, "fail to detach the volume")
		return err
	}

	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	awschaos, ok := chaos.(*v1alpha1.AwsChaos)
	if !ok {
		err := errors.New("chaos is not stresschaos")
		e.Log.Error(err, "chaos is not StressChaos", "chaos", chaos)
		return err
	}

	awschaos.Finalizers = make([]string, 0)
	secret := &v1.Secret{}
	err := e.Client.Get(ctx, types.NamespacedName{
		Name:      awschaos.Spec.SecretName,
		Namespace: awschaos.Namespace,
	}, secret)
	if err != nil {
		e.Log.Error(err, "fail to get cloud secret")
		return err
	}

	ec2client := ec2.New(ec2.Options{
		Region:      awschaos.Spec.AwsRegion,
		Credentials: credentials.NewStaticCredentialsProvider(string(secret.Data["aws_access_key_id"]), string(secret.Data["aws_secret_access_key"]), ""),
	})

	_, err = ec2client.AttachVolume(context.TODO(), &ec2.AttachVolumeInput{
		Device:     &awschaos.Spec.DeviceName,
		InstanceId: &awschaos.Spec.Ec2Instance,
		VolumeId:   &awschaos.Spec.EbsVolume,
	})

	if err != nil {
		e.Log.Error(err, "fail to attach the volume")
		return err
	}

	return nil
}

func (e *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.AwsChaos{}
}

func init() {
	router.Register("awschaos", &v1alpha1.AwsChaos{}, func(obj runtime.Object) bool {
		chaos, ok := obj.(*v1alpha1.AwsChaos)
		if !ok {
			return false
		}

		return chaos.Spec.Action == v1alpha1.DetachVolume
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
