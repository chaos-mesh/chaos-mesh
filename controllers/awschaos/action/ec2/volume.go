package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	DefaultMaxVolumeResults = 100
)

type attachment struct {
	volumeID   string
	instanceID string
	device     string
}

func (s *stopper) listVolumeAttachments(ctx context.Context, selector *v1alpha1.AWSSelector) ([]attachment, error) {
	filters := make([]types.Filter, 0, len(selector.Filters))
	for _, f := range selector.Filters {
		filters = append(filters, types.Filter{
			Name:   f.Name,
			Values: f.Values,
		})
	}
	vs := []attachment{}
	var nextToken *string
	for {
		out, err := s.c.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
			MaxResults: DefaultMaxVolumeResults,
			Filters:    filters,
			VolumeIds:  selector.IDs,
			NextToken:  nextToken,
		})
		if err != nil {
			return nil, err
		}
		for _, vol := range out.Volumes {
			for _, attach := range vol.Attachments {
				vs = append(vs, attachment{
					volumeID:   *attach.VolumeId,
					instanceID: *attach.InstanceId,
					device:     *attach.Device,
				})
			}
		}
		nextToken = out.NextToken
		if nextToken == nil {
			return vs, nil
		}
	}
}

func (s *stopper) SnapshotVolumes(ctx context.Context) (*v1alpha1.AWSStatusSnapshot, error) {
	vols, err := s.listVolumeAttachments(ctx, s.selector)
	if err != nil {
		return nil, err
	}
	snapshot := v1alpha1.AWSStatusSnapshot{}
	for _, vol := range vols {
		snapshot.Resources = append(snapshot.Resources, v1alpha1.AWSResource{
			Tuple: []string{
				vol.instanceID,
				vol.volumeID,
				vol.device,
			},
		})
	}
	return &snapshot, nil
}

func convertSnapshotToVolumeAttachments(snapshot *v1alpha1.AWSStatusSnapshot) ([]attachment, error) {
	vols := make([]attachment, 0, len(snapshot.Resources))
	for i, r := range snapshot.Resources {
		if len(r.Tuple) != 3 {
			return nil, fmt.Errorf("invalid snapshot, resource tuple %v is invalid: %v", i, r.Tuple)
		}
		vols = append(vols, attachment{
			instanceID: r.Tuple[0],
			volumeID:   r.Tuple[1],
			device:     r.Tuple[2],
		})
	}
	return vols, nil
}

func (s *stopper) StopVolumes(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	vols, err := convertSnapshotToVolumeAttachments(snapshot)
	if err != nil {
		return err
	}

	for _, vol := range vols {
		if _, err := s.c.DetachVolume(ctx, &ec2.DetachVolumeInput{
			InstanceId: &vol.instanceID,
			VolumeId:   &vol.volumeID,
			Force:      true,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *recoverer) RecoverVolumes(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error {
	vols, err := convertSnapshotToVolumeAttachments(snapshot)
	if err != nil {
		return err
	}
	for _, vol := range vols {
		if _, err := r.c.AttachVolume(ctx, &ec2.AttachVolumeInput{
			Device:     &vol.device,
			VolumeId:   &vol.volumeID,
			InstanceId: &vol.instanceID,
		}); err != nil {
			return err
		}
	}
	return nil
}
