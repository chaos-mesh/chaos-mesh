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

package chaosdaemon

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
)

type fakeStressorExecutor struct {
	cpu             *bpm.Process
	memory          *bpm.Process
	cpuErr          error
	memoryErr       error
	cleanupErr      error
	memoryCalls     int
	runningCPU      int
	cleanupRequests []*pb.CancelStressRequest
	checkCleanup    func(context.Context)
}

func (f *fakeStressorExecutor) ExecCPUStressors(context.Context, *pb.ExecStressRequest) (*bpm.Process, error) {
	if f.cpu != nil && f.cpuErr == nil {
		f.runningCPU++
	}
	return f.cpu, f.cpuErr
}

func (f *fakeStressorExecutor) ExecMemoryStressors(context.Context, *pb.ExecStressRequest) (*bpm.Process, error) {
	f.memoryCalls++
	return f.memory, f.memoryErr
}

func (f *fakeStressorExecutor) CancelStressors(ctx context.Context, req *pb.CancelStressRequest) (*empty.Empty, error) {
	if f.checkCleanup != nil {
		f.checkCleanup(ctx)
	}
	f.cleanupRequests = append(f.cleanupRequests, req)
	if f.cleanupErr == nil && f.cpu != nil && req.CpuInstanceUid == f.cpu.Uid {
		f.runningCPU--
	}
	return &empty.Empty{}, f.cleanupErr
}

func TestExecStressorsRollsBackCPUOnMemoryFailure(t *testing.T) {
	startupErr := errors.New("memory startup failed")
	executor := &fakeStressorExecutor{
		cpu:       &bpm.Process{Uid: "cpu-uid", Pair: bpm.ProcessPair{Pid: 123, CreateTime: 456}},
		memoryErr: startupErr,
	}
	for attempt := 0; attempt < 3; attempt++ {
		resp, err := execStressors(context.Background(), &pb.ExecStressRequest{}, executor)
		if resp != nil || !errors.Is(err, startupErr) {
			t.Fatalf("unexpected response: %+v, error: %v", resp, err)
		}
		if executor.runningCPU != 0 {
			t.Fatalf("attempt %d left %d CPU workers", attempt, executor.runningCPU)
		}
	}
	if len(executor.cleanupRequests) != 3 {
		t.Fatalf("expected cleanup after each attempt, got %d", len(executor.cleanupRequests))
	}
	for _, req := range executor.cleanupRequests {
		if req.CpuInstanceUid != "cpu-uid" || req.MemoryInstanceUid != "" || req.CpuInstance != "" {
			t.Fatalf("cleanup did not use the CPU process UID: %+v", req)
		}
	}
}

func TestExecStressorsRollbackSurvivesCanceledRequest(t *testing.T) {
	type contextKey struct{}
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request"))
	cancel()
	checked := false
	executor := &fakeStressorExecutor{
		cpu:       &bpm.Process{Uid: "cpu-uid"},
		memoryErr: context.Canceled,
		checkCleanup: func(cleanupCtx context.Context) {
			checked = true
			if cleanupCtx.Err() != nil || cleanupCtx.Value(contextKey{}) != "request" {
				t.Fatalf("cleanup inherited cancellation or lost request context: %v", cleanupCtx.Err())
			}
			deadline, ok := cleanupCtx.Deadline()
			if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > stressRollbackTimeout {
				t.Fatalf("cleanup has no bounded independent deadline: %v", deadline)
			}
		},
	}
	_, err := execStressors(ctx, &pb.ExecStressRequest{}, executor)
	if !checked || !errors.Is(err, context.Canceled) || executor.runningCPU != 0 {
		t.Fatalf("cleanup failed: checked=%v err=%v running=%d", checked, err, executor.runningCPU)
	}
}

func TestExecStressorsReportsRollbackFailure(t *testing.T) {
	startupErr := errors.New("memory startup failed")
	executor := &fakeStressorExecutor{
		cpu:        &bpm.Process{Uid: "cpu-uid"},
		memoryErr:  startupErr,
		cleanupErr: errors.New("stop timed out"),
	}
	resp, err := execStressors(context.Background(), &pb.ExecStressRequest{}, executor)
	if resp != nil || !errors.Is(err, startupErr) || !strings.Contains(err.Error(), "cpu-uid") || !strings.Contains(err.Error(), "stop timed out") {
		t.Fatalf("startup or cleanup error lost: response=%+v err=%v", resp, err)
	}
}

func TestExecStressorsDoesNotCancelUnstartedWorkers(t *testing.T) {
	startupErr := errors.New("startup failed")
	for _, tc := range []struct {
		name        string
		executor    fakeStressorExecutor
		memoryCalls int
	}{
		{name: "CPU failed", executor: fakeStressorExecutor{cpuErr: startupErr}},
		{name: "memory only failed", executor: fakeStressorExecutor{memoryErr: startupErr}, memoryCalls: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := execStressors(context.Background(), &pb.ExecStressRequest{}, &tc.executor)
			if resp != nil || !errors.Is(err, startupErr) || len(tc.executor.cleanupRequests) != 0 || tc.executor.memoryCalls != tc.memoryCalls {
				t.Fatalf("unexpected startup/cleanup: response=%+v err=%v executor=%+v", resp, err, tc.executor)
			}
		})
	}
}

func TestExecStressorsSuccessfulResponsePreservesHandles(t *testing.T) {
	for _, tc := range []struct {
		name   string
		cpu    *bpm.Process
		memory *bpm.Process
	}{
		{name: "both", cpu: &bpm.Process{Uid: "cpu", Pair: bpm.ProcessPair{Pid: 12, CreateTime: 34}}, memory: &bpm.Process{Uid: "memory", Pair: bpm.ProcessPair{Pid: 56, CreateTime: 78}}},
		{name: "CPU only", cpu: &bpm.Process{Uid: "cpu", Pair: bpm.ProcessPair{Pid: 12, CreateTime: 34}}},
		{name: "memory only", memory: &bpm.Process{Uid: "memory", Pair: bpm.ProcessPair{Pid: 56, CreateTime: 78}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			executor := &fakeStressorExecutor{cpu: tc.cpu, memory: tc.memory}
			resp, err := execStressors(context.Background(), &pb.ExecStressRequest{}, executor)
			if err != nil || resp == nil || len(executor.cleanupRequests) != 0 {
				t.Fatalf("successful startup rolled back: response=%+v err=%v", resp, err)
			}
			if tc.cpu != nil && (resp.CpuInstance != "12" || resp.CpuStartTime != 34 || resp.CpuInstanceUid != "cpu") {
				t.Fatalf("CPU handle changed: %+v", resp)
			}
			if tc.memory != nil && (resp.MemoryInstance != "56" || resp.MemoryStartTime != 78 || resp.MemoryInstanceUid != "memory") {
				t.Fatalf("memory handle changed: %+v", resp)
			}
		})
	}
}
