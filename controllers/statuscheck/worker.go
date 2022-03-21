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
	"github.com/chaos-mesh/chaos-mesh/controllers/statuscheck/http"
	"github.com/go-logr/logr"
	"time"
)

type worker struct {
	stopCh  chan struct{}
	manager *manager

	statusCheck v1alpha1.StatusCheck
	executor    Executor
	logger      logr.Logger
}

func newWorker(manager *manager, statusCheck v1alpha1.StatusCheck, logger logr.Logger) *worker {
	var executor Executor
	switch statusCheck.Spec.Type {
	case v1alpha1.TypeHTTP:
		executor = http.NewExecutor()
	default:
		// TODO handle this error
	}

	return &worker{
		stopCh:      make(chan struct{}),
		manager:     manager,
		statusCheck: statusCheck,
		executor:    executor,
		logger:      logger,
	}
}

// run periodically execute the status check.
func (w *worker) run() {
	interval := time.Duration(w.statusCheck.Spec.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer func() {
		ticker.Stop()
		w.manager.removeWorker(string(w.statusCheck.UID))
	}()

	for w.execute() {
		select {
		case <-ticker.C:
		case <-w.stopCh:
			return
		}
	}
}

func (w *worker) stop() {
	close(w.stopCh)
}

// execute the status check once and records the result.
// Returns whether the worker should continue.
func (w *worker) execute() bool {
	// TODO
	result, err := w.executor.Do(w.statusCheck.Spec)
	if err != nil {
		return false
	}
	w.logger.Info("execute", "result", result)
	return true
}
