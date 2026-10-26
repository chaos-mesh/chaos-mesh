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

package cgroups

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/containerd/cgroups/v3"
	"github.com/containerd/cgroups/v3/cgroup1"
)

func TestAttachCGroupV1WritesControllerMembership(t *testing.T) {
	root := t.TempDir()
	proc := filepath.Join(root, "proc-cgroup")
	writeCgroupFixture(t, proc, "7:cpu,cpuacct:/work/cpu\n6:memory:/work/memory\n")
	for _, file := range []string{"cpu/work/cpu/cgroup.procs", "memory/work/memory/cgroup.procs"} {
		writeCgroupFixture(t, filepath.Join(root, file), "")
	}
	for _, file := range []string{"cpu/work/sibling/cgroup.procs", "memory/work/sibling/cgroup.procs"} {
		writeCgroupFixture(t, filepath.Join(root, file), "unchanged")
	}
	attacher := &AttachCGroupV1{mode: cgroups.Legacy, path: pidPathFromFile(proc)}
	if err := attacher.attachProcess(4321, func() ([]cgroup1.Subsystem, error) { return v1Hierarchy(root) }); err != nil {
		t.Fatal(err)
	}
	for file, want := range map[string]string{
		"cpu/work/cpu/cgroup.procs": "4321", "memory/work/memory/cgroup.procs": "4321",
		"cpu/work/sibling/cgroup.procs": "unchanged", "memory/work/sibling/cgroup.procs": "unchanged",
	} {
		data, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Errorf("%s = %q; want %q", file, data, want)
		}
	}
}

func TestAttachCGroupV1PreservesLoadAndWriteErrors(t *testing.T) {
	for _, failure := range []string{"load", "write"} {
		t.Run(failure, func(t *testing.T) {
			root := t.TempDir()
			proc := filepath.Join(root, "proc-cgroup")
			writeCgroupFixture(t, proc, "7:cpu:/work/cpu\n6:memory:/work/memory\n")
			if err := os.Mkdir(filepath.Join(root, "cpu"), 0755); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(root, "cpu/work/cpu/cgroup.procs")
			if failure == "write" {
				if err := os.MkdirAll(target, 0755); err != nil {
					t.Fatal(err)
				}
			}
			attacher := &AttachCGroupV1{mode: cgroups.Legacy, path: pidPathFromFile(proc)}
			err := attacher.attachProcess(4321, func() ([]cgroup1.Subsystem, error) { return v1Hierarchy(root) })
			if err == nil {
				t.Fatal("expected attachment to fail")
			}
			context := "load cgroup v1 manager"
			if failure == "load" {
				if !errors.Is(err, cgroup1.ErrCgroupDeleted) {
					t.Errorf("lost deleted-cgroup cause: %v", err)
				}
			} else {
				context = "add process to cgroup"
				var pathErr *os.PathError
				if !errors.As(err, &pathErr) || pathErr.Path != target {
					t.Errorf("lost write cause for %s: %v", target, err)
				}
			}
			for _, part := range []string{context, "pid 4321", "cpu path /work/cpu", "memory path /work/memory"} {
				if !strings.Contains(err.Error(), part) {
					t.Errorf("missing %q in error: %v", part, err)
				}
			}
		})
	}
}

func fakeNsenter(t *testing.T, output string, exitCode int) func(...string) {
	t.Helper()
	dir := t.TempDir()
	script := `#!/bin/sh
printf '%s\n' "$@" > "$CGROUP_TEST_ARGS"
if [ "$CGROUP_TEST_EXIT" -ne 0 ]; then
 printf '%s' "$CGROUP_TEST_OUTPUT" >&2
else
 printf '%s' "$CGROUP_TEST_OUTPUT"
fi
exit "$CGROUP_TEST_EXIT"
`
	if err := os.WriteFile(filepath.Join(dir, "nsenter"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	argsFile := filepath.Join(dir, "args")
	t.Setenv("PATH", dir)
	t.Setenv("CGROUP_TEST_ARGS", argsFile)
	t.Setenv("CGROUP_TEST_OUTPUT", output)
	t.Setenv("CGROUP_TEST_EXIT", strconv.Itoa(exitCode))
	return func(want ...string) {
		t.Helper()
		data, err := os.ReadFile(argsFile)
		if err != nil {
			t.Fatal(err)
		}
		got := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		if !reflect.DeepEqual(got, want) {
			t.Errorf("nsenter arguments = %q; want %q", got, want)
		}
	}
}

func TestGetAttacherForPIDUsesV1InLegacyAndHybrid(t *testing.T) {
	for _, mode := range []cgroups.CGMode{cgroups.Legacy, cgroups.Hybrid} {
		t.Run(strconv.Itoa(int(mode)), func(t *testing.T) {
			t.Setenv("PATH", t.TempDir())
			attacher, err := getAttacherForPID(-1, mode)
			if err != nil {
				t.Fatal(err)
			}
			info := attacher.TargetCGroup()
			if info.CGMode != mode || info.V1Path == nil || info.V2CGroupPath != "" {
				t.Fatalf("unexpected target: %+v", info)
			}
			_, err = info.V1Path(cgroup1.Cpu)
			if !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "/proc/-1/cgroup") {
				t.Errorf("lost deferred target PID lookup: %v", err)
			}
		})
	}
}

func TestUnifiedAttacherUsesHostNamespaceForLookupAndAttach(t *testing.T) {
	assertArgs := fakeNsenter(t, "5:memory:/ignored\n0::/work/target\n", 0)
	attacher, err := getAttacherForPID(2468, cgroups.Unified)
	if err != nil {
		t.Fatal(err)
	}
	assertArgs("-C", "-t", "1", "cat", "/proc/2468/cgroup")
	info := attacher.TargetCGroup()
	if info.CGMode != cgroups.Unified || info.V1Path != nil || info.V2CGroupPath != "/work/target" {
		t.Fatalf("unexpected target: %+v", info)
	}
	if err := attacher.AttachProcess(4321); err != nil {
		t.Fatal(err)
	}
	assertArgs("-C", "-t", "1", "--", "sh", "-c", "echo 4321 >> /host-sys/fs/cgroup/work/target/cgroup.procs")
}

func TestUnifiedAttacherPreservesCommandErrors(t *testing.T) {
	t.Run("lookup", func(t *testing.T) {
		assertArgs := fakeNsenter(t, "permission denied", 7)
		attacher, err := getAttacherForPID(2468, cgroups.Unified)
		var exitErr *exec.ExitError
		if attacher != nil || !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
			t.Fatalf("attacher = %v, error = %v", attacher, err)
		}
		if !strings.Contains(err.Error(), "pid 2468") {
			t.Errorf("missing target PID: %v", err)
		}
		assertArgs("-C", "-t", "1", "cat", "/proc/2468/cgroup")
	})
	t.Run("malformed membership", func(t *testing.T) {
		fakeNsenter(t, "invalid membership\n", 0)
		if attacher, err := getAttacherForPID(2468, cgroups.Unified); err == nil || attacher != nil {
			t.Errorf("attacher = %v, error = %v; want failed lookup", attacher, err)
		}
	})
	t.Run("attach", func(t *testing.T) {
		assertArgs := fakeNsenter(t, "permission denied", 7)
		attacher := &AttachCGroupV2{mode: cgroups.Unified, path: "/work/target"}
		err := attacher.AttachProcess(4321)
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 7 {
			t.Fatalf("lost command cause: %v", err)
		}
		for _, part := range []string{"pid 4321", "/host-sys/fs/cgroup/work/target/cgroup.procs", "permission denied"} {
			if !strings.Contains(err.Error(), part) {
				t.Errorf("missing %q in error: %v", part, err)
			}
		}
		assertArgs("-C", "-t", "1", "--", "sh", "-c", "echo 4321 >> /host-sys/fs/cgroup/work/target/cgroup.procs")
	})
}
