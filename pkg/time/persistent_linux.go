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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/tasks"
	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
)

// PersistentTimeChaos keeps the desired tasks and recovery data on the host, so
// replacing the daemon does not lose the information needed to undo a vDSO patch.
// The directory must survive daemon container replacement (for example /host-run).
type PersistentTimeChaos struct {
	directory string
	logger    logr.Logger
	bootID    func() (string, error)
	identity  func(int) (processIdentity, error)
	children  func(int, string, string) ([]processIdentity, error)
	cgroup    func(int) (string, error)
	execute   func(*processRecovery, *Config, []processRecovery, func() error) error
}

type processIdentity struct {
	PID       int    `json:"pid"`
	StartTime string `json:"startTime"`
}

type imageRecovery struct {
	OriginalCode    []byte           `json:"originalCode,omitempty"`
	OriginalAddress uint64           `json:"originalAddress,omitempty"`
	FakeEntry       *mapreader.Entry `json:"fakeEntry,omitempty"`
	Content         []byte           `json:"content,omitempty"`
	Offsets         map[string]int   `json:"offsets,omitempty"`
}

type processRecovery struct {
	ContainerID  string          `json:"containerID"`
	Identity     processIdentity `json:"identity"`
	ClockGetTime imageRecovery   `json:"clockGetTime"`
	GetTimeOfDay imageRecovery   `json:"getTimeOfDay"`
}

type persistentConfig struct {
	Seconds     int64  `json:"seconds"`
	Nanoseconds int64  `json:"nanoseconds"`
	ClockIDs    uint64 `json:"clockIDs"`
}

type timeJournal struct {
	Version          int                         `json:"version"`
	BootID           string                      `json:"bootID"`
	ContainerID      string                      `json:"containerID"`
	PodContainerName string                      `json:"podContainerName"`
	Cgroup           string                      `json:"cgroup"`
	Tasks            map[string]persistentConfig `json:"tasks"`
	// Completed is retained for idempotent Recover RPC retries, including a
	// successful recovery whose response was lost before a daemon restart.
	Completed map[string]bool             `json:"completed"`
	Processes map[string]*processRecovery `json:"processes"`
}

// NewPersistentTimeChaos creates a coordinator backed by a host-mounted directory.
func NewPersistentTimeChaos(directory string, logger logr.Logger) *PersistentTimeChaos {
	p := &PersistentTimeChaos{
		directory: directory,
		logger:    logger,
		bootID: func() (string, error) {
			data, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
			return strings.TrimSpace(string(data)), err
		},
		identity: readProcessIdentity,
		children: func(pid int, cgroup, containerID string) ([]processIdentity, error) {
			return enumerateTimeProcesses("/proc", pid, cgroup, containerID)
		},
		cgroup: func(pid int) (string, error) {
			data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
			return string(data), err
		},
	}
	p.execute = p.executeProcess
	return p
}

// Apply reconciles the aggregate offset for all desired tasks on the container.
func (p *PersistentTimeChaos) Apply(uid, podContainerName, containerID string, pid int, config Config) error {
	return p.reconcile(uid, podContainerName, containerID, pid, &config)
}

// Recover removes a task and restores the remaining aggregate or original code.
func (p *PersistentTimeChaos) Recover(uid, podContainerName, containerID string, pid int) error {
	return p.reconcile(uid, podContainerName, containerID, pid, nil)
}

func (p *PersistentTimeChaos) reconcile(uid, podContainerName, containerID string, pid int, config *Config) error {
	if uid == "" || podContainerName == "" || containerID == "" || pid <= 0 {
		return errors.New("time chaos requires a task UID, container identity and positive PID")
	}
	bootID, err := p.bootID()
	if err != nil {
		return errors.Wrap(err, "read boot identity")
	}
	if bootID == "" {
		return errors.New("empty boot identity")
	}
	// Hash the complete identity; RPC fields must never become path components.
	key := fmt.Sprintf("%x", sha256.Sum256([]byte(bootID+"\x00"+podContainerName)))
	if err := os.MkdirAll(p.directory, 0700); err != nil {
		return err
	}
	lock, err := os.OpenFile(filepath.Join(p.directory, key+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	// Also serialize separate daemon processes during a DaemonSet replacement.
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	path := filepath.Join(p.directory, key+".json")
	journal := timeJournal{Version: 1, BootID: bootID, ContainerID: containerID, PodContainerName: podContainerName,
		Tasks: make(map[string]persistentConfig), Completed: make(map[string]bool), Processes: make(map[string]*processRecovery)}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &journal); err != nil {
			return errors.Wrap(err, "decode time chaos recovery journal")
		}
		if journal.Version != 1 || journal.BootID != bootID || journal.PodContainerName != podContainerName || journal.Tasks == nil || journal.Completed == nil || journal.Processes == nil {
			return errors.New("invalid time chaos recovery journal")
		}
		if err := validateTimeJournal(&journal); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	} else if config == nil {
		// A pre-upgrade experiment has no durable originals. Never report it as
		// recovered just because the daemon has no in-memory task.
		return errors.New("time chaos recovery journal not found; cannot recover an experiment injected without durable state")
	}

	if journal.Cgroup == "" || (config != nil && journal.ContainerID != containerID) {
		journal.Cgroup, err = p.cgroup(pid)
		if err != nil {
			return errors.Wrap(err, "read container process cgroup")
		}
	}

	oldConfig, hadConfig := journal.Tasks[uid]
	wasCompleted := journal.Completed[uid]
	if config != nil {
		journal.ContainerID = containerID
		journal.Tasks[uid] = persistentConfig{config.deltaSeconds, config.deltaNanoSeconds, config.clockIDsMask}
		delete(journal.Completed, uid)
	} else {
		if _, exists := journal.Tasks[uid]; !exists && !journal.Completed[uid] {
			return errors.New("time chaos task not found in recovery journal")
		}
		delete(journal.Tasks, uid)
		journal.Completed[uid] = true
	}
	save := func() error { return writeTimeJournal(path, &journal) }
	// Persist intent first. A retry after any crash reconciles this desired set,
	// rather than adding/subtracting an offset for a second time.
	if err := save(); err != nil {
		return err
	}

	reconcileErr := p.reconcileProcesses(&journal, containerID, pid, save)
	if reconcileErr == nil || config == nil {
		return reconcileErr
	}
	// Failed Apply remains pending in the controller. Restore the previous
	// desired set so an unrelated task does not include this failed change in
	// its aggregate while the original request waits to be retried.
	if hadConfig {
		journal.Tasks[uid] = oldConfig
	} else {
		delete(journal.Tasks, uid)
	}
	if wasCompleted {
		journal.Completed[uid] = true
	} else {
		delete(journal.Completed, uid)
	}
	// Keep a tombstone so a concurrent/lost-response recovery remains safe.
	if !hadConfig {
		journal.Completed[uid] = true
	}
	if err := save(); err != nil {
		return errors.Wrapf(reconcileErr, "save apply rollback: %v", err)
	}
	if err := p.reconcileProcesses(&journal, containerID, pid, save); err != nil {
		return errors.Wrapf(reconcileErr, "apply rollback also failed: %v", err)
	}
	return reconcileErr
}

func (p *PersistentTimeChaos) reconcileProcesses(journal *timeJournal, containerID string, pid int, save func() error) error {

	var aggregate *Config
	if len(journal.Tasks) != 0 {
		merged := NewConfig(0, 0, 0)
		for _, task := range journal.Tasks {
			c := NewConfig(task.Seconds, task.Nanoseconds, task.ClockIDs)
			if err := merged.Merge(&c); err != nil {
				return err
			}
		}
		aggregate = &merged
	}

	var templates []processRecovery
	for _, process := range journal.Processes {
		templates = append(templates, *process)
	}
	// Include current container members as well as saved processes: fork
	// inherits patches, and a child may be reparented before the next request.
	processed := make(map[string]bool)
	for round := 0; round < 5; round++ {
		var selected []processIdentity
		var discoveryErr error
		if journal.ContainerID == containerID {
			leader, err := p.identity(pid)
			if err == nil {
				selected = append(selected, leader)
			} else if !os.IsNotExist(err) {
				discoveryErr = err
			}
			children, err := p.children(pid, journal.Cgroup, containerID)
			if err != nil {
				discoveryErr = err
			} else {
				selected = append(selected, children...)
			}
		}
		for _, identity := range selected {
			key := processIdentityKey(identity)
			if _, exists := journal.Processes[key]; !exists {
				journal.Processes[key] = &processRecovery{Identity: identity, ContainerID: containerID}
			}
		}

		if err := save(); err != nil {
			return err
		}
		keys := make([]string, 0, len(journal.Processes))
		for key := range journal.Processes {
			if !processed[key] {
				keys = append(keys, key)
			}
		}
		if len(keys) == 0 {
			return discoveryErr
		}
		sort.Slice(keys, func(i, j int) bool {
			a, b := journal.Processes[keys[i]].Identity.PID, journal.Processes[keys[j]].Identity.PID
			if a == b {
				return keys[i] < keys[j]
			}
			if a == pid {
				return true
			}
			if b == pid {
				return false
			}
			return a < b
		})
		var failures []string
		if discoveryErr != nil {
			failures = append(failures, discoveryErr.Error())
		}
		for _, key := range keys {
			process := journal.Processes[key]
			processed[key] = true
			current, err := p.identity(process.Identity.PID)
			if os.IsNotExist(err) || (err == nil && current != process.Identity) {
				continue
			}
			if err == nil {
				desired := aggregate
				if process.ContainerID != containerID {
					desired = nil
				}
				err = p.execute(process, desired, templates, save)
			}
			templates = append(templates, *process)
			if err != nil {
				failures = append(failures, fmt.Sprintf("pid %d: %v", process.Identity.PID, err))
			}
		}
		if len(failures) != 0 {
			return errors.New(strings.Join(failures, "; "))
		}
		// Rescan after restoring parents. A fork created between the previous
		// scan and ptrace attachment can still have inherited the old patch.
	}
	return errors.New("time chaos process group is changing; retry reconciliation")
}

func processIdentityKey(identity processIdentity) string {
	return strconv.Itoa(identity.PID) + ":" + identity.StartTime
}

func validateTimeJournal(journal *timeJournal) error {
	for key, process := range journal.Processes {
		if process == nil || process.Identity.PID <= 0 || process.ContainerID == "" || key != processIdentityKey(process.Identity) {
			return errors.New("invalid process in time chaos recovery journal")
		}
		if _, err := strconv.ParseUint(process.Identity.StartTime, 10, 64); err != nil {
			return errors.Wrap(err, "invalid recovery process start time")
		}
		for _, image := range []imageRecovery{process.ClockGetTime, process.GetTimeOfDay} {
			if len(image.OriginalCode) == 0 && image.OriginalAddress == 0 && image.FakeEntry == nil {
				continue
			}
			if len(image.OriginalCode) != 16 || image.OriginalAddress == 0 || image.FakeEntry == nil || len(image.Content) == 0 || image.FakeEntry.StartAddress == 0 || image.FakeEntry.EndAddress <= image.FakeEntry.StartAddress || uint64(len(image.Content)) > image.FakeEntry.EndAddress-image.FakeEntry.StartAddress || len(image.Offsets)*varLength >= len(image.Content) {
				return errors.New("incomplete image in time chaos recovery journal")
			}
			for _, offset := range image.Offsets {
				if offset < 0 || offset > len(image.Content)-varLength {
					return errors.New("invalid variable offset in time chaos recovery journal")
				}
			}
		}
	}
	return nil
}

func writeTimeJournal(path string, journal *timeJournal) (err error) {
	data, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".timechaos-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func (p *PersistentTimeChaos) executeProcess(process *processRecovery, config *Config, templates []processRecovery, save func() error) error {
	c := NewConfig(0, 0, 0)
	if config != nil {
		c = *config
	}
	skew, err := GetSkew(p.logger, c)
	if err != nil {
		return err
	}
	restoreImageRecovery(skew.clockGetTime, process.ClockGetTime)
	restoreImageRecovery(skew.getTimeOfDay, process.GetTimeOfDay)
	validate := func(pid int) error {
		identity, err := p.identity(pid)
		if err != nil {
			return err
		}
		if identity != process.Identity {
			return errors.New("process identity changed before ptrace attachment")
		}
		return nil
	}
	saveImages := func() error {
		process.ClockGetTime = snapshotImageRecovery(skew.clockGetTime)
		process.GetTimeOfDay = snapshotImageRecovery(skew.getTimeOfDay)
		return save()
	}
	for _, template := range templates {
		if template.ContainerID != process.ContainerID {
			continue
		}
		if len(template.ClockGetTime.OriginalCode) != 0 {
			skew.clockGetTime.recoveryCandidates = append(skew.clockGetTime.recoveryCandidates, template.ClockGetTime)
		}
		if len(template.GetTimeOfDay.OriginalCode) != 0 {
			skew.getTimeOfDay.recoveryCandidates = append(skew.getTimeOfDay.recoveryCandidates, template.GetTimeOfDay)
		}
	}
	for _, image := range []*FakeImage{skew.clockGetTime, skew.getTimeOfDay} {
		image.validateProcess = validate
		image.saveRecovery = saveImages
	}
	if config != nil {
		return skew.Inject(tasks.SysPID(process.Identity.PID))
	}
	return skew.Recover(tasks.SysPID(process.Identity.PID))
}

func snapshotImageRecovery(image *FakeImage) imageRecovery {
	return imageRecovery{
		OriginalCode:    append([]byte(nil), image.OriginFuncCode...),
		OriginalAddress: image.OriginAddress,
		FakeEntry:       image.fakeEntry,
		Content:         append([]byte(nil), image.content...),
		Offsets:         image.offset,
	}
}

func restoreImageRecovery(image *FakeImage, recovery imageRecovery) {
	image.OriginFuncCode = append([]byte(nil), recovery.OriginalCode...)
	image.OriginAddress = recovery.OriginalAddress
	image.fakeEntry = recovery.FakeEntry
	if len(recovery.Content) != 0 {
		image.content = append([]byte(nil), recovery.Content...)
		image.offset = recovery.Offsets
	}
}
