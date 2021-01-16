package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	DefaultMaxInstanceResults = 500
)

func (s *stopper) listInstances(ctx context.Context, selector *v1alpha1.AWSSelector) ([]string, error) {
	ids := []string{}
	filters := make([]types.Filter, 0, len(selector.Filters))
	for _, f := range selector.Filters {
		filters = append(filters, types.Filter{
			Name:   f.Name,
			Values: f.Values,
		})
	}
	var nextToken *string
	for {
		out, err := s.c.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
			MaxResults:  DefaultMaxInstanceResults,
			Filters:     filters,
			InstanceIds: selector.IDs,
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, err
		}
		for _, reserved := range out.Reservations {
			for _, instance := range reserved.Instances {
				ids = append(ids, *instance.InstanceId)
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			return ids, nil
		}
	}
}

func (s *stopper) SnapshotInstances(ctx context.Context) (*v1alpha1.AWSStatusSnapshot, error) {
	ids, err := s.listInstances(ctx, s.selector)
	if err != nil {
		return nil, err
	}
	snapshot := v1alpha1.AWSStatusSnapshot{}
	for _, id := range ids {
		snapshot.Resources = append(snapshot.Resources, v1alpha1.AWSResource{
			Tuple: []string{
				id,
			},
		})
	}

	return &snapshot, nil
}

func (s *stopper) StopInstances(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	ids, err := convertSnapshotToInstanceIDs(snapshot)
	if err != nil {
		return err
	}
	if _, err := s.c.StopInstances(ctx, &ec2.StopInstancesInput{
		InstanceIds: ids,
		Force:       true,
	}); err != nil {
		return err
	}
	return nil
}

func convertSnapshotToInstanceIDs(snapshot *v1alpha1.AWSStatusSnapshot) ([]string, error) {
	ids := make([]string, 0, len(snapshot.Resources))
	for i, r := range snapshot.Resources {
		if len(r.Tuple) != 1 {
			return nil, fmt.Errorf("invalid snapshot, resource tuple %v is invalid: %v", i, r)
		}
		ids = append(ids, r.Tuple[0])
	}
	return ids, nil
}

func (r *recoverer) RecoverInstances(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	ids, err := convertSnapshotToInstanceIDs(snapshot)
	if err != nil {
		return err
	}
	if _, err := r.c.StartInstances(ctx, &ec2.StartInstancesInput{
		InstanceIds: ids,
	}); err != nil {
		return err
	}
	return nil
}
