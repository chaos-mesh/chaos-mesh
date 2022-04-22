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
	"sync"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type Executor interface {
	// Do will execute according to the status check configuration,
	// returns:
	// 1. the result status (true for success, false for failure).
	// 2. output of execution.
	// 3. errors if any, it will lead to throw away the result of the execution.
	Do() (bool, string, error)
	// Type provides the type of executor
	Type() string
}

type worker struct {
	logger        logr.Logger
	eventRecorder recorder.ChaosRecorder

	// stopCh is a channel for stopping the worker.
	stopCh chan struct{}
	once   sync.Once

	manager *manager
	// Describes the status check configuration (read-only)
	statusCheck v1alpha1.StatusCheck
	executor    Executor

	lastResult      bool
	sameResultCount int
}

func newWorker(logger logr.Logger, eventRecorder recorder.ChaosRecorder,
	manager *manager, statusCheck v1alpha1.StatusCheck, executor Executor) *worker {
	return &worker{
		logger:        logger,
		eventRecorder: eventRecorder,
		manager:       manager,
		statusCheck:   statusCheck,
		executor:      executor,
		stopCh:        make(chan struct{}, 1), // non-blocking
	}
}

// run periodically execute the status check.
func (w *worker) run() {
	w.logger.V(1).Info("worker start")
	interval := time.Duration(w.statusCheck.Spec.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer func() {
		w.logger.V(1).Info("worker stop")
		ticker.Stop()
		key := types.NamespacedName{Namespace: w.statusCheck.Namespace, Name: w.statusCheck.Name}
		// delete worker from manager cache
		w.manager.workers.delete(key)
	}()

	for {
		select {
		case <-ticker.C:
			if !w.execute() {
				return
			}
		case <-w.stopCh:
			return
		}
	}
}

// stop stops the worker, it is safe to call stop multiple times.
func (w *worker) stop() {
	w.once.Do(func() {
		close(w.stopCh)
	})
}

// execute the status check once and records the result.
// Returns whether the worker should continue.
func (w *worker) execute() bool {
	startTime := time.Now()
	result, output, err := w.executor.Do()
	if err != nil {
		// executor error, throw away the result.
		w.logger.Error(err, "executor internal error")
		return true
	}

	if w.lastResult == result {
		w.sameResultCount++
	} else {
		w.lastResult = result
		w.sameResultCount = 1
	}

	key := types.NamespacedName{Namespace: w.statusCheck.Namespace, Name: w.statusCheck.Name}
	if result {
		w.logger.V(1).Info("status check execution succeed", "msg", output)
		w.eventRecorder.Event(&w.statusCheck, recorder.StatusCheckExecutionSucceed{ExecutorType: w.executor.Type()})
		w.manager.results.append(key, v1alpha1.StatusCheckRecord{
			StartTime: &metav1.Time{Time: startTime},
			Outcome:   v1alpha1.StatusCheckOutcomeSuccess,
		})

		// check if the success threshold is exceeded
		// Notice: the function `setSuccessThresholdExceedCondition` in `controllers/statuscheck/conditions.go`
		// also checks the success threshold, so if you want to modify the logic here, don't forget to modify that
		// function as well.
		if w.statusCheck.Spec.Mode == v1alpha1.StatusCheckSynchronous &&
			w.sameResultCount >= w.statusCheck.Spec.SuccessThreshold {
			w.logger.Info("exceed the success threshold")
			// if status check mode is Synchronous, and it exceeds the SuccessThreshold,
			// then stop the worker
			return false
		}
	} else {
		w.logger.Info("status check execution failed", "msg", output)
		w.eventRecorder.Event(&w.statusCheck, recorder.StatusCheckExecutionFailed{ExecutorType: w.executor.Type(), Msg: output})
		w.manager.results.append(key, v1alpha1.StatusCheckRecord{
			StartTime: &metav1.Time{Time: startTime},
			Outcome:   v1alpha1.StatusCheckOutcomeFailure,
		})

		// check if the failure threshold is exceeded
		// Notice: the function `setFailureThresholdExceedCondition` in `controllers/statuscheck/conditions.go`
		// also checks the failure threshold, so if you want to modify the logic here, don't forget to modify that
		// function as well.
		if w.sameResultCount >= w.statusCheck.Spec.FailureThreshold {
			w.logger.Info("exceed the failure threshold")
			// if it exceeds the FailureThreshold, stop the worker
			return false
		}
	}

	return true
}
