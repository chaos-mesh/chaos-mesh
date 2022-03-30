package diskchaos

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/disk"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log     logr.Logger
	decoder *utils.ContainerRecordDecoder
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	pbClient := decodedContainer.PbClient
	containerId := decodedContainer.ContainerId
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	diskchaos := obj.(*v1alpha1.DiskChaos)

	switch diskchaos.Spec.Action {
	case v1alpha1.Fill:
		_, err = pbClient.DiskFill(ctx, &pb.DiskFillRequest{
			Path:            diskchaos.Spec.Path,
			Size:            diskchaos.Spec.Size,
			Percent:         diskchaos.Spec.Percent,
			SpaceLockSize:   diskchaos.Spec.SpaceLockSize,
			FillByFallocate: diskchaos.Spec.FillByFAllocate,
			ContainerId:     containerId,
		})
		if err != nil {
			return v1alpha1.NotInjected, errors.WithStack(err)
		}
	case v1alpha1.DRead:
		_, err = pbClient.DiskPayload(ctx, &pb.DiskPayloadRequest{
			Path:          diskchaos.Spec.Path,
			Size:          diskchaos.Spec.Size,
			Percent:       diskchaos.Spec.Percent,
			SpaceLockSize: diskchaos.Spec.SpaceLockSize,
			ProcessNum:    uint32(diskchaos.Spec.ProcessNum),
			Action:        int32(disk.Read),
			ContainerId:   containerId,
		})
		if err != nil {
			return v1alpha1.NotInjected, errors.WithStack(err)
		}
	case v1alpha1.DWrite:
		_, err = pbClient.DiskPayload(ctx, &pb.DiskPayloadRequest{
			Path:          diskchaos.Spec.Path,
			Size:          diskchaos.Spec.Size,
			Percent:       diskchaos.Spec.Percent,
			SpaceLockSize: diskchaos.Spec.SpaceLockSize,
			ProcessNum:    uint32(diskchaos.Spec.ProcessNum),
			Action:        int32(disk.Write),
			ContainerId:   containerId,
		})
		if err != nil {
			return v1alpha1.NotInjected, errors.WithStack(err)
		}
	default:
		return v1alpha1.NotInjected, errors.New("unexpected disk chaos action")
	}
	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	decodedContainer, err := impl.decoder.DecodeContainerRecord(ctx, records[index], obj)
	pbClient := decodedContainer.PbClient
	containerId := decodedContainer.ContainerId
	if pbClient != nil {
		defer pbClient.Close()
	}
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	diskchaos := obj.(*v1alpha1.DiskChaos)

	switch diskchaos.Spec.Action {
	case v1alpha1.Fill:
		_, err = pbClient.RecoverDiskFill(ctx, &pb.DiskFillRecoverRequest{ContainerId: containerId})
		if err != nil {
			return v1alpha1.NotInjected, errors.WithStack(err)
		}
	case v1alpha1.DWrite:
		_, err = pbClient.RecoverDiskPayload(ctx, &pb.DiskPayloadRecoverRequest{ContainerId: containerId})
		if err != nil {
			return v1alpha1.NotInjected, errors.WithStack(err)
		}
	case v1alpha1.DRead:
		_, err = pbClient.RecoverDiskPayload(ctx, &pb.DiskPayloadRecoverRequest{ContainerId: containerId})
		if err != nil {
			return v1alpha1.NotInjected, errors.WithStack(err)
		}
	default:
		return v1alpha1.NotInjected, errors.New("unexpected disk chaos action")
	}
	return v1alpha1.Injected, nil
}

func NewImpl(c client.Client, log logr.Logger, decoder *utils.ContainerRecordDecoder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "diskchaos",
		Object: &v1alpha1.DiskChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("diskchaos"),
			decoder: decoder,
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
