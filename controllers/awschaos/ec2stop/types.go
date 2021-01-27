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

package ec2stop

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
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
		err := errors.New("chaos is not awschaos")
		e.Log.Error(err, "chaos is not AwsChaos", "chaos", chaos)
		return err
	}

	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(awschaos.Spec.AwsRegion),
	}
	if awschaos.Spec.SecretName != nil {
		secret := &v1.Secret{}
		err := e.Client.Get(ctx, types.NamespacedName{
			Name:      *awschaos.Spec.SecretName,
			Namespace: awschaos.Namespace,
		}, secret)
		if err != nil {
			e.Log.Error(err, "fail to get cloud secret")
			return err
		}
		opts = append(opts, awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			string(secret.Data["aws_access_key_id"]),
			string(secret.Data["aws_secret_access_key"]),
			"",
		)))
	}
	cfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		e.Log.Error(err, "unable to load aws SDK config")
		return err
	}
	ec2client := ec2.NewFromConfig(cfg)

	awschaos.Finalizers = []string{AwsFinalizer}
	_, err = ec2client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
		InstanceIds: []string{awschaos.Spec.Ec2Instance},
	})

	if err != nil {
		awschaos.Finalizers = make([]string, 0)
		e.Log.Error(err, "fail to stop the instance")
		return err
	}

	return nil
}

func (e *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	awschaos, ok := chaos.(*v1alpha1.AwsChaos)
	if !ok {
		err := errors.New("chaos is not awschaos")
		e.Log.Error(err, "chaos is not AwsChaos", "chaos", chaos)
		return err
	}

	awschaos.Finalizers = make([]string, 0)
	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(awschaos.Spec.AwsRegion),
	}
	if awschaos.Spec.SecretName != nil {
		secret := &v1.Secret{}
		err := e.Client.Get(ctx, types.NamespacedName{
			Name:      *awschaos.Spec.SecretName,
			Namespace: awschaos.Namespace,
		}, secret)
		if err != nil {
			e.Log.Error(err, "fail to get cloud secret")
			return err
		}
		opts = append(opts, awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			string(secret.Data["aws_access_key_id"]),
			string(secret.Data["aws_secret_access_key"]),
			"",
		)))
	}
	cfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		e.Log.Error(err, "unable to load aws SDK config")
		return err
	}
	ec2client := ec2.NewFromConfig(cfg)

	_, err = ec2client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
		InstanceIds: []string{awschaos.Spec.Ec2Instance},
	})

	if err != nil {
		e.Log.Error(err, "fail to start the instance")
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

		return chaos.Spec.Action == v1alpha1.Ec2Stop
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
