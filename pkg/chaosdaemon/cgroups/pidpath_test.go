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

package cgroups

import (
	"strings"
	"testing"
)

func TestParseCgroupFromReader_Valid(t *testing.T) {
	input := "0::/user.slice/user-1001.slice/session-1.scope\n"
	r := strings.NewReader(input)
	got, err := parseCgroupFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/user.slice/user-1001.slice/session-1.scope" {
		t.Errorf("expected /user.slice/user-1001.slice/session-1.scope, got %q", got)
	}
}

func TestParseCgroupFromReader_MultipleLines(t *testing.T) {
	input := "11:blkio:/system.slice\n10:memory:/system.slice\n0::/system.slice/test.scope\n"
	r := strings.NewReader(input)
	got, err := parseCgroupFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/system.slice/test.scope" {
		t.Errorf("expected /system.slice/test.scope, got %q", got)
	}
}

func TestParseCgroupFromReader_RootPath(t *testing.T) {
	input := "0::/\n"
	r := strings.NewReader(input)
	got, err := parseCgroupFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/" {
		t.Errorf("expected /, got %q", got)
	}
}

func TestParseCgroupFromReader_NoCgroupV2Line(t *testing.T) {
	input := "11:blkio:/system.slice\n10:memory:/system.slice\n"
	r := strings.NewReader(input)
	_, err := parseCgroupFromReader(r)
	if err == nil {
		t.Error("expected error for missing cgroup v2 line, got nil")
	}
}

func TestParseCgroupFromReader_EmptyInput(t *testing.T) {
	r := strings.NewReader("")
	_, err := parseCgroupFromReader(r)
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestParseCgroupFromReader_InvalidEntry(t *testing.T) {
	input := "invalidline\n"
	r := strings.NewReader(input)
	_, err := parseCgroupFromReader(r)
	if err == nil {
		t.Error("expected error for invalid entry, got nil")
	}
}

func TestParseCgroupFromReader_OnlyTwoParts(t *testing.T) {
	input := "0:/only-two-parts\n"
	r := strings.NewReader(input)
	_, err := parseCgroupFromReader(r)
	if err == nil {
		t.Error("expected error for entry with only 2 parts, got nil")
	}
}
