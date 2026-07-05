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

package pidfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestNew_CreatesFileWithCurrentPID verifies that New writes the current
// process PID to the specified path.
func TestNew_CreatesFileWithCurrentPID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pid")

	pf, err := New(path)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	defer pf.Remove()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read PID file: %v", err)
	}

	got, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		t.Fatalf("PID file content is not a valid integer: %q", string(content))
	}

	if got != os.Getpid() {
		t.Errorf("expected PID %d, got %d", os.Getpid(), got)
	}
}

// TestNew_CreatesMissingDirectories verifies that New creates parent
// directories that do not yet exist.
func TestNew_CreatesMissingDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "test.pid")

	pf, err := New(path)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}
	defer pf.Remove()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected PID file to exist at %s", path)
	}
}

// TestNew_OverwritesStaleFile verifies that New succeeds when a PID file
// already exists but references a non-existent (dead) process.
func TestNew_OverwritesStaleFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stale.pid")

	// Write a PID that is guaranteed not to correspond to any running process.
	// PID 0 is never valid for a user process; os.Stat("/proc/0") will fail on Linux.
	// On macOS /proc does not exist, so processExists always returns false.
	stalePID := 0
	if err := os.WriteFile(path, []byte(fmt.Sprintf("%d", stalePID)), 0644); err != nil {
		t.Fatalf("failed to write stale PID file: %v", err)
	}

	pf, err := New(path)
	if err != nil {
		t.Fatalf("New() returned unexpected error for stale PID file: %v", err)
	}
	defer pf.Remove()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read PID file after overwrite: %v", err)
	}

	got, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		t.Fatalf("PID file content is not a valid integer: %q", string(content))
	}

	if got != os.Getpid() {
		t.Errorf("expected PID %d after overwrite, got %d", os.Getpid(), got)
	}
}

// TestNew_RejectsRunningProcess verifies that New returns an error when a PID
// file already exists and references a still-running process.
func TestNew_RejectsRunningProcess(t *testing.T) {
	// processExists checks /proc/<pid>; on systems where /proc is mounted
	// (Linux), writing the current PID makes the check return true.
	// On macOS /proc does not exist so processExists always returns false —
	// skip on non-Linux systems.
	if _, err := os.Stat("/proc"); os.IsNotExist(err) {
		t.Skip("skipping: /proc not available on this platform")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "running.pid")

	// Write the current process's PID so processExists returns true.
	if err := os.WriteFile(path, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		t.Fatalf("failed to write running PID file: %v", err)
	}

	_, err := New(path)
	if err == nil {
		t.Fatal("New() expected error for running process PID file, got nil")
	}
}

// TestNew_IgnoresUnreadableFile verifies that New proceeds normally when the
// existing PID file cannot be read (e.g. permission denied or non-integer
// content) rather than returning an error.
func TestNew_InvalidContentInExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.pid")

	// Write non-integer content; strconv.Atoi will fail so the check is skipped.
	if err := os.WriteFile(path, []byte("not-a-pid"), 0644); err != nil {
		t.Fatalf("failed to write invalid PID file: %v", err)
	}

	pf, err := New(path)
	if err != nil {
		t.Fatalf("New() returned unexpected error for invalid PID file content: %v", err)
	}
	defer pf.Remove()
}

// TestRemove_DeletesFile verifies that Remove deletes the PID file from disk.
func TestRemove_DeletesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "remove.pid")

	pf, err := New(path)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	if err := pf.Remove(); err != nil {
		t.Fatalf("Remove() returned unexpected error: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected PID file to be deleted at %s", path)
	}
}

// TestRemove_NonExistentFile verifies that Remove returns an error when the
// PID file has already been deleted.
func TestRemove_NonExistentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gone.pid")

	pf := &PIDFile{path: path}
	if err := pf.Remove(); err == nil {
		t.Fatal("Remove() expected error for non-existent file, got nil")
	}
}
