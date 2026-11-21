// Copyright 2022 Chaos Mesh Authors.
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
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

type blockDelayCall struct {
	path          string
	operation     int
	pid           uint
	delay, jitter int64
	correlation   float64
}

type blockDriverFixture struct {
	delayCalls            []blockDelayCall
	recoveries            []int
	closes                int
	injectErr, recoverErr error
}

func (f *blockDriverFixture) InjectIOEMDelay(path string, op int, pid uint, delay, jitter int64, correlation float64) (int, error) {
	f.delayCalls = append(f.delayCalls, blockDelayCall{path, op, pid, delay, jitter, correlation})
	return 73, f.injectErr
}

func (f *blockDriverFixture) Recover(id int) error {
	f.recoveries = append(f.recoveries, id)
	return f.recoverErr
}

func (f *blockDriverFixture) Close() error {
	f.closes++
	return nil
}

func TestBlockChaosDeviceAndDriverLifecycle(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	roots := blockDeviceRoots{dev: t.TempDir(), sysBlock: t.TempDir()}
	const disk = "testdisk"
	device := filepath.Join(roots.dev, disk)
	scheduler := filepath.Join(roots.sysBlock, disk, "queue", "scheduler")
	g.Expect(os.MkdirAll(filepath.Dir(scheduler), 0700)).To(Succeed())
	installFixture := func(name string, value any) {
		t.Helper()
		cleanup := mock.With(name, value)
		t.Cleanup(func() { g.Expect(cleanup()).To(Succeed()) })
	}
	installFixture("MockBlockDeviceRoots", roots)
	driver := &blockDriverFixture{}
	var driverErr error
	installFixture("MockBlockChaosDriver", func() (blockChaosDriver, error) { return driver, driverErr })
	normalizeFails := false
	normalizations := 0
	installFixture("MockProcessBuild", func(ctx context.Context, command string, args ...string) *exec.Cmd {
		g.Expect(filepath.Base(command)).To(Equal("nsexec"))
		g.Expect(args).To(Equal([]string{"-l", "-m", "/proc/1/ns/mnt", "--", "cdh", "normalize-volume-name", "/mount/fixture"}))
		normalizations++
		if normalizeFails {
			return daemonTestCommand(ctx, "invalid fixture volume", true)
		}
		return daemonTestCommand(ctx, disk, false)
	})
	runtime := &httpChaosRuntimeClientStub{containerPID: 82}
	server := &DaemonServer{crClient: runtime, rootLogger: logr.Discard()}
	req := &pb.ApplyBlockChaosRequest{ContainerId: "containerd://fixture", VolumePath: "/mount/fixture"}
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(ContainSubstring("delay specification is required")))
	req.Action = 99
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError("unknown action"))
	g.Expect(normalizations).To(BeZero())
	req.Action = pb.ApplyBlockChaosRequest_Delay
	req.Delay = &pb.BlockDelaySpec{Delay: 15, Jitter: 2, Correlation: 0.5}

	normalizeFails = true
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(ContainSubstring("normalize volume name")))
	normalizeFails = false
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(ContainSubstring("reading schedulers")))
	g.Expect(os.WriteFile(scheduler, []byte("[none] mq-deadline\n"), 0600)).To(Succeed())
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(ContainSubstring("ioem scheduler not found")))
	g.Expect(runtime.containerCalls).To(BeZero())

	g.Expect(os.WriteFile(scheduler, []byte("[none] ioem-mq\n"), 0600)).To(Succeed())
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(ContainSubstring("does not exist")))
	g.Expect(os.ReadFile(scheduler)).To(Equal([]byte("ioem-mq")))
	g.Expect(os.WriteFile(scheduler, []byte("none [ioem-mq]\n"), 0600)).To(Succeed())
	g.Expect(os.WriteFile(device, nil, 0600)).To(Succeed())
	failure := errors.New("fixture unavailable")
	runtime.containerErr = failure
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(failure))
	runtime.containerErr = nil
	driverErr = failure
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(failure))
	g.Expect(server.RecoverBlockChaos(ctx, &pb.RecoverBlockChaosRequest{})).Error().To(MatchError(failure))
	g.Expect(driver.closes).To(BeZero())
	driverErr = nil
	driver.injectErr = failure
	g.Expect(server.ApplyBlockChaos(ctx, req)).Error().To(MatchError(failure))
	driver.injectErr = nil
	g.Expect(server.ApplyBlockChaos(ctx, req)).To(Equal(&pb.ApplyBlockChaosResponse{InjectionId: 73}))
	want := blockDelayCall{device, 0, 82, 15, 2, 0.5}
	g.Expect(driver.delayCalls).To(Equal([]blockDelayCall{want, want}))
	g.Expect(os.ReadFile(scheduler)).To(Equal([]byte("none [ioem-mq]\n")))

	recoverReq := &pb.RecoverBlockChaosRequest{InjectionId: 73}
	driver.recoverErr = failure
	g.Expect(server.RecoverBlockChaos(ctx, recoverReq)).Error().To(MatchError(failure))
	driver.recoverErr = nil
	_, err := server.RecoverBlockChaos(ctx, recoverReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(driver.recoveries).To(Equal([]int{73, 73}))
	g.Expect(driver.closes).To(Equal(4))
}
