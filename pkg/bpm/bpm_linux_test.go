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
	"strings"
	"testing"

	"github.com/go-logr/logr"
)

func TestStartProcessCleansUpWhenMetadataFails(t *testing.T) {
	// gopsutil reads process metadata from HOST_PROC on Linux. An empty
	// directory makes CreateTime fail after the child has started.
	t.Setenv("HOST_PROC", t.TempDir())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := StartBackgroundProcessManager(nil, logr.Discard())
	defer m.Shutdown(context.Background())
	cmd := DefaultProcessBuilder("sleep", "30").
		SetIdentifier("metadata-failure").
		SetContext(ctx).
		Build(ctx)
	defer func() {
		// Reap the child even when the regression makes this test fail.
		if cmd.Process != nil && cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()

	proc, err := m.StartProcess(ctx, cmd)
	if err == nil || !strings.Contains(err.Error(), "get process create time") {
		t.Fatalf("expected a process metadata error, got process %v and error %v", proc, err)
	}
	if cmd.ProcessState == nil {
		t.Fatal("startup failure left the child running or unreaped")
	}
	if identifiers := m.GetIdentifiers(); len(identifiers) != 0 {
		t.Fatalf("startup failure retained identifiers: %v", identifiers)
	}
}
