package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/awschaos/action"
)

type stopper struct {
	c        *ec2.Client
	resource string
	selector *v1alpha1.AWSSelector
}

func NewStopper(cfg aws.Config, resource string, selector *v1alpha1.AWSSelector) action.Stopper {
	c := ec2.NewFromConfig(cfg)
	s := stopper{
		c:        c,
		resource: resource,
		selector: selector,
	}

	return &s
}

func (s *stopper) Stop(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	switch s.resource {
	case "instance":
		if err := s.StopInstances(ctx, snapshot); err != nil {
			return err
		}
	case "volume":
		if err := s.StopVolumes(ctx, snapshot); err != nil {
			return err
		}
	case "subnet":
		if err := s.StopSubnets(ctx, snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (s *stopper) Snapshot(ctx context.Context) (*v1alpha1.AWSStatusSnapshot, error) {
	switch s.resource {
	case "instance":
		return s.SnapshotInstances(ctx)
	case "volume":
		return s.SnapshotVolumes(ctx)
	case "subnet":
		return s.SnapshotSubnets(ctx)
	}
	return nil, fmt.Errorf("unknown resource: %v", s.resource)
}

type recoverer struct {
	c        *ec2.Client
	resource string
}

func NewRecoverer(cfg aws.Config, resource string) action.Recoverer {
	c := ec2.NewFromConfig(cfg)
	r := recoverer{
		c:        c,
		resource: resource,
	}

	return &r
}

func (r *recoverer) Recover(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	switch r.resource {
	case "instance":
		return r.RecoverInstances(ctx, snapshot)
	case "volume":
		return r.RecoverVolumes(ctx, snapshot)
	case "subnet":
		return r.RecoverSubnets(ctx, snapshot)
	}
	return nil
}
