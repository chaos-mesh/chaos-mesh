// Copyright Chaos Mesh Authors.
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

//go:build linux

package chaosdaemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/cgroups"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("Daemon process lifecycle", func() {
	var server *DaemonServer
	var runtimeClient *httpChaosRuntimeClientStub

	BeforeEach(func() {
		runtimeClient = &httpChaosRuntimeClientStub{containerPID: 42}
		server = NewDaemonServerWithCRClient(runtimeClient, nil, logr.Discard())
		DeferCleanup(func() { server.backgroundProcessManager.Shutdown(context.Background()) })
	})

	It("checks the runtime before launching either type of stressor", func(ctx SpecContext) {
		response, err := server.ExecStressors(ctx, &pb.ExecStressRequest{})
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&pb.ExecStressResponse{}))
		Expect(runtimeClient.containerCalls).To(BeZero())

		runtimeClient.containerErr = errors.New("runtime unavailable")
		for _, req := range []*pb.ExecStressRequest{
			{Target: "fixture", CpuStressors: "--cpu 1", MemoryStressors: "--vm 1"},
			{Target: "fixture", MemoryStressors: "--vm 1"},
		} {
			response, err := server.ExecStressors(ctx, req)
			Expect(err).To(MatchError(runtimeClient.containerErr))
			Expect(response).To(BeNil())
		}
		Expect(runtimeClient.containerCalls).To(Equal(2))
	})

	It("reaps a CPU stressor after a memory launch failure and reports only the retried processes", func(ctx SpecContext) {
		failure := errors.New("fixture attach failure")
		var cpuPID, memoryPID, workerPID int
		workerFile := filepath.Join(GinkgoT().TempDir(), "worker-pid")
		pendingFailure := true
		resolveCount := 0
		DeferCleanup(mock.With("MockStressCGroupAttacher", func(target int) (cgroups.AttachCGroup, error) {
			Expect(target).To(Equal(42))
			resolveCount++
			if resolveCount%2 == 1 {
				return stressCGroupFixture{attach: func(pid int) error {
					cpuPID = pid
					if pendingFailure {
						Eventually(ctx, func() error {
							value, err := os.ReadFile(workerFile)
							if err != nil {
								return err
							}
							workerPID, err = strconv.Atoi(string(value))
							return err
						}).Should(Succeed())
						process, err := os.FindProcess(workerPID)
						Expect(err).NotTo(HaveOccurred())
						DeferCleanup(func() {
							_ = process.Kill()
							_ = process.Release()
						})
					}
					return nil
				}}, nil
			}
			return stressCGroupFixture{attach: func(pid int) error {
				memoryPID = pid
				if pendingFailure {
					pendingFailure = false
					return failure
				}
				return nil
			}}, nil
		}))
		DeferCleanup(mock.With("MockProcessBuild", func(_ context.Context, command string, args ...string) *exec.Cmd {
			Expect(command).To(Equal("/usr/local/bin/pause"))
			Expect(args).To(Or(ContainElement("stress-ng"), ContainElement("memStress")))
			if resolveCount == 1 {
				return exec.CommandContext(ctx, "sh", "-c", `
sleep 60 &
worker=$!
trap 'kill "$worker"; wait "$worker"; exit 0' TERM
printf '%s' "$worker" > "$1"
wait "$worker"
`, "fixture-stressor", workerFile)
			}
			return exec.CommandContext(ctx, "sleep", "60")
		}))
		Expect(mock.On("MockProcessBuild")).NotTo(BeNil(), "run with failpoint instrumentation")

		req := &pb.ExecStressRequest{Target: "fixture", CpuStressors: "--cpu 1", MemoryStressors: "--size 32M", EnterNS: true, OomScoreAdj: 10}
		response, err := server.ExecStressors(ctx, req)
		Expect(response).To(BeNil())
		Expect(err).To(MatchError(failure))
		Expect(cpuPID).To(BeNumerically(">", 0))
		Expect(memoryPID).To(BeNumerically(">", 0))
		for _, pid := range []int{cpuPID, memoryPID, workerPID} {
			// Waiting for each child to be reaped also removes it from the process manager.
			_, err := os.Stat("/proc/" + strconv.Itoa(pid))
			Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
		}

		response, err = server.ExecStressors(ctx, req)
		Expect(err).NotTo(HaveOccurred())
		Expect(response.GetCpuInstance()).To(Equal(strconv.Itoa(cpuPID)))
		Expect(response.GetMemoryInstance()).To(Equal(strconv.Itoa(memoryPID)))
		Expect(response.GetCpuInstanceUid()).NotTo(BeEmpty())
		Expect(response.GetMemoryInstanceUid()).NotTo(BeEmpty())
		_, err = server.CancelStressors(ctx, &pb.CancelStressRequest{
			CpuInstanceUid: response.GetCpuInstanceUid(), MemoryInstanceUid: response.GetMemoryInstanceUid(),
		})
		Expect(err).NotTo(HaveOccurred())
	}, SpecTimeout(15*time.Second))

	It("cancels stressors by current UID and by legacy PID with start time", func(ctx SpecContext) {
		cpu, err := server.backgroundProcessManager.StartProcess(ctx, &bpm.ManagedCommand{Cmd: exec.CommandContext(ctx, "sleep", "60")})
		Expect(err).NotTo(HaveOccurred())
		memory, err := server.backgroundProcessManager.StartProcess(ctx, &bpm.ManagedCommand{Cmd: exec.CommandContext(ctx, "sleep", "60")})
		Expect(err).NotTo(HaveOccurred())
		_, err = server.CancelStressors(ctx, &pb.CancelStressRequest{
			CpuInstance: strconv.Itoa(cpu.Pair.Pid), CpuStartTime: cpu.Pair.CreateTime + 1,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(cpu.Stopped()).NotTo(BeClosed())

		_, err = server.CancelStressors(ctx, &pb.CancelStressRequest{
			CpuInstance: strconv.Itoa(cpu.Pair.Pid), CpuStartTime: cpu.Pair.CreateTime, MemoryInstanceUid: memory.UID,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(cpu.Stopped()).To(BeClosed())
		Expect(memory.Stopped()).To(BeClosed())
		Expect(server.backgroundProcessManager.GetUID(cpu.Pair)).To(BeEmpty())
		Expect(server.backgroundProcessManager.GetUID(memory.Pair)).To(BeEmpty())
	}, SpecTimeout(5*time.Second))

	It("rejects invalid legacy PIDs and reports stale explicit process identifiers", func(ctx SpecContext) {
		for _, req := range []*pb.CancelStressRequest{
			{CpuInstance: "invalid"},
			{MemoryInstance: "invalid"},
			{CpuInstanceUid: "unknown-cpu"},
			{MemoryInstanceUid: "unknown-memory"},
		} {
			response, err := server.CancelStressors(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())
		}
	})

	It("validates I/O rules before looking up a container PID", func(ctx SpecContext) {
		Expect(server.ApplyIOChaos(ctx, &pb.ApplyIOChaosRequest{Actions: "malformed"})).Error().To(MatchError(ContainSubstring("unmarshal json bytes")))
		Expect(server.ApplyIOChaos(ctx, &pb.ApplyIOChaosRequest{Actions: "[]"})).To(Equal(&pb.ApplyIOChaosResponse{}))
		Expect(runtimeClient.containerCalls).To(BeZero())

		runtimeClient.containerErr = errors.New("runtime unavailable")
		Expect(server.ApplyIOChaos(ctx, &pb.ApplyIOChaosRequest{Actions: "[{}]"})).Error().To(MatchError(runtimeClient.containerErr))
		Expect(runtimeClient.containerCalls).To(Equal(1))
	})

	for _, mode := range []string{"ready", "reject update", "not ready"} {
		It("manages the I/O helper when it is "+mode, func(ctx SpecContext) {
			executable, err := os.Executable()
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(mock.With("MockProcessBuild", func(_ context.Context, command string, args ...string) *exec.Cmd {
				Expect(command).To(Equal(todaBin))
				Expect(args).To(Equal([]string{"--path", "/fixture", "--verbose", "info"}))
				child := exec.CommandContext(ctx, executable, "-test.run=^TestIOHelperFixture$")
				child.Env = append(os.Environ(), "CHAOS_MESH_IO_TEST_MODE="+mode)
				return child
			}))
			Expect(mock.On("MockProcessBuild")).NotTo(BeNil(), "run with the repository's failpoint instrumentation")

			req := &pb.ApplyIOChaosRequest{ContainerId: "fixture", Volume: "/fixture", Actions: `[{"type":"latency","path":"/sample","percent":100,"latency":"5ms"}]`}
			if mode != "ready" {
				for range 2 {
					response, err := server.ApplyIOChaos(ctx, req)
					Expect(response).To(BeNil())
					msg := "toda update RPC failed"
					if mode == "not ready" {
						msg = "startup takes too long"
					}
					Expect(err).To(MatchError(ContainSubstring(msg)))
				}
				return
			}

			first, err := server.ApplyIOChaos(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(first.GetInstanceUid()).NotTo(BeEmpty())
			Expect(first.GetInstance()).To(BeNumerically(">", 0))
			Expect(first.GetStartTime()).To(BeNumerically(">", 0))

			// A replacement from an older controller identifies the first process by PID and start time.
			req.Instance, req.StartTime = first.Instance, first.StartTime
			second, err := server.ApplyIOChaos(ctx, req)
			Expect(err).NotTo(HaveOccurred())
			Expect(second.GetInstanceUid()).NotTo(Equal(first.GetInstanceUid()))
			Expect(server.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: int(first.Instance), CreateTime: first.StartTime})).To(BeEmpty())

			// Applying an empty rule set removes the existing helper without starting another.
			Expect(server.ApplyIOChaos(ctx, &pb.ApplyIOChaosRequest{InstanceUid: second.InstanceUid, Actions: "[]"})).
				To(Equal(&pb.ApplyIOChaosResponse{}))
			Expect(server.backgroundProcessManager.GetUID(bpm.ProcessPair{Pid: int(second.Instance), CreateTime: second.StartTime})).To(BeEmpty())
		}, SpecTimeout(10*time.Second))
	}
})

type stressCGroupFixture struct {
	cgroups.AttachCGroup
	attach func(int) error
}

func (f stressCGroupFixture) AttachProcess(pid int) error { return f.attach(pid) }

func TestIOHelperFixture(t *testing.T) {
	mode := os.Getenv("CHAOS_MESH_IO_TEST_MODE")
	if mode == "" {
		return
	}
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	for {
		var request struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := decoder.Decode(&request); err != nil {
			if errors.Is(err, io.EOF) {
				os.Exit(0)
			}
			t.Fatal(err)
		}
		body := map[string]any{"jsonrpc": "2.0", "id": request.ID}
		switch {
		case request.Method == "update" && mode != "reject update":
			var params [][]v1alpha1.IOChaosAction
			if err := json.Unmarshal(request.Params, &params); err != nil {
				t.Fatal(err)
			}
			if len(params) != 1 || len(params[0]) != 1 ||
				params[0][0].Type != "latency" || params[0][0].Path != "/sample" || params[0][0].Latency != "5ms" || params[0][0].Percent != 100 {
				t.Fatalf("unexpected I/O rules: %s", request.Params)
			}
			body["result"] = "updated"
		case request.Method == "get_status" && string(request.Params) == `["ping"]`:
			body["result"] = "ok"
			if mode == "not ready" {
				body["result"] = "starting"
			}
		default:
			body["error"] = map[string]any{"code": -32000, "message": fmt.Sprintf("rejected %s", request.Method)}
		}
		if err := encoder.Encode(body); err != nil {
			t.Fatal(err)
		}
	}
}
