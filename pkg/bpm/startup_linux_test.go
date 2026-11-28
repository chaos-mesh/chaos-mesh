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

package bpm

import (
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/go-logr/logr"
)

func TestStartProcessMetadataFailure(t *testing.T) {
	t.Setenv("HOST_PROC", t.TempDir())
	manager := StartBackgroundProcessManager(nil, logr.Discard())
	defer manager.Shutdown(context.Background())
	cmd := DefaultProcessBuilder("sleep", "60").SetIdentifier("metadata-failure").Build(context.Background())
	defer func() {
		if cmd.Process != nil && cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	proc, err := manager.StartProcess(context.Background(), cmd)
	if err == nil || !strings.Contains(err.Error(), "get process") {
		t.Fatalf("expected metadata failure, got process %v and error %v", proc, err)
	}
	if proc != nil || cmd.Process == nil {
		t.Fatalf("expected a started but unregistered child, got process %v and child %v", proc, cmd.Process)
	}
	if cmd.ProcessState == nil {
		t.Fatal("metadata failure left the child running or unreaped")
	}
	if err := cmd.Process.Signal(syscall.Signal(0)); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("child was not reaped: %v", err)
	}
	if ids := manager.GetIdentifiers(); len(ids) != 0 {
		t.Fatalf("metadata failure retained identifiers: %v", ids)
	}
}

func TestStartProcessPreservesExternallyStartedChild(t *testing.T) {
	ctx := context.Background()
	manager := StartBackgroundProcessManager(nil, logr.Discard())
	defer manager.Shutdown(ctx)
	cmd := DefaultProcessBuilder("sleep", "60").SetIdentifier("external-child").Build(ctx)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	proc, err := manager.StartProcess(ctx, cmd)
	if err == nil || proc != nil {
		t.Fatalf("expected an already-started command to be rejected, got process %v and error %v", proc, err)
	}
	if cmd.ProcessState != nil {
		t.Fatal("manager reaped an externally started child")
	}
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("manager stopped an externally started child: %v", err)
	}
	if ids := manager.GetIdentifiers(); len(ids) != 0 {
		t.Fatalf("rejected command retained identifiers: %v", ids)
	}
}
