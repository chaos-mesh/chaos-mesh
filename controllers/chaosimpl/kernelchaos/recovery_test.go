// Copyright 2026 Chaos Mesh Authors.
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

package kernelchaos_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/kernelchaos"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	commonrecords "github.com/chaos-mesh/chaos-mesh/controllers/common/records"
	controllerconfig "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	pbkernel "github.com/chaos-mesh/chaos-mesh/pkg/chaoskernel/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	containerselector "github.com/chaos-mesh/chaos-mesh/pkg/selector/container"
	podselector "github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

const kernelRecordID = "kernel-ci/target/selected"

func TestKernelChaosPersistsReservationBeforeArm(t *testing.T) {
	f := newKernelLifecycleFixture(t)
	f.api.failWrites = 1
	require.True(t, f.reconcile().Requeue)
	require.Nil(t, f.load().Status.Experiment.Records)
	require.Empty(t, f.load().Status.KernelFaults)
	require.Equal(t, 1, f.rpc.snapshot().reserves)
	require.Zero(t, f.rpc.snapshot().arms)
	require.Empty(t, f.rpc.snapshot().active)
	require.Equal(t, v1alpha1.KernelReserved, f.api.attempts[0].Status.Experiment.Records[0].Phase)
	require.Contains(t, f.api.attempts[0].Status.KernelFaults, kernelRecordID)

	before := f.load()
	f.reconcile()
	reserved := f.load()
	require.Equal(t, v1alpha1.KernelReserved, reserved.Status.Experiment.Records[0].Phase)
	binding := reserved.Status.KernelFaults[kernelRecordID]
	require.NotEmpty(t, binding.Handle)
	require.Zero(t, f.rpc.snapshot().arms)
	require.True(t, (common.StatusRecordEventsChangePredicate{}).Update(event.UpdateEvent{
		ObjectOld: before, ObjectNew: reserved,
	}), "persisting the reservation must enqueue the next reconciliation")

	f.rpc.change(func(rpc *kernelLifecycleRPC) {
		rpc.armGuard = func(reference *pbkernel.KernelFaultReference) error {
			var persisted v1alpha1.KernelChaos
			if err := f.kube.Get(f.ctx, f.key, &persisted); err != nil {
				return err
			}
			if persisted.Status.Experiment.Records[0].Phase != v1alpha1.KernelReserved ||
				persisted.Status.KernelFaults[kernelRecordID].Handle != reference.Handle.Id {
				return errors.New("Arm reached the service without a durable matching reservation")
			}
			return nil
		}
	})
	f.reconcile()
	injected := f.load()
	require.Equal(t, v1alpha1.Injected, injected.Status.Experiment.Records[0].Phase)
	require.Equal(t, 1, injected.Status.Experiment.Records[0].InjectedCount)
	require.Equal(t, binding, injected.Status.KernelFaults[kernelRecordID])
	require.Equal(t, map[string]uint32{binding.Handle: 4242}, f.rpc.snapshot().active)
	require.Equal(t, 1, f.rpc.snapshot().armWrites)
}

func TestKernelChaosRestartsTargetBetweenReservationAndArm(t *testing.T) {
	f := newKernelLifecycleFixture(t)
	f.reconcile()
	first := f.load().Status.KernelFaults[kernelRecordID]
	require.Equal(t, v1alpha1.KernelReserved, f.load().Status.Experiment.Records[0].Phase)
	require.Equal(t, map[string]uint32{first.Handle: 4242}, f.rpc.snapshot().pending)

	var pod v1.Pod
	require.NoError(t, f.kube.Get(f.ctx, client.ObjectKey{Namespace: "kernel-ci", Name: "target"}, &pod))
	pod.Status.ContainerStatuses[0].ContainerID = "containerd://replacement"
	require.NoError(t, f.kube.Status().Update(f.ctx, &pod))
	f.rpc.change(func(rpc *kernelLifecycleRPC) { rpc.pid = 9000 })
	podReads := f.api.podReads
	f.reconcile()
	retired := f.load()
	snapshot := f.rpc.snapshot()
	require.Equal(t, v1alpha1.NotInjected, retired.Status.Experiment.Records[0].Phase)
	require.Empty(t, retired.Status.KernelFaults)
	require.Empty(t, snapshot.active)
	require.Empty(t, snapshot.pending)
	require.Zero(t, snapshot.armWrites)
	require.Equal(t, 1, snapshot.arms)
	require.Equal(t, 1, snapshot.releases)
	require.Equal(t, first.Handle, snapshot.references[0].Handle.Id)
	require.Equal(t, 1, snapshot.reserves)
	require.Equal(t, 1, snapshot.pidCalls, "retiring a stale reservation must not resolve the replacement PID")
	require.Equal(t, podReads, f.api.podReads, "retiring a stale reservation must not read the target Pod")

	f.reconcile()
	reserved := f.load()
	second := reserved.Status.KernelFaults[kernelRecordID]
	snapshot = f.rpc.snapshot()
	require.NotEqual(t, first.Handle, second.Handle)
	require.Equal(t, v1alpha1.KernelReserved, reserved.Status.Experiment.Records[0].Phase)
	require.Equal(t, map[string]uint32{second.Handle: 9000}, snapshot.pending)
	require.Empty(t, snapshot.active)
	require.Equal(t, 2, snapshot.reserves)
	require.Equal(t, 2, snapshot.pidCalls)
	require.Equal(t, 1, snapshot.arms)
	f.reconcile()
	require.Equal(t, v1alpha1.Injected, f.load().Status.Experiment.Records[0].Phase)
	require.Equal(t, map[string]uint32{second.Handle: 9000}, f.rpc.snapshot().active)
	require.Empty(t, f.rpc.snapshot().pending)
	require.Equal(t, 1, f.rpc.snapshot().armWrites)
}

func TestKernelChaosRetriesArmWithoutResettingFault(t *testing.T) {
	for _, mode := range []string{"lost reply", "status write rejected"} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.reconcile()
			binding := f.load().Status.KernelFaults[kernelRecordID]
			if mode == "lost reply" {
				f.rpc.change(func(rpc *kernelLifecycleRPC) { rpc.loseArmReply = true })
			} else {
				f.api.failWrites = 1
			}
			require.True(t, f.reconcile().Requeue)
			require.Equal(t, v1alpha1.KernelReserved, f.load().Status.Experiment.Records[0].Phase)
			require.Equal(t, binding, f.load().Status.KernelFaults[kernelRecordID])
			require.Equal(t, map[string]uint32{binding.Handle: 4242}, f.rpc.snapshot().active)
			f.rpc.change(func(rpc *kernelLifecycleRPC) { rpc.pid = 9000 })

			f.reconcile()
			require.Equal(t, v1alpha1.Injected, f.load().Status.Experiment.Records[0].Phase)
			require.Equal(t, 2, f.rpc.snapshot().arms)
			require.Equal(t, 1, f.rpc.snapshot().armWrites)
			require.Equal(t, 1, f.rpc.snapshot().reserves)
			f.reconcile()
			require.Equal(t, 2, f.rpc.snapshot().arms)
		})
	}
}

func TestKernelChaosStopReservedReleasesWithoutArm(t *testing.T) {
	f := newKernelLifecycleFixture(t)
	f.reconcile()
	f.stop()
	f.reconcile()
	require.Equal(t, v1alpha1.NotInjected, f.load().Status.Experiment.Records[0].Phase)
	require.Empty(t, f.load().Status.KernelFaults)
	require.Zero(t, f.rpc.snapshot().arms)
	require.Equal(t, 1, f.rpc.snapshot().releases)
	require.Empty(t, f.rpc.snapshot().active)
}

func TestKernelChaosReleaseRetriesKeepOriginalHandle(t *testing.T) {
	for _, mode := range []string{"RPC unavailable", "lost reply", "status write rejected"} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.inject()
			binding := f.load().Status.KernelFaults[kernelRecordID]
			f.stop()
			switch mode {
			case "RPC unavailable":
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					rpc.releaseErr = status.Error(codes.Unavailable, "temporary failure")
				})
			case "lost reply":
				f.rpc.change(func(rpc *kernelLifecycleRPC) { rpc.loseReleaseReply = true })
			case "status write rejected":
				f.api.failWrites = 1
			}
			require.True(t, f.reconcile().Requeue)
			require.Equal(t, binding, f.load().Status.KernelFaults[kernelRecordID])
			require.Equal(t, v1alpha1.Injected, f.load().Status.Experiment.Records[0].Phase)
			f.rpc.change(func(rpc *kernelLifecycleRPC) { rpc.releaseErr = nil })
			f.reconcile()
			require.Empty(t, f.load().Status.KernelFaults)
			require.Equal(t, v1alpha1.NotInjected, f.load().Status.Experiment.Records[0].Phase)
			require.Empty(t, f.rpc.snapshot().active)
			require.Equal(t, 2, f.rpc.snapshot().releases)
			require.Equal(t, binding.Handle, f.rpc.snapshot().references[0].Handle.Id)
			require.True(t, proto.Equal(f.rpc.snapshot().references[0], f.rpc.snapshot().references[1]))
			f.reconcile()
			require.Equal(t, 2, f.rpc.snapshot().releases)
			require.Equal(t, 1, f.rpc.snapshot().pidCalls)
		})
	}
}

func TestKernelChaosRecoversOriginalFaultAfterTargetChanges(t *testing.T) {
	for _, mode := range []string{"container restarted", "Pod deleted", "Pod recreated on another node"} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.inject()
			binding := f.load().Status.KernelFaults[kernelRecordID]
			var pod v1.Pod
			require.NoError(t, f.kube.Get(f.ctx, client.ObjectKey{Namespace: "kernel-ci", Name: "target"}, &pod))
			switch mode {
			case "container restarted":
				pod.Status.ContainerStatuses[0].ContainerID = "containerd://replacement"
				require.NoError(t, f.kube.Status().Update(f.ctx, &pod))
			case "Pod deleted":
				require.NoError(t, f.kube.Delete(f.ctx, &pod))
			case "Pod recreated on another node":
				require.NoError(t, f.kube.Delete(f.ctx, &pod))
				pod.ResourceVersion = ""
				pod.UID = "replacement-pod"
				pod.Spec.NodeName = "node-b"
				pod.Status.ContainerStatuses[0].ContainerID = "containerd://replacement"
				require.NoError(t, f.kube.Create(f.ctx, &pod))
			}
			f.rpc.change(func(rpc *kernelLifecycleRPC) {
				rpc.pid = 9000
				rpc.owned["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"] = &kernelOwnedFault{
					fault: &pbkernel.FailKernRequest{Pid: 9000}, active: true,
				}
			})
			podReads := f.api.podReads
			f.stop()
			f.reconcile()
			require.Equal(t, podReads, f.api.podReads, "recovery must not read the target Pod")
			require.Equal(t, 1, f.rpc.snapshot().pidCalls, "recovery must not resolve a replacement PID")
			require.Equal(t, map[string]uint32{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": 9000}, f.rpc.snapshot().active)
			require.Equal(t, binding.Handle, f.rpc.snapshot().references[0].Handle.Id)
			require.Equal(t, kernelTestBootID, f.rpc.snapshot().references[0].NodeBootId)
			require.Equal(t, "node-a", f.api.nodeReads[len(f.api.nodeReads)-1])
			require.Empty(t, f.load().Status.KernelFaults)
		})
	}
}

func TestKernelChaosRetainsBindingWhenCleanupIsUncertain(t *testing.T) {
	for _, mode := range []string{
		"original node missing", "original node replaced", "node boot unknown", "node API forbidden",
		"service unavailable", "unspecified release", "unknown release", "old backend",
		"original service still alive",
	} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.inject()
			binding := f.load().Status.KernelFaults[kernelRecordID]
			var node v1.Node
			require.NoError(t, f.kube.Get(f.ctx, client.ObjectKey{Name: "node-a"}, &node))
			switch mode {
			case "original node missing":
				require.NoError(t, f.kube.Delete(f.ctx, &node))
			case "original node replaced":
				require.NoError(t, f.kube.Delete(f.ctx, &node))
				node.ResourceVersion = ""
				node.UID = "replacement-node"
				require.NoError(t, f.kube.Create(f.ctx, &node))
			case "node boot unknown":
				node.Status.NodeInfo.BootID = ""
				require.NoError(t, f.kube.Status().Update(f.ctx, &node))
			case "node API forbidden":
				f.api.nodeErr = apierrors.NewForbidden(schema.GroupResource{Resource: "nodes"}, "node-a", errors.New("denied"))
			case "service unavailable":
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					rpc.releaseErr = status.Error(codes.Unavailable, "service unavailable")
				})
			case "unspecified release":
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					outcome := pbkernel.ReleaseKernelFaultResponse_UNSPECIFIED
					rpc.releaseOutcome = &outcome
				})
			case "unknown release":
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					outcome := pbkernel.ReleaseKernelFaultResponse_Outcome(99)
					rpc.releaseOutcome = &outcome
				})
			case "old backend":
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					rpc.releaseErr = status.Error(codes.Unimplemented, "old service")
				})
			case "original service still alive":
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					rpc.session.Instance = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
				})
			}
			f.stop()
			require.True(t, f.reconcile().Requeue)
			require.Equal(t, binding, f.load().Status.KernelFaults[kernelRecordID])
			require.Equal(t, v1alpha1.Injected, f.load().Status.Experiment.Records[0].Phase)
			require.Equal(t, 1, f.rpc.snapshot().pidCalls)
			require.Zero(t, f.rpc.snapshot().legacyCalls)
		})
	}
}

func TestKernelChaosOldBackendAndLegacyRecordsFailClosed(t *testing.T) {
	t.Run("old backend does not fall back to legacy injection", func(t *testing.T) {
		f := newKernelLifecycleFixture(t)
		f.rpc.change(func(rpc *kernelLifecycleRPC) {
			rpc.reserveErr = status.Error(codes.Unimplemented, "old service")
		})
		require.True(t, f.reconcile().Requeue)
		require.Equal(t, v1alpha1.NotInjected, f.load().Status.Experiment.Records[0].Phase)
		require.Empty(t, f.load().Status.KernelFaults)
		require.Zero(t, f.rpc.snapshot().arms)
		require.Zero(t, f.rpc.snapshot().legacyCalls)
	})
	for _, phase := range []v1alpha1.Phase{v1alpha1.Injected, v1alpha1.KernelReserved} {
		t.Run(string(phase), func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			chaos := f.load()
			chaos.Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
			chaos.Status.Experiment.Records = []*v1alpha1.Record{{Id: kernelRecordID, Phase: phase}}
			require.NoError(t, f.kube.Update(f.ctx, chaos))
			require.NoError(t, f.kube.Delete(f.ctx, &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "kernel-ci", Name: "target"}}))
			require.True(t, f.reconcile().Requeue)
			record := f.load().Status.Experiment.Records[0]
			require.Equal(t, phase, record.Phase)
			require.Contains(t, record.Events[0].Message, "binding is missing")
			require.Zero(t, f.rpc.snapshot().pidCalls)
			require.Zero(t, f.rpc.snapshot().releases)
			require.Zero(t, f.rpc.snapshot().legacyCalls)
		})
	}
}

func TestKernelChaosBindingWithoutReservedPhaseCannotArm(t *testing.T) {
	f := newKernelLifecycleFixture(t)
	f.reconcile()
	chaos := f.load()
	chaos.Status.Experiment.Records[0].Phase = v1alpha1.NotInjected
	require.NoError(t, f.kube.Update(f.ctx, chaos))
	f.reconcile()
	require.Equal(t, v1alpha1.KernelReserved, f.load().Status.Experiment.Records[0].Phase)
	require.Zero(t, f.rpc.snapshot().arms)
	f.reconcile()
	require.Equal(t, v1alpha1.Injected, f.load().Status.Experiment.Records[0].Phase)
	require.Equal(t, 1, f.rpc.snapshot().arms)
}

func TestKernelChaosRetiresExpiredReservationBeforeReservingAgain(t *testing.T) {
	f := newKernelLifecycleFixture(t)
	f.reconcile()
	first := f.load().Status.KernelFaults[kernelRecordID]
	f.rpc.change(func(rpc *kernelLifecycleRPC) { delete(rpc.owned, first.Handle) })
	f.reconcile()
	require.Equal(t, v1alpha1.NotInjected, f.load().Status.Experiment.Records[0].Phase)
	require.Empty(t, f.load().Status.KernelFaults)
	require.Equal(t, 1, f.rpc.snapshot().reserves)
	require.Empty(t, f.rpc.snapshot().active)
	f.reconcile()
	second := f.load().Status.KernelFaults[kernelRecordID]
	require.NotEqual(t, first.Handle, second.Handle)
	require.Equal(t, v1alpha1.KernelReserved, f.load().Status.Experiment.Records[0].Phase)
	require.Empty(t, f.rpc.snapshot().active)
	f.reconcile()
	require.Equal(t, map[string]uint32{second.Handle: 4242}, f.rpc.snapshot().active)
}

func TestKernelChaosAcceptsConfirmedServiceTermination(t *testing.T) {
	for _, mode := range []string{"process gone", "node rebooted"} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.inject()
			f.rpc.change(func(rpc *kernelLifecycleRPC) {
				rpc.session.Instance = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
				rpc.originalServiceGone = true
				rpc.owned = map[string]*kernelOwnedFault{
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {fault: &pbkernel.FailKernRequest{Pid: 9000}, active: true},
				}
				if mode == "node rebooted" {
					rpc.session.BootId = "66666666-7777-8888-9999-aaaaaaaaaaaa"
				}
			})
			if mode == "node rebooted" {
				var node v1.Node
				require.NoError(t, f.kube.Get(f.ctx, client.ObjectKey{Name: "node-a"}, &node))
				node.Status.NodeInfo.BootID = "66666666-7777-8888-9999-aaaaaaaaaaaa"
				require.NoError(t, f.kube.Status().Update(f.ctx, &node))
			}
			f.stop()
			f.reconcile()
			require.Empty(t, f.load().Status.KernelFaults)
			require.Equal(t, map[string]uint32{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": 9000}, f.rpc.snapshot().active)
			if mode == "node rebooted" {
				require.Equal(t, "66666666-7777-8888-9999-aaaaaaaaaaaa", f.rpc.snapshot().references[0].NodeBootId)
				require.Equal(t, kernelTestBootID, f.rpc.snapshot().references[0].Handle.Session.BootId)
			}
		})
	}
}

func TestKernelChaosReleaseUsesOriginalNodeEndpoints(t *testing.T) {
	for _, mode := range []string{"original reachable", "replacement confirmed", "replacement uncertain", "wrong node", "wrong boot"} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.inject()
			binding := f.load().Status.KernelFaults[kernelRecordID]
			replacement := &kernelLifecycleRPC{
				session: proto.Clone(kernelTestHandle().Session).(*pbkernel.KernelFaultSession),
				owned: map[string]*kernelOwnedFault{
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {fault: &pbkernel.FailKernRequest{Pid: 9000}, active: true},
				},
			}
			replacement.session.Instance = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
			replacement.originalServiceGone = mode == "replacement confirmed"
			if mode == "wrong boot" {
				replacement.session.BootId = "66666666-7777-8888-9999-aaaaaaaaaaaa"
			}
			serveKernelTestRPCAt(t, fmt.Sprintf("[::1]:%d", controllerconfig.ControllerCfg.BPFKIPort), func(server *grpc.Server) {
				pbkernel.RegisterBPFKIServiceServer(server, replacement)
			})
			var endpoints discoveryv1.EndpointSlice
			require.NoError(t, f.kube.Get(f.ctx, client.ObjectKey{Namespace: "chaos-mesh", Name: "chaos-daemon-kernel"}, &endpoints))
			endpoints.AddressType = discoveryv1.AddressTypeIPv6
			endpoints.Endpoints[0].Addresses = []string{"::1"}
			if mode == "wrong node" {
				nodeName := "node-b"
				endpoints.Endpoints[0].NodeName = &nodeName
			}
			require.NoError(t, f.kube.Update(f.ctx, &endpoints))
			if mode != "original reachable" {
				f.rpc.change(func(rpc *kernelLifecycleRPC) {
					rpc.releaseErr = status.Error(codes.Unavailable, "original endpoint unavailable")
				})
			}
			f.stop()
			result := f.reconcile()
			if mode == "original reachable" || mode == "replacement confirmed" {
				require.False(t, result.Requeue)
				require.Empty(t, f.load().Status.KernelFaults)
			} else {
				require.True(t, result.Requeue)
				require.Equal(t, binding, f.load().Status.KernelFaults[kernelRecordID])
			}
			require.Equal(t, 1, f.rpc.snapshot().releases)
			if mode == "original reachable" || mode == "wrong node" {
				require.Zero(t, replacement.snapshot().releases)
			} else {
				require.Equal(t, 1, replacement.snapshot().releases)
				require.Equal(t, binding.Handle, replacement.snapshot().references[0].Handle.Id)
				require.Equal(t, kernelTestBootID, replacement.snapshot().references[0].NodeBootId)
			}
			require.Equal(t, map[string]uint32{"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": 9000}, replacement.snapshot().active)
			require.Equal(t, 1, f.rpc.snapshot().pidCalls)
		})
	}
}

func TestKernelChaosRejectsMalformedReservation(t *testing.T) {
	for _, mode := range []string{"empty", "wrong boot", "missing process", "invalid handle"} {
		t.Run(mode, func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.rpc.change(func(rpc *kernelLifecycleRPC) {
				rpc.reserveResult = kernelTestHandle()
				switch mode {
				case "empty":
					rpc.reserveResult = &pbkernel.KernelFaultHandle{}
				case "wrong boot":
					rpc.reserveResult.Session.BootId = "different-boot"
				case "missing process":
					rpc.reserveResult.Session.StartTime = 0
				case "invalid handle":
					rpc.reserveResult.Id = "invalid"
				}
			})
			require.True(t, f.reconcile().Requeue)
			require.Empty(t, f.load().Status.KernelFaults)
			require.Equal(t, v1alpha1.NotInjected, f.load().Status.Experiment.Records[0].Phase)
			require.Zero(t, f.rpc.snapshot().arms)
			require.Empty(t, f.rpc.snapshot().active)
		})
	}
}

func TestKernelChaosRetiresChangedSessionOnlyAfterConfirmedRelease(t *testing.T) {
	for _, gone := range []bool{false, true} {
		t.Run(fmt.Sprintf("service gone %v", gone), func(t *testing.T) {
			f := newKernelLifecycleFixture(t)
			f.reconcile()
			binding := f.load().Status.KernelFaults[kernelRecordID]
			f.rpc.change(func(rpc *kernelLifecycleRPC) {
				rpc.session.Instance = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
				rpc.originalServiceGone = gone
			})
			result := f.reconcile()
			require.Equal(t, 1, f.rpc.snapshot().reserves)
			require.Equal(t, 1, f.rpc.snapshot().arms)
			require.Equal(t, 1, f.rpc.snapshot().releases)
			require.Empty(t, f.rpc.snapshot().active)
			if gone {
				require.False(t, result.Requeue)
				require.Equal(t, v1alpha1.NotInjected, f.load().Status.Experiment.Records[0].Phase)
				require.Empty(t, f.load().Status.KernelFaults)
			} else {
				require.True(t, result.Requeue)
				require.Equal(t, v1alpha1.KernelReserved, f.load().Status.Experiment.Records[0].Phase)
				require.Equal(t, binding, f.load().Status.KernelFaults[kernelRecordID])
			}
		})
	}
}

type kernelLifecycleFixture struct {
	t        *testing.T
	ctx      context.Context
	kube     client.Client
	api      *kernelAPIClient
	rpc      *kernelLifecycleRPC
	key      types.NamespacedName
	selector *selector.Selector
}

func newKernelLifecycleFixture(t *testing.T) *kernelLifecycleFixture {
	t.Helper()
	rpc := &kernelLifecycleRPC{
		pid: 4242, session: proto.Clone(kernelTestHandle().Session).(*pbkernel.KernelFaultSession),
		owned: make(map[string]*kernelOwnedFault),
	}
	daemonPort := serveKernelTestRPC(t, func(server *grpc.Server) { pb.RegisterChaosDaemonServer(server, rpc) })
	kernelPort := serveKernelTestRPC(t, func(server *grpc.Server) { pbkernel.RegisterBPFKIServiceServer(server, rpc) })
	previousConfig := controllerconfig.ControllerCfg
	controllerconfig.ControllerCfg = &config.ChaosControllerConfig{
		Namespace: "chaos-mesh", ClusterScoped: true, MaxEvents: 20,
		ChaosDaemonPort: daemonPort, BPFKIPort: kernelPort,
	}
	t.Cleanup(func() { controllerconfig.ControllerCfg = previousConfig })
	scheme := runtime.NewScheme()
	require.NoError(t, v1.AddToScheme(scheme))
	require.NoError(t, discoveryv1.AddToScheme(scheme))
	require.NoError(t, v1alpha1.AddToScheme(scheme))
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Namespace: "kernel-ci", Name: "target", UID: "original-pod"},
		Spec:       v1.PodSpec{NodeName: "node-a", Containers: []v1.Container{{Name: "selected"}}},
		Status: v1.PodStatus{
			Phase:             v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{{Name: "selected", ContainerID: "containerd://original"}},
		},
	}
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-a", UID: "original-node"},
		Status:     v1.NodeStatus{NodeInfo: v1.NodeSystemInfo{BootID: kernelTestBootID}},
	}
	endpoints := &discoveryv1.EndpointSlice{
		ObjectMeta:  metav1.ObjectMeta{Namespace: "chaos-mesh", Name: "chaos-daemon-kernel"},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints:   []discoveryv1.Endpoint{{NodeName: &pod.Spec.NodeName, Addresses: []string{"127.0.0.1"}}},
	}
	chaos := &v1alpha1.KernelChaos{
		ObjectMeta: metav1.ObjectMeta{Namespace: "kernel-ci", Name: "kernel-lifecycle", UID: "chaos-uid"},
		Spec: v1alpha1.KernelChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode: v1alpha1.AllMode, Selector: v1alpha1.PodSelectorSpec{Pods: map[string][]string{"kernel-ci": {"target"}}},
				},
				ContainerNames: []string{"selected"},
			},
			FailKernRequest: v1alpha1.FailKernRequest{FailType: 1, Probability: 25, Times: 7},
		},
		Status: v1alpha1.KernelChaosStatus{
			ChaosStatus: v1alpha1.ChaosStatus{Experiment: v1alpha1.ExperimentStatus{DesiredPhase: v1alpha1.RunningPhase}},
		},
	}
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod, node, endpoints, chaos).Build()
	api := &kernelAPIClient{Client: kube}
	return &kernelLifecycleFixture{
		t: t, ctx: context.Background(), kube: kube, api: api, rpc: rpc,
		key: types.NamespacedName{Namespace: chaos.Namespace, Name: chaos.Name},
		selector: selector.New(selector.Params{
			PodSelector:       podselector.New(podselector.Params{Client: api, Reader: api}),
			ContainerSelector: containerselector.New(containerselector.Params{Client: api, Reader: api}),
		}),
	}
}

func (f *kernelLifecycleFixture) reconcile() ctrl.Result {
	f.t.Helper()
	impl := kernelchaos.NewImpl(f.api, logr.Discard(), &chaosdaemon.ClientBuilder{Reader: f.api}).Impl
	reconciler := &commonrecords.Reconciler{
		Impl: impl, Object: &v1alpha1.KernelChaos{}, Client: f.api, Reader: f.api,
		Recorder: recorder.NewDebugRecorder(), Selector: f.selector, Log: logr.Discard(),
	}
	result, err := reconciler.Reconcile(f.ctx, ctrl.Request{NamespacedName: f.key})
	require.NoError(f.t, err)
	return result
}

func (f *kernelLifecycleFixture) load() *v1alpha1.KernelChaos {
	f.t.Helper()
	chaos := &v1alpha1.KernelChaos{}
	require.NoError(f.t, f.kube.Get(f.ctx, f.key, chaos))
	return chaos
}

func (f *kernelLifecycleFixture) inject() {
	f.t.Helper()
	f.reconcile()
	f.reconcile()
	require.Equal(f.t, v1alpha1.Injected, f.load().Status.Experiment.Records[0].Phase)
}

func (f *kernelLifecycleFixture) stop() {
	f.t.Helper()
	chaos := f.load()
	chaos.Status.Experiment.DesiredPhase = v1alpha1.StoppedPhase
	require.NoError(f.t, f.kube.Update(f.ctx, chaos))
}

type kernelAPIClient struct {
	client.Client
	failWrites int
	podReads   int
	nodeReads  []string
	nodeErr    error
	attempts   []*v1alpha1.KernelChaos
}

func (c *kernelAPIClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	switch obj.(type) {
	case *v1.Pod:
		c.podReads++
	case *v1.Node:
		c.nodeReads = append(c.nodeReads, key.Name)
		if c.nodeErr != nil {
			return c.nodeErr
		}
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *kernelAPIClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if chaos, ok := obj.(*v1alpha1.KernelChaos); ok {
		c.attempts = append(c.attempts, chaos.DeepCopy())
		if c.failWrites > 0 {
			c.failWrites--
			return errors.New("injected Kubernetes write failure")
		}
	}
	return c.Client.Update(ctx, obj, opts...)
}

type kernelOwnedFault struct {
	fault  *pbkernel.FailKernRequest
	active bool
}

type kernelLifecycleRPC struct {
	pb.UnimplementedChaosDaemonServer
	pbkernel.UnimplementedBPFKIServiceServer
	mu                  sync.Mutex
	pid                 uint32
	session             *pbkernel.KernelFaultSession
	owned               map[string]*kernelOwnedFault
	nextHandle          int
	pidCalls            int
	reserves            int
	arms                int
	releases            int
	armWrites           int
	legacyCalls         int
	references          []*pbkernel.KernelFaultReference
	reserveErr          error
	reserveResult       *pbkernel.KernelFaultHandle
	releaseErr          error
	loseArmReply        bool
	loseReleaseReply    bool
	originalServiceGone bool
	releaseOutcome      *pbkernel.ReleaseKernelFaultResponse_Outcome
	armGuard            func(*pbkernel.KernelFaultReference) error
}

type kernelRPCSnapshot struct {
	pidCalls, reserves, arms, releases, armWrites, legacyCalls int
	active                                                     map[string]uint32
	pending                                                    map[string]uint32
	references                                                 []*pbkernel.KernelFaultReference
}

func (rpc *kernelLifecycleRPC) snapshot() kernelRPCSnapshot {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	snapshot := kernelRPCSnapshot{
		pidCalls: rpc.pidCalls, reserves: rpc.reserves, arms: rpc.arms, releases: rpc.releases,
		armWrites: rpc.armWrites, legacyCalls: rpc.legacyCalls, active: make(map[string]uint32),
		pending: make(map[string]uint32),
	}
	for id, fault := range rpc.owned {
		if fault.active {
			snapshot.active[id] = fault.fault.Pid
		} else {
			snapshot.pending[id] = fault.fault.Pid
		}
	}
	for _, reference := range rpc.references {
		snapshot.references = append(snapshot.references, proto.Clone(reference).(*pbkernel.KernelFaultReference))
	}
	return snapshot
}

func (rpc *kernelLifecycleRPC) change(change func(*kernelLifecycleRPC)) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	change(rpc)
}

func (rpc *kernelLifecycleRPC) ContainerGetPid(context.Context, *pb.ContainerRequest) (*pb.ContainerResponse, error) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.pidCalls++
	return &pb.ContainerResponse{Pid: rpc.pid}, nil
}

func (rpc *kernelLifecycleRPC) ReserveKernelFault(_ context.Context, request *pbkernel.ReserveKernelFaultRequest) (*pbkernel.KernelFaultHandle, error) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.reserves++
	if rpc.reserveErr != nil {
		return nil, rpc.reserveErr
	}
	if request.NodeBootId != rpc.session.BootId {
		return nil, status.Error(codes.FailedPrecondition, "wrong node boot")
	}
	if request.Fault == nil || request.Fault.Pid == 0 || request.Fault.Pid != rpc.pid {
		return nil, status.Error(codes.NotFound, "target process is absent")
	}
	if rpc.reserveResult != nil {
		return proto.Clone(rpc.reserveResult).(*pbkernel.KernelFaultHandle), nil
	}
	rpc.nextHandle++
	id := fmt.Sprintf("%032x", rpc.nextHandle)
	rpc.owned[id] = &kernelOwnedFault{fault: proto.Clone(request.Fault).(*pbkernel.FailKernRequest)}
	return &pbkernel.KernelFaultHandle{
		Id: id, Session: proto.Clone(rpc.session).(*pbkernel.KernelFaultSession),
	}, nil
}

func (rpc *kernelLifecycleRPC) ArmKernelFault(_ context.Context, reference *pbkernel.KernelFaultReference) (*pbkernel.StatusResponse, error) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.arms++
	if rpc.armGuard != nil {
		if err := rpc.armGuard(reference); err != nil {
			return nil, err
		}
	}
	if reference.NodeBootId != rpc.session.BootId {
		return nil, status.Error(codes.FailedPrecondition, "wrong node boot")
	}
	if !proto.Equal(reference.Handle.Session, rpc.session) {
		return nil, status.Error(codes.Aborted, "different service")
	}
	fault, ok := rpc.owned[reference.Handle.Id]
	if !ok {
		return nil, status.Error(codes.NotFound, "expired reservation")
	}
	if !fault.active {
		if fault.fault.Pid != rpc.pid {
			return nil, status.Error(codes.NotFound, "reserved target process is absent")
		}
		fault.active = true
		rpc.armWrites++
	}
	if rpc.loseArmReply {
		rpc.loseArmReply = false
		return nil, status.Error(codes.Unavailable, "Arm reply lost")
	}
	return &pbkernel.StatusResponse{}, nil
}

func (rpc *kernelLifecycleRPC) ReleaseKernelFault(_ context.Context, reference *pbkernel.KernelFaultReference) (*pbkernel.ReleaseKernelFaultResponse, error) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.releases++
	rpc.references = append(rpc.references, proto.Clone(reference).(*pbkernel.KernelFaultReference))
	if rpc.releaseErr != nil {
		return nil, rpc.releaseErr
	}
	if reference.NodeBootId != rpc.session.BootId {
		return nil, status.Error(codes.FailedPrecondition, "wrong node boot")
	}
	if rpc.releaseOutcome != nil {
		return &pbkernel.ReleaseKernelFaultResponse{Outcome: *rpc.releaseOutcome}, nil
	}
	if !proto.Equal(reference.Handle.Session, rpc.session) {
		if rpc.originalServiceGone {
			return &pbkernel.ReleaseKernelFaultResponse{Outcome: pbkernel.ReleaseKernelFaultResponse_SERVICE_GONE}, nil
		}
		return nil, status.Error(codes.FailedPrecondition, "original service has not been proved dead")
	}
	outcome := pbkernel.ReleaseKernelFaultResponse_ABSENT
	if _, ok := rpc.owned[reference.Handle.Id]; ok {
		delete(rpc.owned, reference.Handle.Id)
		outcome = pbkernel.ReleaseKernelFaultResponse_RELEASED
	}
	if rpc.loseReleaseReply {
		rpc.loseReleaseReply = false
		return nil, status.Error(codes.Unavailable, "Release reply lost")
	}
	return &pbkernel.ReleaseKernelFaultResponse{Outcome: outcome}, nil
}

func (rpc *kernelLifecycleRPC) FailMMOrBIO(context.Context, *pbkernel.FailKernRequest) (*pbkernel.StatusResponse, error) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.legacyCalls++
	return &pbkernel.StatusResponse{}, nil
}

func (rpc *kernelLifecycleRPC) RecoverMMOrBIO(context.Context, *pbkernel.FailKernRequest) (*pbkernel.StatusResponse, error) {
	rpc.mu.Lock()
	defer rpc.mu.Unlock()
	rpc.legacyCalls++
	return &pbkernel.StatusResponse{}, nil
}
