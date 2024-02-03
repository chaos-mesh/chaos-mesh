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

package aws_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/require"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/aws"
)

// StubClient implements the interface
type StubClient struct {
	Input  *ec2.DescribeInstancesInput
	Output *ec2.DescribeInstancesOutput
}

func (s StubClient) DescribeInstances(ctx context.Context, in *ec2.DescribeInstancesInput, opt ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
	if s.Input != nil {
		*s.Input = *in
	}
	return s.Output, nil
}

func TestSelect(t *testing.T) {
	ctx := context.Background()

	sel := &v1alpha1.AWSSelector{
		Filters: []*v1alpha1.AWSFilter{{
			Name:   "tag:Stack",
			Values: []string{"staging"},
		}},
		Mode: v1alpha1.OneMode,
	}

	ec2client := StubClient{
		Input:  &ec2.DescribeInstancesInput{},
		Output: buildInstancesOutput("1111", "2222", "3333"),
	}
	defer mock.With("MockCreateEc2Client", ec2client)()

	s := selector.New(
		selector.SelectorParams{
			AWSSelector: &aws.SelectImpl{},
		})

	result, err := s.Select(ctx, sel)

	require.NoError(t, err)
	require.NotNil(t, result)

	require.Len(t, result, 1)
	require.Subset(t,
		[]string{"1111", "2222", "3333"},
		[]string{result[0].(*aws.Instance).InstanceID},
	)
	require.Equal(t, &ec2.DescribeInstancesInput{
		Filters: []ec2types.Filter{{
			Name:   ptr.String("tag:Stack"),
			Values: []string{"staging"},
		}},
	}, ec2client.Input)
}

func buildInstancesOutput(instanceIDs ...string) *ec2.DescribeInstancesOutput {
	reservations := []ec2types.Reservation{}

	for _, instanceID := range instanceIDs {
		reservations = append(reservations, ec2types.Reservation{
			Instances: []ec2types.Instance{{
				InstanceId: &instanceID,
			}},
		})
	}

	return &ec2.DescribeInstancesOutput{Reservations: reservations}
}
