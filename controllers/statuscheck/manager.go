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
	"crypto/x509"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/statuscheck/http"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type Manager interface {
	// Add creates new workers for every status check.
	Add(statusCheck v1alpha1.StatusCheck) error
	// Get returns the cached results about the status check.
	Get(statusCheck v1alpha1.StatusCheck) (Result, bool)
	// Delete handles cleaning up the removed status check state, including terminating workers and
	// deleting cached results.
	// This should be called when StatusCheck is deleted.
	Delete(key types.NamespacedName)
	// Complete handles terminating workers, but not deleting cached results.
	// This should be called when StatusCheck is completed.
	Complete(statusCheck v1alpha1.StatusCheck)
}

type manager struct {
	logger        logr.Logger
	eventRecorder recorder.ChaosRecorder

	workers     workerCache
	results     resultCache
	newExecutor newExecutorFunc
	certPool    *x509.CertPool
}

type newExecutorFunc func(logger logr.Logger, certPool *x509.CertPool, statusCheck v1alpha1.StatusCheck) (Executor, error)

func NewManager(logger logr.Logger, eventRecorder recorder.ChaosRecorder, certPool *x509.CertPool, newExecutorFunc newExecutorFunc) Manager {
	return &manager{
		logger:        logger,
		eventRecorder: eventRecorder,
		workers:       workerCache{workers: sync.Map{}},
		results:       resultCache{results: make(map[types.NamespacedName]Result)},
		newExecutor:   newExecutorFunc,
		certPool:      certPool,
	}
}

func (m *manager) Add(statusCheck v1alpha1.StatusCheck) error {
	key := types.NamespacedName{Namespace: statusCheck.Namespace, Name: statusCheck.Name}
	if _, ok := m.results.get(key); ok {
		return nil
	}
	m.results.init(key, statusCheck.Status.Records, statusCheck.Status.Count, uint(statusCheck.Spec.RecordsHistoryLimit))

	if statusCheck.IsCompleted() {
		// if status check is completed, there is no need to create a worker
		return errors.New("status check is completed")
	}

	executor, err := m.newExecutor(m.logger, m.certPool, statusCheck)
	if err != nil {
		return errors.Wrap(err, "new executor")
	}
	worker := newWorker(m.logger.WithName("worker").WithValues("statuscheck", key), m.eventRecorder, m, statusCheck, executor)
	m.workers.add(key, worker)
	return nil
}

func (m *manager) Get(statusCheck v1alpha1.StatusCheck) (Result, bool) {
	key := types.NamespacedName{Namespace: statusCheck.Namespace, Name: statusCheck.Name}
	result, ok := m.results.get(key)
	if !ok {
		return Result{}, false
	}
	return result, true
}

func (m *manager) Delete(key types.NamespacedName) {
	m.results.delete(key)
	m.workers.delete(key)
}

func (m *manager) Complete(statusCheck v1alpha1.StatusCheck) {
	key := types.NamespacedName{Namespace: statusCheck.Namespace, Name: statusCheck.Name}
	m.workers.delete(key)
}

// workerCache provides cached workers.
type workerCache struct {
	// Map of NamespacedName of StatusCheck -> *worker
	workers sync.Map
}

func (c *workerCache) add(key types.NamespacedName, worker *worker) {
	_, ok := c.workers.LoadOrStore(key, worker)
	if !ok {
		go worker.run()
	}
}

func (c *workerCache) delete(key types.NamespacedName) {
	obj, ok := c.workers.LoadAndDelete(key)
	if !ok {
		return
	}
	worker := obj.(*worker)
	worker.stop()
}

// resultCache provides cached status check results.
type resultCache struct {
	// Map of NamespacedName of StatusCheck -> *result
	results map[types.NamespacedName]Result
	lock    sync.RWMutex
}

type Result struct {
	Records []v1alpha1.StatusCheckRecord
	Count   int64
	// recordsHistoryLimit defines the number of record to retain.
	recordsHistoryLimit uint
}

// init should only be called when adding a new worker.
func (c *resultCache) init(key types.NamespacedName, obj []v1alpha1.StatusCheckRecord, count int64, limit uint) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.results[key]; ok {
		return
	}
	if len(obj) == 0 {
		obj = make([]v1alpha1.StatusCheckRecord, 0)
		count = 0
	}
	c.results[key] = Result{
		Records:             limitRecords(obj, limit),
		Count:               count,
		recordsHistoryLimit: limit,
	}
}

// append will append the record to the cache.
// It should be only called by worker
func (c *resultCache) append(key types.NamespacedName, obj v1alpha1.StatusCheckRecord) {
	c.lock.Lock()
	defer c.lock.Unlock()

	result := c.results[key]
	result.Records = append(result.Records, obj)
	result.Records = limitRecords(result.Records, result.recordsHistoryLimit)
	result.Count++
	c.results[key] = result
}

func (c *resultCache) delete(key types.NamespacedName) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.results, key)
}

func (c *resultCache) get(key types.NamespacedName) (Result, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	result, ok := c.results[key]
	return result, ok
}

func limitRecords(records []v1alpha1.StatusCheckRecord, limit uint) []v1alpha1.StatusCheckRecord {
	length := len(records)
	if length < int(limit) {
		return records
	}
	return records[length-int(limit):]
}

func newExecutor(logger logr.Logger, certPool *x509.CertPool, statusCheck v1alpha1.StatusCheck) (Executor, error) {
	var executor Executor
	switch statusCheck.Spec.Type {
	case v1alpha1.TypeHTTP:
		if statusCheck.Spec.EmbedStatusCheck == nil || statusCheck.Spec.HTTPStatusCheck == nil {
			// this should not happen, if the webhook works as expected
			return nil, errors.New("illegal status check, http should not be empty")
		}
		executor = http.NewExecutor(
			logger.WithName("http-executor").WithValues("url", statusCheck.Spec.HTTPStatusCheck.RequestUrl),
			certPool,
			statusCheck.Spec.TimeoutSeconds,
			*statusCheck.Spec.HTTPStatusCheck,
		)
	default:
		return nil, errors.Errorf("unsupported type '%s'", statusCheck.Spec.Type)
	}
	return executor, nil
}
