// Copyright Chaos Mesh Authors.
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

package statuscheck

import (
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/go-logr/logr"
	"sync"
)

type Manager interface {
	// Add creates new probe workers for every container probe. This should be called for every
	// pod created.
	Add(statusCheck v1alpha1.StatusCheck)
	// Remove handles cleaning up the removed pod state, including terminating probe workers and
	// deleting cached results.
	Remove(statusCheck v1alpha1.StatusCheck)

	Get(statusCheck v1alpha1.StatusCheck)
}

type manager struct {
	// Map of active workers for probes
	workers map[string]*worker
	// Lock for accessing & mutating workers
	workerLock sync.RWMutex
	// result
	result map[string]string

	logger        logr.Logger
	eventRecorder recorder.ChaosRecorder
}

func NewManager(eventRecorder recorder.ChaosRecorder, logger logr.Logger) Manager {
	return &manager{
		workers:       make(map[string]*worker),
		eventRecorder: eventRecorder,
		logger:        logger,
	}
}

func (m *manager) Add(statusCheck v1alpha1.StatusCheck) {
	m.workerLock.Lock()
	defer m.workerLock.Unlock()

	key := string(statusCheck.UID)
	if _, ok := m.workers[key]; ok {
		return
	}

	worker := newWorker(m, statusCheck, m.logger.WithName("statuscheck-worker"))
	m.workers[key] = worker
	go worker.run()
}

func (m *manager) Remove(statusCheck v1alpha1.StatusCheck) {
	m.workerLock.Lock()
	defer m.workerLock.Unlock()

	key := string(statusCheck.UID)
	worker, ok := m.workers[key]
	if !ok {
		return
	}
	worker.stop()
}

func (m *manager) Get(statusCheck v1alpha1.StatusCheck) {
	m.workerLock.RLock()
	defer m.workerLock.RUnlock()
}

func (m *manager) getWorker(uid string) (*worker, bool) {
	m.workerLock.RLock()
	defer m.workerLock.RUnlock()
	worker, ok := m.workers[uid]
	return worker, ok
}

func (m *manager) removeWorker(uid string) {
	m.workerLock.Lock()
	defer m.workerLock.Unlock()
	delete(m.workers, uid)
}
