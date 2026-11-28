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

package chaosdaemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type timeImageCall struct {
	pid  int
	vars map[string]uint64
}

type timeRuntimeFixture struct {
	crclients.ContainerRuntimeInfoClient
	pids map[string]uint32
}

func (f timeRuntimeFixture) GetPidFromContainerID(_ context.Context, id string) (uint32, error) {
	pid, ok := f.pids[id]
	if !ok {
		return 0, fmt.Errorf("unknown fixture container %q", id)
	}
	return pid, nil
}

type timeImageFixture struct {
	attached, recovered   []timeImageCall
	injectErr, recoverErr error
}

func (f *timeImageFixture) AttachToProcess(pid int, vars map[string]uint64) error {
	f.attached = append(f.attached, timeImageCall{pid, vars})
	return f.injectErr
}

func (f *timeImageFixture) Recover(pid int, vars map[string]uint64) error {
	f.recovered = append(f.recovered, timeImageCall{pid, vars})
	return f.recoverErr
}

func TestTimeOffsetOverlappingExperiments(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	// Synthetic PID with a test double for the vDSO operations.
	runtime := &httpChaosRuntimeClientStub{containerPID: 5000000}
	server := NewDaemonServerWithCRClient(runtime, nil, logr.Discard())
	t.Cleanup(func() { server.backgroundProcessManager.Shutdown(context.Background()) })
	image := &timeImageFixture{}
	cleanup := mock.With("MockTimeSkewProcessImage", image)
	t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })

	container := "ns/pod/app"
	req := &pb.TimeRequest{ContainerId: "containerd://fixture", PodContainerName: container, Uid: "first", Sec: 2, Nsec: 10, ClkIdsMask: 1}
	second := &pb.TimeRequest{ContainerId: req.ContainerId, PodContainerName: container, Uid: "second", Sec: -1, Nsec: 20, ClkIdsMask: 4}
	failure := errors.New("fixture runtime unavailable")
	runtime.containerErr = failure
	_, err := server.SetTimeOffset(ctx, req)
	g.Expect(err).To(MatchError(failure))
	_, err = server.RecoverTimeOffset(ctx, req)
	g.Expect(err).To(MatchError(failure))
	g.Expect(image.attached).To(BeEmpty())
	runtime.containerErr = nil

	image.injectErr = errors.New("fixture failed to apply")
	_, err = server.SetTimeOffset(ctx, req)
	g.Expect(err).To(MatchError(ContainSubstring(image.injectErr.Error())))
	id := tasks.PodContainerName(container)
	g.Expect(server.timeChaosServer.manager.GetUIDsWithPID(id)).To(BeEmpty())
	image.injectErr = nil
	image.attached = nil

	_, err = server.SetTimeOffset(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(image.attached).To(HaveLen(2))
	g.Expect(image.attached[0]).To(Equal(timeImageCall{int(runtime.containerPID), map[string]uint64{
		"CLOCK_IDS_MASK": 1, "TV_SEC_DELTA": 2, "TV_NSEC_DELTA": 10,
	}}))
	_, err = server.SetTimeOffset(ctx, second)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(image.attached).To(HaveLen(4))
	g.Expect(image.attached[2]).To(Equal(timeImageCall{int(runtime.containerPID), map[string]uint64{
		"CLOCK_IDS_MASK": 5, "TV_SEC_DELTA": 1, "TV_NSEC_DELTA": 30,
	}}))
	_, err = server.RecoverTimeOffset(ctx, second)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(image.attached).To(HaveLen(6))
	g.Expect(image.attached[4]).To(Equal(image.attached[0]))
	g.Expect(image.recovered).To(BeEmpty())

	image.recoverErr = errors.New("fixture failed to recover")
	_, err = server.RecoverTimeOffset(ctx, req)
	g.Expect(err).To(MatchError(ContainSubstring(image.recoverErr.Error())))
	g.Expect(server.timeChaosServer.manager.GetUIDsWithPID(id)).To(Equal([]tasks.TaskID{timeOffsetTaskID(req.Uid, id)}))
	image.recoverErr = nil
	_, err = server.RecoverTimeOffset(ctx, req)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(image.recovered).To(HaveLen(4))
	g.Expect(image.recovered[2:]).To(Equal(image.attached[:2]))
	g.Expect(server.timeChaosServer.manager.GetUIDsWithPID(id)).To(BeEmpty())
}

func TestTimeOffsetRetryAndMultipleContainers(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	runtime := timeRuntimeFixture{pids: map[string]uint32{"first-runtime": 5000001, "second-runtime": 5000002}}
	server := NewDaemonServerWithCRClient(runtime, nil, logr.Discard())
	t.Cleanup(func() { server.backgroundProcessManager.Shutdown(context.Background()) })
	image := &timeImageFixture{}
	cleanup := mock.With("MockTimeSkewProcessImage", image)
	t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })
	g.Expect(mock.On("MockTimeSkewProcessImage")).To(BeIdenticalTo(image))

	first := &pb.TimeRequest{ContainerId: "first-runtime", PodContainerName: "pod-uid:first", Uid: "experiment-uid-pod-uid", Sec: 2, ClkIdsMask: 1}
	second := &pb.TimeRequest{ContainerId: "second-runtime", PodContainerName: "pod-uid:second", Uid: first.Uid, Sec: 2, ClkIdsMask: 1}
	for _, req := range []*pb.TimeRequest{first, second, first} {
		_, err := server.SetTimeOffset(ctx, req)
		g.Expect(err).NotTo(HaveOccurred())
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err := server.SetTimeOffset(canceled, first)
	g.Expect(err).To(MatchError(context.Canceled))
	_, err = server.RecoverTimeOffset(canceled, first)
	g.Expect(err).To(MatchError(context.Canceled))
	g.Expect(image.attached).To(HaveLen(6))
	g.Expect(image.recovered).To(BeEmpty())

	// A retried injection must not add a second offset. Recovery of one
	// container must also leave the other container's task registered.
	_, err = server.RecoverTimeOffset(ctx, first)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = server.RecoverTimeOffset(ctx, second)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(image.recovered).To(HaveLen(4))
	g.Expect(image.recovered[0]).To(Equal(timeImageCall{5000001, map[string]uint64{
		"CLOCK_IDS_MASK": 1, "TV_SEC_DELTA": 2, "TV_NSEC_DELTA": 0,
	}}))
	g.Expect(image.recovered[2]).To(Equal(timeImageCall{5000002, image.recovered[0].vars}))
}

type concurrentTimeImage struct {
	mu        sync.Mutex
	recovered []int
}

func (*concurrentTimeImage) AttachToProcess(int, map[string]uint64) error { return nil }

func (f *concurrentTimeImage) Recover(pid int, _ map[string]uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recovered = append(f.recovered, pid)
	return nil
}

func TestTimeOffsetConcurrentTargets(t *testing.T) {
	g := NewWithT(t)
	runtime := timeRuntimeFixture{pids: map[string]uint32{}}
	server := NewDaemonServerWithCRClient(runtime, nil, logr.Discard())
	t.Cleanup(func() { server.backgroundProcessManager.Shutdown(context.Background()) })
	image := &concurrentTimeImage{}
	cleanup := mock.With("MockTimeSkewProcessImage", image)
	t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })
	g.Expect(mock.On("MockTimeSkewProcessImage")).To(BeIdenticalTo(image))

	var requests []*pb.TimeRequest
	var recovered []int
	for i := range 8 {
		id, pid := fmt.Sprintf("container-%d", i), 5000100+i
		runtime.pids[id] = uint32(pid)
		requests = append(requests, &pb.TimeRequest{ContainerId: id, PodContainerName: id, Uid: id, Sec: 2})
		recovered = append(recovered, pid, pid)
	}
	start := make(chan struct{})
	group, ctx := errgroup.WithContext(t.Context())
	for _, req := range requests {
		group.Go(func() error {
			<-start
			if _, err := server.SetTimeOffset(ctx, req); err != nil {
				return err
			}
			_, err := server.RecoverTimeOffset(ctx, req)
			return err
		})
	}
	close(start)
	g.Expect(group.Wait()).To(Succeed())
	g.Expect(image.recovered).To(ConsistOf(recovered))
}
