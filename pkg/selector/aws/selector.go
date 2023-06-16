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

package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

// EC2Client defines the minimum client interface required for this package
type EC2Client interface {
	DescribeInstances(context.Context, *ec2.DescribeInstancesInput, ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

type SelectImpl struct {
	c client.Client
	generic.Option
}

type Instance struct {
	InstanceID string
	AWSRegion  string
	Endpoint   *string
	SecretName *string
	EbsVolume  *string
	DeviceName *string
}

func (instance *Instance) Id() string {
	json, _ := json.Marshal(instance)

	return string(json)
}

func (impl *SelectImpl) Select(ctx context.Context, awsSelector *v1alpha1.AWSSelector) ([]*Instance, error) {
	if len(awsSelector.Filters) == 0 {
		return []*Instance{{
			InstanceID: awsSelector.Ec2Instance,
			Endpoint:   awsSelector.Endpoint,
			AWSRegion:  awsSelector.AWSRegion,
			SecretName: awsSelector.SecretName,
			EbsVolume:  awsSelector.EbsVolume,
			DeviceName: awsSelector.DeviceName,
		}}, nil
	}

	// we have filters, so we should lookup the cloud resources
	instances := []*Instance{}

	ec2client, err := impl.newEc2Client(ctx, awsSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	result, err := ec2client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: buildEc2Filters(awsSelector.Filters),
	})
	if err != nil {
		return instances, err
	}
	for _, r := range result.Reservations {
		// Set the Ec2Instance, and copy over the other attributes, except the filter
		instances = append(instances, &Instance{
			InstanceID: *r.Instances[0].InstanceId,
			Endpoint:   awsSelector.Endpoint,
			AWSRegion:  awsSelector.AWSRegion,
			SecretName: awsSelector.SecretName,
			EbsVolume:  awsSelector.EbsVolume,
			DeviceName: awsSelector.DeviceName,
		})
	}
	mode := awsSelector.Mode
	value := awsSelector.Value

	filteredInstances, err := filterInstancesByMode(instances, mode, value)
	if err != nil {
		return nil, err
	}

	return filteredInstances, nil
}

type Params struct {
	fx.In

	Client client.Client
}

func New(params Params) *SelectImpl {
	return &SelectImpl{
		params.Client,
		generic.Option{
			TargetNamespace: config.ControllerCfg.TargetNamespace,
		},
	}
}

func buildEc2Filters(filters []*v1alpha1.AWSFilter) []ec2types.Filter {

	ec2Filters := []ec2types.Filter{}
	for _, filter := range filters {
		ec2Filters = append(ec2Filters, ec2types.Filter{
			Name:   aws.String(filter.Name),
			Values: filter.Values,
		})
	}
	return ec2Filters
}

func (impl *SelectImpl) newEc2Client(ctx context.Context, awsSelector *v1alpha1.AWSSelector) (EC2Client, error) {

	if ec2client := mock.On("MockCreateEc2Client"); ec2client != nil {
		return ec2client.(EC2Client), nil
	}
	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(awsSelector.AWSRegion),
	}

	if awsSelector.Endpoint != nil {
		opts = append(opts, awscfg.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: *awsSelector.Endpoint, SigningRegion: region}, nil
		})))
	}

	if awsSelector.SecretName != nil {
		secret := &v1.Secret{}
		err := impl.c.Get(ctx, types.NamespacedName{
			Name:      *awsSelector.SecretName,
			Namespace: impl.TargetNamespace,
		}, secret)
		if err != nil {
			return nil, fmt.Errorf("fail to get cloud secret: %w", err)
		}
		opts = append(opts, awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			string(secret.Data["aws_access_key_id"]),
			string(secret.Data["aws_secret_access_key"]),
			"",
		)))
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return ec2.NewFromConfig(cfg), nil
}

// filterInstancesByMode filters instances by mode from a list
func filterInstancesByMode(instances []*Instance, mode v1alpha1.SelectorMode, value string) ([]*Instance, error) {
	indexes, err := generic.FilterObjectsByMode(mode, value, len(instances))
	if err != nil {
		return nil, err
	}

	var filtered []*Instance

	for _, index := range indexes {
		index := index
		filtered = append(filtered, instances[index])
	}
	return filtered, nil
}
