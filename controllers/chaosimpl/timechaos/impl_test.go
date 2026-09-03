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

package timechaos

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type testTimeDaemon struct {
	pb.ChaosDaemonClient
	applyErr  error
	applied   int
	recovered int
}

func (d *testTimeDaemon) Close() error { return nil }
func (d *testTimeDaemon) SetTimeOffset(context.Context, *pb.TimeRequest, ...grpc.CallOption) (*empty.Empty, error) {
	d.applied++
	return &empty.Empty{}, d.applyErr
}
func (d *testTimeDaemon) RecoverTimeOffset(context.Context, *pb.TimeRequest, ...grpc.CallOption) (*empty.Empty, error) {
	d.recovered++
	return &empty.Empty{}, nil
}

type testTimeDecoder struct {
	daemon *testTimeDaemon
	err    error
}

func (d *testTimeDecoder) DecodeContainerRecord(context.Context, *v1alpha1.Record, v1alpha1.InnerObject) (utils.DecodedContainerRecord, error) {
	return utils.DecodedContainerRecord{PbClient: d.daemon, ContainerId: "container", ContainerName: "app", Pod: &corev1.Pod{ObjectMeta: metav1.ObjectMeta{UID: "pod-uid"}}}, d.err
}

type testTimePodClient struct {
	client.Client
	getErr error
}

func (c *testTimePodClient) Get(context.Context, client.ObjectKey, client.Object, ...client.GetOption) error {
	return c.getErr
}

func timeChaosTestObject() *v1alpha1.TimeChaos {
	return &v1alpha1.TimeChaos{ObjectMeta: metav1.ObjectMeta{UID: "experiment-uid"}, Spec: v1alpha1.TimeChaosSpec{TimeOffset: "1s", ClockIds: []string{"CLOCK_REALTIME"}}}
}

func TestTimeChaosAmbiguousApplyKeepsCleanupPending(t *testing.T) {
	daemon := &testTimeDaemon{applyErr: errors.New("reply lost")}
	decoder := &testTimeDecoder{daemon: daemon}
	impl := &Impl{Log: logr.Discard(), decoder: decoder}
	records := []*v1alpha1.Record{{Id: "default/pod/app", Phase: v1alpha1.NotInjected}}
	phase, err := impl.Apply(context.Background(), 0, records, timeChaosTestObject())
	if phase != waitForApply || err == nil {
		t.Fatalf("ambiguous apply: phase=%s err=%v", phase, err)
	}
	records[0].Phase = phase
	daemon.applyErr = nil
	// A stopped record in Not Injected/Wait is driven through Apply and Recover
	// by the common records controller, rather than skipped as NotInjected.
	phase, err = impl.Apply(context.Background(), 0, records, timeChaosTestObject())
	if phase != v1alpha1.Injected || err != nil {
		t.Fatalf("retry apply: phase=%s err=%v", phase, err)
	}
	records[0].Phase = phase
	phase, err = impl.Recover(context.Background(), 0, records, timeChaosTestObject())
	if phase != v1alpha1.NotInjected || err != nil || daemon.recovered != 1 {
		t.Fatalf("cleanup: phase=%s err=%v recovered=%d", phase, err, daemon.recovered)
	}
}

func TestTimeChaosApplyPreservesPendingPhaseOnTransientDecodeError(t *testing.T) {
	decoder := &testTimeDecoder{daemon: &testTimeDaemon{}, err: utils.ErrContainerNotFound}
	podClient := &testTimePodClient{getErr: errors.New("API timeout")}
	impl := &Impl{Log: logr.Discard(), decoder: decoder, Client: podClient}
	records := []*v1alpha1.Record{{Id: "default/pod/app", Phase: waitForApply}}
	phase, err := impl.Apply(context.Background(), 0, records, timeChaosTestObject())
	if phase != waitForApply || err == nil {
		t.Fatalf("lost pending cleanup: phase=%s err=%v", phase, err)
	}
	podClient.getErr = apierrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "pod")
	phase, err = impl.Apply(context.Background(), 0, records, timeChaosTestObject())
	if phase != v1alpha1.NotInjected || err != nil {
		t.Fatalf("confirmed missing pod: phase=%s err=%v", phase, err)
	}
}
