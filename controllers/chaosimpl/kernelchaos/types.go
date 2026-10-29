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

package kernelchaos

import (
	"context"
	"encoding/hex"
	"net"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	pbkernel "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
)

var _ impltypes.ChaosImpl = (*Impl)(nil)

type Impl struct {
	client.Client
	Log logr.Logger

	chaosDaemonClientBuilder *chaosdaemon.ClientBuilder
}

// Apply reserves a kernel fault before a later reconciliation arms it.
func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.KernelChaos)
	record := records[index]
	binding, bound := chaos.Status.KernelFaults[record.Id]
	if chaos.Status.Experiment.DesiredPhase == v1alpha1.StoppedPhase {
		return impl.Recover(ctx, index, records, obj)
	}
	if bound {
		if record.Phase == v1alpha1.NotInjected {
			return v1alpha1.KernelReserved, nil
		}
		if record.Phase != v1alpha1.KernelReserved {
			return record.Phase, errors.Errorf("cannot arm kernel fault in phase %q", record.Phase)
		}
		if err := impl.arm(ctx, binding); err != nil {
			if status.Code(err) == codes.NotFound || status.Code(err) == codes.Aborted {
				return impl.Recover(ctx, index, records, obj)
			}
			return v1alpha1.KernelReserved, err
		}
		return v1alpha1.Injected, nil
	}
	if record.Phase != v1alpha1.NotInjected {
		return record.Phase, missingBinding(record.Id)
	}

	podID, containerName, err := controller.ParseNamespacedNameContainer(record.Id)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	var pod v1.Pod
	if err := impl.Client.Get(ctx, podID, &pod); err != nil {
		if k8sError.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.NotInjected, err
	}
	containerID, err := containerIDForName(&pod, containerName)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	binding, err = impl.reserve(ctx, &pod, chaos, containerID)
	if err != nil {
		return v1alpha1.NotInjected, err
	}
	if chaos.Status.KernelFaults == nil {
		chaos.Status.KernelFaults = make(map[string]v1alpha1.KernelFaultBinding)
	}
	chaos.Status.KernelFaults[record.Id] = binding
	return v1alpha1.KernelReserved, nil
}

// Recover releases the original service-owned resource without resolving the target again.
func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	chaos := obj.(*v1alpha1.KernelChaos)
	record := records[index]
	binding, bound := chaos.Status.KernelFaults[record.Id]
	if !bound {
		if record.Phase == v1alpha1.NotInjected {
			return v1alpha1.NotInjected, nil
		}
		return record.Phase, missingBinding(record.Id)
	}
	if err := impl.release(ctx, binding); err != nil {
		return record.Phase, err
	}
	delete(chaos.Status.KernelFaults, record.Id)
	return v1alpha1.NotInjected, nil
}

func missingBinding(recordID string) error {
	return errors.Errorf("kernel fault binding is missing for record %s; recovery of the original fault is required", recordID)
}

func containerIDForName(pod *v1.Pod, containerName string) (string, error) {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName && status.ContainerID != "" {
			return status.ContainerID, nil
		}
	}
	return "", errors.Wrapf(utils.ErrContainerNotFound, "container %s in pod %s/%s has no runtime ID", containerName, pod.Namespace, pod.Name)
}

func (impl *Impl) reserve(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.KernelChaos, containerID string) (v1alpha1.KernelFaultBinding, error) {
	var binding v1alpha1.KernelFaultBinding
	var node v1.Node
	if err := impl.Client.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
		return binding, err
	}
	if node.UID == "" || node.Status.NodeInfo.BootID == "" {
		return binding, errors.Errorf("node %s has no UID or boot ID", node.Name)
	}
	address, err := impl.chaosDaemonClientBuilder.FindDaemonIPByNode(ctx, node.Name)
	if err != nil {
		return binding, err
	}
	conn, err := kernelConnection(address)
	if err != nil {
		return binding, err
	}
	defer conn.Close()

	daemon, err := impl.chaosDaemonClientBuilder.Build(ctx, pod, &types.NamespacedName{Namespace: chaos.Namespace, Name: chaos.Name})
	if err != nil {
		return binding, err
	}
	defer daemon.Close()
	container, err := daemon.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{Action: pb.ContainerAction_GETPID}, ContainerId: containerID,
	})
	if err != nil {
		return binding, err
	}
	if container == nil || container.Pid == 0 {
		return binding, errors.New("daemon returned an empty container PID")
	}

	var callchain []*pbkernel.FailKernRequestFrame
	for _, frame := range chaos.Spec.FailKernRequest.Callchain {
		callchain = append(callchain, &pbkernel.FailKernRequestFrame{
			Funcname: frame.Funcname, Parameters: frame.Parameters, Predicate: frame.Predicate,
		})
	}
	handle, err := pbkernel.NewBPFKIServiceClient(conn).ReserveKernelFault(ctx, &pbkernel.ReserveKernelFaultRequest{
		NodeBootId: node.Status.NodeInfo.BootID,
		Fault: &pbkernel.FailKernRequest{
			Pid: container.Pid, Ftype: pbkernel.FailKernRequest_FAILTYPE(chaos.Spec.FailKernRequest.FailType),
			Headers: chaos.Spec.FailKernRequest.Headers, Callchain: callchain,
			Probability: float32(chaos.Spec.FailKernRequest.Probability) / 100, Times: chaos.Spec.FailKernRequest.Times,
		},
	})
	if err != nil {
		return binding, err
	}
	if handle == nil || handle.Session == nil {
		return binding, errors.New("kernel service returned an empty reservation")
	}
	binding = v1alpha1.KernelFaultBinding{
		NodeName: node.Name, NodeUID: node.UID, Address: address, Handle: handle.Id,
		ServiceInstance: handle.Session.Instance, ServiceBootID: handle.Session.BootId,
		ServicePID: handle.Session.Pid, ServiceStartTime: handle.Session.StartTime, ServicePIDNamespace: handle.Session.PidNamespace,
	}
	if err := validateBinding(binding); err != nil {
		return v1alpha1.KernelFaultBinding{}, err
	}
	if binding.ServiceBootID != node.Status.NodeInfo.BootID {
		return v1alpha1.KernelFaultBinding{}, errors.New("kernel reservation came from a different node boot")
	}
	return binding, nil
}

func (impl *Impl) arm(ctx context.Context, binding v1alpha1.KernelFaultBinding) error {
	reference, err := impl.reference(ctx, binding)
	if err != nil {
		return err
	}
	conn, err := kernelConnection(binding.Address)
	if err != nil {
		return err
	}
	defer conn.Close()
	response, err := pbkernel.NewBPFKIServiceClient(conn).ArmKernelFault(ctx, reference)
	if err != nil {
		return err
	}
	if response == nil || response.Ret != 0 {
		return errors.Errorf("kernel service failed to arm reservation: %v", response)
	}
	return nil
}

func (impl *Impl) release(ctx context.Context, binding v1alpha1.KernelFaultBinding) error {
	reference, err := impl.reference(ctx, binding)
	if err != nil {
		return err
	}
	originalErr := releaseAt(ctx, binding.Address, reference)
	if originalErr == nil {
		return nil
	}
	address, err := impl.chaosDaemonClientBuilder.FindDaemonIPByNode(ctx, binding.NodeName)
	if err != nil {
		return errors.Wrapf(originalErr, "cannot locate kernel service on original node %s: %v", binding.NodeName, err)
	}
	if address == binding.Address {
		return originalErr
	}
	return releaseAt(ctx, address, reference)
}

func releaseAt(ctx context.Context, address string, reference *pbkernel.KernelFaultReference) error {
	conn, err := kernelConnection(address)
	if err != nil {
		return err
	}
	defer conn.Close()
	response, err := pbkernel.NewBPFKIServiceClient(conn).ReleaseKernelFault(ctx, reference)
	if err != nil {
		return err
	}
	if response != nil {
		switch response.Outcome {
		case pbkernel.ReleaseKernelFaultResponse_RELEASED, pbkernel.ReleaseKernelFaultResponse_ABSENT, pbkernel.ReleaseKernelFaultResponse_SERVICE_GONE:
			return nil
		}
	}
	return errors.Errorf("kernel service did not confirm release: %v", response)
}

func (impl *Impl) reference(ctx context.Context, binding v1alpha1.KernelFaultBinding) (*pbkernel.KernelFaultReference, error) {
	if err := validateBinding(binding); err != nil {
		return nil, err
	}
	var node v1.Node
	if err := impl.Client.Get(ctx, types.NamespacedName{Name: binding.NodeName}, &node); err != nil {
		return nil, err
	}
	if node.UID != binding.NodeUID || node.Status.NodeInfo.BootID == "" {
		return nil, errors.Errorf("original kernel fault node %s is replaced or has no boot ID", binding.NodeName)
	}
	return &pbkernel.KernelFaultReference{
		NodeBootId: node.Status.NodeInfo.BootID,
		Handle: &pbkernel.KernelFaultHandle{
			Id: binding.Handle,
			Session: &pbkernel.KernelFaultSession{
				Instance: binding.ServiceInstance, BootId: binding.ServiceBootID, Pid: binding.ServicePID,
				StartTime: binding.ServiceStartTime, PidNamespace: binding.ServicePIDNamespace,
			},
		},
	}, nil
}

func validateBinding(binding v1alpha1.KernelFaultBinding) error {
	if binding.NodeName == "" || binding.NodeUID == "" || net.ParseIP(binding.Address) == nil ||
		!opaqueID(binding.Handle) || !opaqueID(binding.ServiceInstance) || binding.ServiceBootID == "" ||
		binding.ServicePID == 0 || binding.ServiceStartTime == 0 || binding.ServicePIDNamespace == 0 {
		return errors.New("kernel fault binding has incomplete or invalid ownership information")
	}
	return nil
}

func opaqueID(value string) bool {
	if len(value) != 32 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func kernelConnection(address string) (*grpc.ClientConn, error) {
	if net.ParseIP(address) == nil {
		return nil, errors.Errorf("invalid kernel service IP %q", address)
	}
	return grpcUtils.Builder(address, config.ControllerCfg.BPFKIPort).WithDefaultTimeout().Insecure().Build()
}

func NewImpl(c client.Client, log logr.Logger, builder *chaosdaemon.ClientBuilder) *impltypes.ChaosImplPair {
	return &impltypes.ChaosImplPair{
		Name:   "kernelchaos",
		Object: &v1alpha1.KernelChaos{},
		Impl: &Impl{
			Client: c, Log: log.WithName("kernelchaos"), chaosDaemonClientBuilder: builder,
		},
		ObjectList: &v1alpha1.KernelChaosList{},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
