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
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/containerd/cgroups/v3/cgroup1"
)

func writeCgroupFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestPidPathControllerMembership(t *testing.T) {
	for _, tc := range []struct {
		name, entries, systemd string
	}{
		{"named fallback", "1:name=systemd:/host/session\n", "/host/session"},
		{"direct precedence", "1:name=systemd:/host/session\n2:systemd:/host/direct\n", "/host/direct"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "cgroup")
			writeCgroupFixture(t, file, "7:cpu,cpuacct:/host/cpu\n6:memory:/host/memory\n"+tc.entries)
			path := pidPathFromFile(file)
			for controller, want := range map[cgroup1.Name]string{
				cgroup1.Cpu: "/host/cpu", cgroup1.Cpuacct: "/host/cpu",
				cgroup1.Memory: "/host/memory", "systemd": tc.systemd,
			} {
				got, err := path(controller)
				if err != nil || got != want {
					t.Errorf("%s path = %q, %v; want %q", controller, got, err, want)
				}
			}
			if got, err := path(cgroup1.Blkio); err == nil || got != "" {
				t.Errorf("unsupported controller path = %q, %v", got, err)
			}
		})
	}
}

func TestPidPathCapturesProcErrors(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		name := "missing"
		if malformed {
			name = "malformed"
		}
		t.Run(name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "cgroup")
			if malformed {
				writeCgroupFixture(t, file, "invalid cgroup entry\n")
			}
			path := pidPathFromFile(file)
			writeCgroupFixture(t, file, "7:cpu:/replaced\n6:memory:/replaced\n")
			for _, controller := range []cgroup1.Name{cgroup1.Cpu, cgroup1.Memory} {
				got, err := path(controller)
				if err == nil || got != "" {
					t.Fatalf("%s path = %q, %v", controller, got, err)
				}
				if !strings.Contains(err.Error(), "parse cgroup file "+file) {
					t.Errorf("missing file context: %v", err)
				}
				if !malformed && !errors.Is(err, os.ErrNotExist) {
					t.Errorf("lost missing-file cause: %v", err)
				}
			}
		})
	}
}

func TestV1HierarchyUsesExistingHostControllers(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"cpu", "memory", "systemd", "unrelated"} {
		if err := os.Mkdir(filepath.Join(root, name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	subsystems, err := v1Hierarchy(root)
	if err != nil {
		t.Fatal(err)
	}
	got := map[cgroup1.Name]string{}
	for _, subsystem := range subsystems {
		p, ok := subsystem.(interface{ Path(string) string })
		if !ok {
			t.Fatalf("%s has no filesystem path", subsystem.Name())
		}
		got[subsystem.Name()] = p.Path("/work")
	}
	want := map[cgroup1.Name]string{
		cgroup1.Cpu:    filepath.Join(root, "cpu/work"),
		cgroup1.Memory: filepath.Join(root, "memory/work"),
		"systemd":      filepath.Join(root, "systemd/work"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("hierarchy = %v; want %v", got, want)
	}
}
