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
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/go-logr/logr"

	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
)

type simulatedTimeProcesses struct {
	identities map[int]string
	offsets    map[int]Config
	children   []uint32
	failPID    int
	writes     int
}

func (f *simulatedTimeProcesses) manager(directory string) *PersistentTimeChaos {
	p := NewPersistentTimeChaos(directory, logr.Discard())
	p.bootID = func() (string, error) { return "boot-one", nil }
	p.identity = func(pid int) (processIdentity, error) {
		start, ok := f.identities[pid]
		if !ok {
			return processIdentity{}, os.ErrNotExist
		}
		return processIdentity{pid, start}, nil
	}
	p.children = func(int, string, string) ([]processIdentity, error) {
		var result []processIdentity
		for _, pid := range f.children {
			result = append(result, processIdentity{int(pid), f.identities[int(pid)]})
		}
		return result, nil
	}
	p.cgroup = func(int) (string, error) { return "0::/pod/container", nil }
	p.execute = func(process *processRecovery, c *Config, _ []processRecovery, save func() error) error {
		if len(process.ClockGetTime.OriginalCode) == 0 {
			process.ClockGetTime = imageRecovery{OriginalCode: bytes.Repeat([]byte{0x90}, 16), OriginalAddress: 0x1000, Content: []byte{1, 2, 3}, FakeEntry: &mapreader.Entry{StartAddress: 0x2000, EndAddress: 0x2100}}
		}
		// The simulated injector follows the same required write-before-patch
		// contract as FakeImage. The tests recreate managers from these files.
		if err := save(); err != nil {
			return err
		}
		f.writes++
		if process.Identity.PID == f.failPID {
			return errors.New("simulated write failure")
		}
		if c == nil {
			delete(f.offsets, process.Identity.PID)
		} else {
			f.offsets[process.Identity.PID] = *c
		}
		return nil
	}
	return p
}

func newSimulatedTimeProcesses() *simulatedTimeProcesses {
	return &simulatedTimeProcesses{identities: map[int]string{100: "1000", 101: "1001"}, offsets: make(map[int]Config), children: []uint32{101}}
}

func TestPersistentTimeChaosRestartPreservesOverlappingTasks(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	first := NewConfig(10, 5, 1)
	second := NewConfig(20, 7, 2)
	if err := processes.manager(directory).Apply("first", "pod:app", "container", 100, first); err != nil {
		t.Fatal(err)
	}
	// Every RPC uses a fresh manager, just like replacing the daemon.
	if err := processes.manager(directory).Apply("second", "pod:app", "container", 100, second); err != nil {
		t.Fatal(err)
	}
	if got := processes.offsets[100]; got != NewConfig(30, 12, 3) {
		t.Fatalf("aggregate after restart = %+v", got)
	}
	if err := processes.manager(directory).Recover("first", "pod:app", "container", 100); err != nil {
		t.Fatal(err)
	}
	if got := processes.offsets[100]; got != second {
		t.Fatalf("remaining offset = %+v", got)
	}
	if err := processes.manager(directory).Recover("second", "pod:app", "container", 100); err != nil {
		t.Fatal(err)
	}
	if len(processes.offsets) != 0 {
		t.Fatalf("unrecovered processes: %+v", processes.offsets)
	}
	// Lost successful replies must be safe to retry even after another restart.
	if err := processes.manager(directory).Recover("second", "pod:app", "container", 100); err != nil {
		t.Fatal(err)
	}
}

func TestPersistentTimeChaosApplyRetryDoesNotDoubleOffset(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	config := NewConfig(10, 5, 1)
	for i := 0; i < 2; i++ {
		if err := processes.manager(directory).Apply("task", "pod:app", "container", 100, config); err != nil {
			t.Fatal(err)
		}
	}
	if got := processes.offsets[100]; got != config {
		t.Fatalf("retried offset = %+v", got)
	}
}

func TestPersistentTimeChaosRecoveryRetriesFailedChildAfterRestart(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	if err := processes.manager(directory).Apply("task", "pod:app", "container", 100, NewConfig(5, 0, 1)); err != nil {
		t.Fatal(err)
	}
	// The parent exits and the child is reparented. Its durable record remains.
	delete(processes.identities, 100)
	processes.children = nil
	processes.failPID = 101
	if err := processes.manager(directory).Recover("task", "pod:app", "container", 100); err == nil {
		t.Fatal("expected child recovery failure")
	}
	if _, ok := processes.offsets[101]; !ok {
		t.Fatal("failed child unexpectedly recovered")
	}
	processes.failPID = 0
	if err := processes.manager(directory).Recover("task", "pod:app", "container", 100); err != nil {
		t.Fatal(err)
	}
	if _, ok := processes.offsets[101]; ok {
		t.Fatal("child offset survived retry")
	}
}

func TestPersistentTimeChaosFailedApplyRetainsRecoveryState(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	processes.failPID = 101
	if err := processes.manager(directory).Apply("task", "pod:app", "container", 100, NewConfig(5, 0, 1)); err == nil {
		t.Fatal("expected apply failure")
	}
	processes.failPID = 0
	if err := processes.manager(directory).Recover("task", "pod:app", "container", 100); err != nil {
		t.Fatal(err)
	}
	if len(processes.offsets) != 0 {
		t.Fatalf("unrecovered state: %+v", processes.offsets)
	}
}

func TestPersistentTimeChaosPIDReuseDoesNotRestoreOldProcess(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	if err := processes.manager(directory).Apply("task", "pod:app", "container", 100, NewConfig(5, 0, 1)); err != nil {
		t.Fatal(err)
	}
	processes.identities[101] = "new-start"
	processes.children = nil
	p := processes.manager(directory)
	execute := p.execute
	p.execute = func(process *processRecovery, c *Config, templates []processRecovery, save func() error) error {
		if process.Identity.PID == 101 {
			t.Fatal("attempted to restore a reused PID")
		}
		return execute(process, c, templates, save)
	}
	if err := p.Recover("task", "pod:app", "container", 100); err != nil {
		t.Fatal(err)
	}
}

func TestPersistentTimeChaosRequiresJournalForRecovery(t *testing.T) {
	processes := newSimulatedTimeProcesses()
	p := processes.manager(t.TempDir())
	if err := p.Recover("task", "pod:app", "container", 100); err == nil || !strings.Contains(err.Error(), "journal not found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if processes.writes != 0 {
		t.Fatal("wrote to a process without recovery state")
	}
}

func TestPersistentTimeChaosRejectsCorruptJournalBeforeWrites(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	p := processes.manager(directory)
	if err := p.Apply("task", "pod:app", "container", 100, NewConfig(5, 0, 1)); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("journal files: %v %v", files, err)
	}
	if err := os.WriteFile(files[0], []byte("{truncated"), 0600); err != nil {
		t.Fatal(err)
	}
	writes := processes.writes
	if err := processes.manager(directory).Recover("task", "pod:app", "container", 100); err == nil {
		t.Fatal("accepted corrupt journal")
	}
	if processes.writes != writes {
		t.Fatal("wrote to processes despite corrupt journal")
	}
}

func TestPersistentTimeChaosJournalsAreContainerScoped(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	for _, container := range []string{"container-a", "container-b"} {
		if err := processes.manager(directory).Apply("same-task", "pod:"+container, container, 100, NewConfig(5, 0, 1)); err != nil {
			t.Fatal(err)
		}
	}
	files, _ := filepath.Glob(filepath.Join(directory, "*.json"))
	if len(files) != 2 {
		t.Fatalf("expected independent journals, got %v", files)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var journal timeJournal
	if err := json.Unmarshal(data, &journal); err != nil {
		t.Fatal(err)
	}
	if len(journal.Processes) != 2 || len(journal.Tasks) != 1 {
		t.Fatalf("incomplete journal: %+v", journal)
	}
}

type simulatedImageProgram struct {
	original []byte
	jumped   bool
	jumpErr  error
}

func (p *simulatedImageProgram) MmapSlice([]byte) (*mapreader.Entry, error) {
	return &mapreader.Entry{StartAddress: 0x2000, EndAddress: 0x2100}, nil
}
func (p *simulatedImageProgram) FindSymbolInEntry(string, *mapreader.Entry) (uint64, uint64, error) {
	return 0x1000, uint64(len(p.original)), nil
}
func (p *simulatedImageProgram) ReadSlice(_ uint64, size uint64) (*[]byte, error) {
	if size != 16 {
		return nil, errors.New("must read exactly overwritten bytes")
	}
	copied := append([]byte(nil), p.original...)
	return &copied, nil
}
func (p *simulatedImageProgram) PtraceWriteSlice(_ uint64, data []byte) error {
	p.original = append([]byte(nil), data...)
	return nil
}
func (p *simulatedImageProgram) JumpToFakeFunc(uint64, uint64) error {
	p.jumped = true
	return p.jumpErr
}

func TestPersistentFakeImageSavesOriginalBeforeJump(t *testing.T) {
	program := &simulatedImageProgram{original: bytes.Repeat([]byte{0x90}, 16)}
	image := NewFakeImage("clock_gettime", []byte{1, 2, 3}, nil, logr.Discard())
	saved := false
	image.saveRecovery = func() error {
		if program.jumped {
			t.Fatal("jump preceded durable recovery record")
		}
		if !reflect.DeepEqual(image.OriginFuncCode, program.original) || image.OriginAddress != 0x1000 || image.fakeEntry.StartAddress != 0x2000 {
			t.Fatal("incomplete recovery snapshot")
		}
		saved = true
		return nil
	}
	if _, err := image.InjectFakeImage(program, &mapreader.Entry{}); err != nil {
		t.Fatal(err)
	}
	if !saved || !program.jumped {
		t.Fatal("injection was not completed")
	}
}

func TestPersistentFakeImageJournalFailurePreventsJump(t *testing.T) {
	program := &simulatedImageProgram{original: bytes.Repeat([]byte{0x90}, 16)}
	image := NewFakeImage("clock_gettime", []byte{1, 2, 3}, nil, logr.Discard())
	sentinel := errors.New("disk full")
	image.saveRecovery = func() error { return sentinel }
	if _, err := image.InjectFakeImage(program, &mapreader.Entry{}); !errors.Is(err, sentinel) {
		t.Fatalf("unexpected error %v", err)
	}
	if program.jumped {
		t.Fatal("modified vDSO without durable originals")
	}
}

func TestPersistentFakeImageRejectsUntrackedHook(t *testing.T) {
	program := &simulatedImageProgram{original: timeJump(0x3000)}
	image := NewFakeImage("clock_gettime", []byte{1, 2, 3}, nil, logr.Discard())
	if _, err := image.InjectFakeImage(program, &mapreader.Entry{}); err == nil {
		t.Fatal("accepted an untracked original hook")
	}
	if program.jumped {
		t.Fatal("overwrote an untracked hook")
	}
}

func TestPersistentFakeImageRecognizesInterruptedPatchAndRestore(t *testing.T) {
	original := bytes.Repeat([]byte{0x90}, 16)
	jump := timeJump(0x2000)
	for _, current := range [][]byte{original, jump, append(append([]byte{}, original[:8]...), jump[8:]...), append(append([]byte{}, jump[:8]...), original[8:]...)} {
		if !matchesTimePatch(current, original, jump) {
			t.Fatalf("rejected recoverable bytes %x", current)
		}
	}
	foreign := append([]byte(nil), jump...)
	foreign[15] ^= 1
	if matchesTimePatch(foreign, original, jump) {
		t.Fatal("accepted unrelated vDSO code")
	}
}

func TestPersistentTimeChaosRecoverAfterContainerRestart(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	if err := processes.manager(directory).Apply("task", "pod:app", "old-container", 100, NewConfig(5, 0, 1)); err != nil {
		t.Fatal(err)
	}
	processes.identities[200] = "replacement"
	p := processes.manager(directory)
	execute := p.execute
	p.execute = func(process *processRecovery, c *Config, templates []processRecovery, save func() error) error {
		if process.Identity.PID == 200 {
			t.Fatal("touched replacement container")
		}
		if c != nil {
			t.Fatal("injected during old-container recovery")
		}
		return execute(process, c, templates, save)
	}
	if err := p.Recover("task", "pod:app", "new-container", 200); err != nil {
		t.Fatal(err)
	}
	if len(processes.offsets) != 0 {
		t.Fatalf("old container offset remains: %+v", processes.offsets)
	}
}

func TestPersistentTimeChaosFailedApplyDoesNotContaminateNextTask(t *testing.T) {
	directory := t.TempDir()
	processes := newSimulatedTimeProcesses()
	processes.failPID = 101
	if err := processes.manager(directory).Apply("failed", "pod:app", "container", 100, NewConfig(10, 0, 1)); err == nil {
		t.Fatal("expected failure")
	}
	processes.failPID = 0
	expected := NewConfig(20, 0, 1)
	if err := processes.manager(directory).Apply("next", "pod:app", "container", 100, expected); err != nil {
		t.Fatal(err)
	}
	if got := processes.offsets[100]; got != expected {
		t.Fatalf("failed task contributed offset: %+v", got)
	}
}

func TestPersistentTimeChaosRejectsReusedDiscoveredPID(t *testing.T) {
	processes := newSimulatedTimeProcesses()
	p := processes.manager(t.TempDir())
	p.children = func(int, string, string) ([]processIdentity, error) {
		return []processIdentity{{101, "old-start"}}, nil
	}
	execute := p.execute
	p.execute = func(process *processRecovery, c *Config, templates []processRecovery, save func() error) error {
		if process.Identity.PID == 101 {
			t.Fatal("injected PID reused after discovery")
		}
		return execute(process, c, templates, save)
	}
	if err := p.Apply("task", "pod:app", "container", 100, NewConfig(5, 0, 1)); err != nil {
		t.Fatal(err)
	}
}

func TestPersistentTimeChaosRecoversKnownProcessesWhenDiscoveryFails(t *testing.T) {
	processes := newSimulatedTimeProcesses()
	directory := t.TempDir()
	if err := processes.manager(directory).Apply("task", "pod:app", "container", 100, NewConfig(5, 0, 1)); err != nil {
		t.Fatal(err)
	}
	p := processes.manager(directory)
	p.children = func(int, string, string) ([]processIdentity, error) { return nil, errors.New("proc unavailable") }
	if err := p.Recover("task", "pod:app", "container", 100); err == nil {
		t.Fatal("expected discovery failure")
	}
	if len(processes.offsets) != 0 {
		t.Fatalf("known processes were not recovered: %+v", processes.offsets)
	}
}

type simulatedRecoveryProgram struct {
	simulatedImageProgram
	memory  map[uint64][]byte
	readErr error
}

func (p *simulatedRecoveryProgram) ReadSlice(address, size uint64) (*[]byte, error) {
	if p.readErr != nil {
		return nil, p.readErr
	}
	data, ok := p.memory[address]
	if !ok || uint64(len(data)) < size {
		return nil, errors.New("unmapped simulated read")
	}
	copied := append([]byte(nil), data[:size]...)
	return &copied, nil
}

func TestPersistentFakeImageRecognizesSavedPayloadAcrossDaemonBuilds(t *testing.T) {
	original := bytes.Repeat([]byte{0x90}, 16)
	saved := imageRecovery{OriginalCode: original, OriginalAddress: 0x1000, FakeEntry: &mapreader.Entry{StartAddress: 0x2000, EndAddress: 0x2010}, Content: []byte{1, 2, 3}}
	entries := []mapreader.Entry{{StartAddress: 0x1000, EndAddress: 0x1100, Path: vdsoEntryName}, {StartAddress: 0x2000, EndAddress: 0x3000}}
	p := &simulatedRecoveryProgram{memory: map[uint64][]byte{0x1000: timeJump(0x2000), 0x2000: {1, 2, 3}}}
	image := NewFakeImage("clock_gettime", []byte{9, 9, 9}, nil, logr.Discard())
	found, err := image.matchesSavedImage(p, entries, saved)
	if err != nil || !found {
		t.Fatalf("failed to recognize injecting build: found=%v err=%v", found, err)
	}
	restoreImageRecovery(image, saved)
	if !bytes.Equal(image.content, saved.Content) {
		t.Fatal("did not preserve injecting payload on adoption")
	}
	p.memory[0x2000] = []byte{9, 9, 9}
	if _, err := image.matchesSavedImage(p, entries, saved); err == nil {
		t.Fatal("reported active damaged image as recovered")
	}
	if _, err := image.matchesSavedImage(p, entries[:1], saved); err == nil {
		t.Fatal("reported active unmapped image as recovered")
	}
	p.memory[0x1000] = original
	if found, err := image.matchesSavedImage(p, entries[:1], saved); found || err != nil {
		t.Fatalf("clean exec should not restore stale addresses: %v %v", found, err)
	}
	p.readErr = errors.New("permission denied")
	if _, err := image.matchesSavedImage(p, entries, saved); err == nil {
		t.Fatal("swallowed process read failure")
	}
}

func TestPersistentTimeChaosRejectsIncompleteProcessRecords(t *testing.T) {
	for _, process := range []*processRecovery{nil, {Identity: processIdentity{PID: 100, StartTime: "bad"}, ContainerID: "container"}, {Identity: processIdentity{PID: 100, StartTime: "1000"}, ContainerID: "container", ClockGetTime: imageRecovery{OriginalCode: []byte{1}}}} {
		key := "100:1000"
		if process != nil {
			key = processIdentityKey(process.Identity)
		}
		if err := validateTimeJournal(&timeJournal{Processes: map[string]*processRecovery{key: process}}); err == nil {
			t.Fatalf("accepted malformed process %+v", process)
		}
	}
}

func TestPersistentFakeImageTriesAllInheritedCandidates(t *testing.T) {
	original := bytes.Repeat([]byte{0x90}, 16)
	first := imageRecovery{OriginalCode: original, OriginalAddress: 0x1000, FakeEntry: &mapreader.Entry{StartAddress: 0x2000, EndAddress: 0x2010}, Content: []byte{1, 2, 3}}
	second := imageRecovery{OriginalCode: original, OriginalAddress: 0x1000, FakeEntry: &mapreader.Entry{StartAddress: 0x3000, EndAddress: 0x3010}, Content: []byte{4, 5, 6}}
	entries := []mapreader.Entry{{StartAddress: 0x1000, EndAddress: 0x1100, Path: vdsoEntryName}, {StartAddress: 0x2000, EndAddress: 0x4000}}
	program := &simulatedRecoveryProgram{memory: map[uint64][]byte{0x1000: timeJump(0x3000), 0x2000: {9, 9, 9}, 0x3000: {4, 5, 6}}}
	image := NewFakeImage("clock_gettime", []byte{0}, nil, logr.Discard())
	image.recoveryCandidates = []imageRecovery{first, second}
	entry, err := image.findSavedImage(program, entries)
	if err != nil || entry == nil || entry.StartAddress != 0x3000 {
		t.Fatalf("failed to try matching later candidate: entry=%+v err=%v", entry, err)
	}
	image = NewFakeImage("clock_gettime", []byte{0}, nil, logr.Discard())
	image.recoveryCandidates = []imageRecovery{first}
	if _, err := image.findSavedImage(program, entries); err == nil {
		t.Fatal("discarded mismatch error when no candidate matched")
	}
}
