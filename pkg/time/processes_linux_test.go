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

package time

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func processStatFixture(pid, parent int, comm string) []byte {
	fields := []string{"S", fmt.Sprint(parent)}
	for len(fields) < 19 {
		fields = append(fields, "0")
	}
	fields = append(fields, "1234")
	return []byte(fmt.Sprintf("%d (%s) %s", pid, comm, strings.Join(fields, " ")))
}

func TestPersistentProcessStatHandlesSpacesAndParentheses(t *testing.T) {
	stat, err := parseTimeProcessStat(10, processStatFixture(10, 2, "a worker) name"))
	if err != nil || stat.Parent != 2 || stat.Identity.StartTime != "1234" {
		t.Fatalf("stat=%+v err=%v", stat, err)
	}
	if _, err := parseTimeProcessStat(10, []byte("10 (truncated) S 2")); err == nil {
		t.Fatal("accepted truncated stat")
	}
}

func TestPersistentProcessDiscoveryPreservesAncestryAndContainerBoundary(t *testing.T) {
	root := t.TempDir()
	id := strings.Repeat("a", 64)
	cgroup := "5:memory:/kubepods/" + id + "\n4:rdma:/\n"
	add := func(pid, parent int, cg string) {
		path := filepath.Join(root, fmt.Sprint(pid))
		if err := os.Mkdir(path, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "stat"), processStatFixture(pid, parent, "worker with spaces)"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(path, "cgroup"), []byte(cg), 0600); err != nil {
			t.Fatal(err)
		}
	}
	add(100, 1, cgroup)
	add(101, 100, cgroup)
	add(102, 101, "0::/nested\n")
	add(103, 1, cgroup)
	add(104, 103, "0::/nested\n")
	add(200, 1, "0::/unrelated\n")
	selected, err := enumerateTimeProcesses(root, 100, cgroup, "containerd://"+id)
	if err != nil {
		t.Fatal(err)
	}
	var pids []int
	for _, identity := range selected {
		pids = append(pids, identity.PID)
		if identity.StartTime != "1234" {
			t.Fatal("lost selected identity")
		}
	}
	sort.Ints(pids)
	if !reflect.DeepEqual(pids, []int{101, 102, 103, 104}) {
		t.Fatalf("wrong process selection: %v", pids)
	}
	selected, err = enumerateTimeProcesses(root, 100, cgroup, "containerd://"+strings.Repeat("b", 64))
	if err != nil {
		t.Fatal(err)
	}
	pids = nil
	for _, identity := range selected {
		pids = append(pids, identity.PID)
	}
	sort.Ints(pids)
	if !reflect.DeepEqual(pids, []int{101, 102}) {
		t.Fatalf("expanded into unverified shared cgroup: %v", pids)
	}
	if isContainerCgroup("0::/\n") || isContainerCgroup("") {
		t.Fatal("accepted root cgroup")
	}
}
