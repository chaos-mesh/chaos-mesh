package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	DefaultMaxSubnetsResults    = 100
	DefaultMaxNetworkAclResults = 100
)

type subnet struct {
	vpcID    string
	azID     string
	subnetID string
}

type networkAssociation struct {
	vpcID    string
	subnetID string
	aclID    string
	assID    string
}

func listSubnets(ctx context.Context, c *ec2.Client, selector *v1alpha1.AWSSelector) ([]subnet, error) {
	subnets := []subnet{}
	filters := make([]types.Filter, 0, len(selector.Filters))
	for _, f := range selector.Filters {
		filters = append(filters, types.Filter{
			Name:   f.Name,
			Values: f.Values,
		})
	}
	var nextToken *string
	for {
		out, err := c.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{
			MaxResults: DefaultMaxSubnetsResults,
			Filters:    filters,
			SubnetIds:  selector.IDs,
			NextToken:  nextToken,
		})
		if err != nil {
			return nil, err
		}
		for _, s := range out.Subnets {
			subnets = append(subnets, subnet{
				vpcID:    *s.VpcId,
				azID:     *s.AvailabilityZoneId,
				subnetID: *s.SubnetId,
			})
		}
		nextToken = out.NextToken
		if nextToken == nil {
			return subnets, nil
		}
	}
}

func listNetworkAclAssociations(ctx context.Context, c *ec2.Client, subnets []subnet, withTag bool) ([]networkAssociation, error) {
	ns := []networkAssociation{}
	for _, sub := range subnets {
		subnetID := sub.subnetID
		var nextToken *string
		for {
			fs := []types.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []string{
						sub.vpcID,
					},
				},
				{
					Name: aws.String("association.subnet-id"),
					Values: []string{
						subnetID,
					},
				},
			}
			if withTag {
				fs = append(fs, types.Filter{
					Name:   aws.String("awschaos.chaos-mesh.org/autoclean"),
					Values: []string{"true"},
				})
			}

			out, err := c.DescribeNetworkAcls(ctx, &ec2.DescribeNetworkAclsInput{
				MaxResults: DefaultMaxNetworkAclResults,
				Filters:    fs,
				NextToken:  nextToken,
			})
			if err != nil {
				return nil, err
			}
			for _, acl := range out.NetworkAcls {
				for _, ass := range acl.Associations {
					if *ass.SubnetId == subnetID {
						ns = append(ns, networkAssociation{
							vpcID:    *acl.VpcId,
							subnetID: subnetID,
							aclID:    *ass.NetworkAclId,
							assID:    *ass.NetworkAclAssociationId,
						})
					}
				}
			}
			nextToken = out.NextToken
			if nextToken == nil {
				break
			}
		}
	}
	return ns, nil
}

func (s *stopper) createNetworkAcls(ctx context.Context, vpcID string) (string, error) {
	acl, err := s.c.CreateNetworkAcl(ctx, &ec2.CreateNetworkAclInput{
		VpcId: aws.String(vpcID),
		TagSpecifications: []types.TagSpecification{
			{
				ResourceType: types.ResourceTypeNetworkAcl,
				Tags: []types.Tag{
					{
						Key:   aws.String("awschaos.chaos-mesh.org/autoclean"),
						Value: aws.String("true"),
					},
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	aclID := acl.NetworkAcl.NetworkAclId

	// The rule deny all out bound traffic to anywhere
	engress := &ec2.CreateNetworkAclEntryInput{
		CidrBlock:    aws.String("0.0.0.0/0"),
		Egress:       true,
		NetworkAclId: aclID,
		PortRange: &types.PortRange{
			From: 53,
			To:   53,
		},
		Protocol:   aws.String("-1"),
		RuleAction: types.RuleActionDeny,
		RuleNumber: 100,
	}
	if _, err := s.c.CreateNetworkAclEntry(ctx, engress); err != nil {
		return "", err
	}

	// The rule deny all in bound traffic from anywhere
	ingress := &ec2.CreateNetworkAclEntryInput{
		CidrBlock:    aws.String("0.0.0.0/0"),
		Egress:       false,
		NetworkAclId: aclID,
		PortRange: &types.PortRange{
			From: 53,
			To:   53,
		},
		Protocol:   aws.String("-1"),
		RuleAction: types.RuleActionDeny,
		RuleNumber: 101,
	}

	if _, err := s.c.CreateNetworkAclEntry(ctx, ingress); err != nil {
		return "", err
	}
	return *aclID, nil
}

func (s *stopper) SnapshotSubnets(ctx context.Context) (*v1alpha1.AWSStatusSnapshot, error) {
	subnets, err := listSubnets(ctx, s.c, s.selector)
	if err != nil {
		return nil, err
	}
	ns, err := listNetworkAclAssociations(ctx, s.c, subnets, false)
	if err != nil {
		return nil, err
	}
	snapshot := v1alpha1.AWSStatusSnapshot{}
	for _, n := range ns {
		snapshot.Resources = append(snapshot.Resources, v1alpha1.AWSResource{
			Tuple: []string{
				n.vpcID,
				n.subnetID,
				n.aclID,
				n.assID,
			},
		})
	}
	return &snapshot, nil
}

func convertSnapshotToNetworkAssociations(snapshot *v1alpha1.AWSStatusSnapshot) ([]networkAssociation, error) {
	ns := make([]networkAssociation, 0, len(snapshot.Resources))
	for i, r := range snapshot.Resources {
		if len(r.Tuple) != 4 {
			return nil, fmt.Errorf("invalid snapshot, resource tuple %v is invalid: %v", i, r.Tuple)
		}
		ns = append(ns, networkAssociation{
			vpcID:    r.Tuple[0],
			subnetID: r.Tuple[1],
			aclID:    r.Tuple[2],
			assID:    r.Tuple[3],
		})
	}
	return ns, nil
}

func (s *stopper) StopSubnets(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	ns, err := convertSnapshotToNetworkAssociations(snapshot)
	if err != nil {
		return err
	}
	vpcToAcl := map[string]string{}
	for _, n := range ns {
		aclID, ok := vpcToAcl[n.vpcID]
		if !ok {
			aclID, err := s.createNetworkAcls(ctx, n.vpcID)
			if err != nil {
				return err
			}
			vpcToAcl[n.vpcID] = aclID
		}
		if _, err := s.c.ReplaceNetworkAclAssociation(ctx, &ec2.ReplaceNetworkAclAssociationInput{
			AssociationId: &n.assID,
			NetworkAclId:  &aclID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *recoverer) RecoverSubnets(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	olds, err := convertSnapshotToNetworkAssociations(snapshot)
	if err != nil {
		return err
	}
	subnets := make([]subnet, 0, len(olds))
	for _, old := range olds {
		subnets = append(subnets, subnet{
			vpcID:    old.vpcID,
			subnetID: old.subnetID,
		})
	}
	currents, err := listNetworkAclAssociations(ctx, r.c, subnets, true)
	if err != nil {
		return err
	}
	assToAcl := map[string]string{}

	aclIDs := map[string]struct{}{}

	for _, current := range currents {
		for _, old := range olds {
			if current.subnetID == old.subnetID {
				assToAcl[current.assID] = old.aclID
			}
		}
		aclIDs[current.aclID] = struct{}{}
	}
	for k, v := range assToAcl {
		assID := k
		aclID := v
		if _, err := r.c.ReplaceNetworkAclAssociation(ctx, &ec2.ReplaceNetworkAclAssociationInput{
			AssociationId: &assID,
			NetworkAclId:  &aclID,
		}); err != nil {
			return err
		}
	}

	for k := range aclIDs {
		aclID := k
		if _, err := r.c.DeleteNetworkAcl(ctx, &ec2.DeleteNetworkAclInput{
			NetworkAclId: &aclID,
		}); err != nil {
			return err
		}
	}

	return nil
}
