// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package twophase

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const iterMax = 1e4

type chaosStateMachine struct {
	Chaos v1alpha1.InnerSchedulerObject
	Req   ctrl.Request
	*Reconciler
}

func unexpected(ctx context.Context, m *chaosStateMachine, targetPhase v1alpha1.ExperimentPhase, now time.Time) (bool, error) {
	currentPhase := m.Chaos.GetStatus().Experiment.Phase

	return true, errors.Errorf("turn from %s into %s is unexpected", currentPhase, targetPhase)
}

func noop(ctx context.Context, m *chaosStateMachine, targetPhase v1alpha1.ExperimentPhase, now time.Time) (bool, error) {
	updated := false
	currentPhase := m.Chaos.GetStatus().Experiment.Phase

	if currentPhase != targetPhase {
		m.Chaos.GetStatus().Experiment.Phase = targetPhase
		updated = true
	}
	return updated, nil
}

func apply(ctx context.Context, m *chaosStateMachine, targetPhase v1alpha1.ExperimentPhase, startTime time.Time) (bool, error) {
	duration, err := m.Chaos.GetDuration()
	if err != nil {
		m.Log.Error(err, "failed to get chaos duration")
		return false, err
	}
	if duration == nil {
		zero := time.Duration(0)
		duration = &zero
	}

	currentPhase := m.Chaos.GetStatus().Experiment.Phase
	status := m.Chaos.GetStatus()

	m.Log.Info("applying", "current phase", currentPhase, "target phase", targetPhase)
	err = m.Apply(ctx, m.Req, m.Chaos)
	if err != nil {
		m.Log.Error(err, "fail to apply")

		status.Experiment.Phase = v1alpha1.ExperimentPhaseFailed
		status.FailedMessage = err.Error()

		return true, err
	}
	// reset failed message
	status.FailedMessage = emptyString
	status.Experiment.Phase = targetPhase

	nextStart, nextRecover, err := m.IterateNextTime(startTime, *duration)
	if err != nil {
		m.Log.Error(err, "failed to get the next start time and recover time")
		return true, err
	}

	m.Chaos.SetNextStart(*nextStart)
	m.Chaos.SetNextRecover(*nextRecover)

	status.Experiment.StartTime = &metav1.Time{Time: startTime}
	status.Experiment.EndTime = nil
	status.Experiment.Duration = duration.String()

	return true, nil
}

func recover(ctx context.Context, m *chaosStateMachine, targetPhase v1alpha1.ExperimentPhase, now time.Time) (bool, error) {
	currentPhase := m.Chaos.GetStatus().Experiment.Phase
	status := m.Chaos.GetStatus()

	m.Log.Info("recovering", "current phase", currentPhase, "target phase", targetPhase)
	err := m.Recover(ctx, m.Req, m.Chaos)
	if err != nil {
		status.FailedMessage = err.Error()

		m.Log.Error(err, "fail to recover")
		return true, err
	}
	status.Experiment.Phase = targetPhase
	status.Experiment.EndTime = &metav1.Time{
		Time: now,
	}
	if status.Experiment.StartTime != nil {
		status.Experiment.Duration = now.Sub(status.Experiment.StartTime.Time).String()
	}
	return true, nil
}

func resume(ctx context.Context, m *chaosStateMachine, _ v1alpha1.ExperimentPhase, now time.Time) (bool, error) {
	startTime := now
	duration, err := m.Chaos.GetDuration()
	if err != nil {
		m.Log.Error(err, "failed to get chaos duration")
		return false, err
	}
	if duration == nil {
		zero := time.Duration(0)
		duration = &zero
	}
	status := m.Chaos.GetStatus()

	nextStart := m.Chaos.GetNextStart()
	nextRecover := m.Chaos.GetNextRecover()
	var lastStart time.Time
	if status.Experiment.StartTime == nil {
		// in this condition, the experiment has never executed
		nextStart = now
		lastStart = now
	} else {
		lastStart = status.Experiment.StartTime.Time
	}

	defer func() {
		m.Chaos.SetNextStart(nextStart)
		m.Chaos.SetNextRecover(nextRecover)
	}()

	counter := 0
	// nextStart is always after nextRecover
	for {
		if nextRecover.After(now) {
			startTime = lastStart

			return apply(ctx, m, v1alpha1.ExperimentPhaseRunning, startTime)
		}

		if nextStart.After(now) {
			return noop(ctx, m, v1alpha1.ExperimentPhaseWaiting, now)
		}

		lastStart = nextStart
		start, recover, err := m.IterateNextTime(nextStart, *duration)
		if err != nil {
			m.Log.Error(err, "failed to get the next start time and recover time")

			return false, err
		}

		nextStart = *start
		nextRecover = *recover

		counter++
		if counter > iterMax {
			err = errors.Errorf("the number of iterations exceeded while resuming from pause with nextRecover(%s) nextStart(%s)", nextRecover, nextStart)
			return false, err
		}
	}
}

// This method changes the phase of an object and do some side effects
// There are 6 different phases, so there could be 6 * 6 = 36 branches
func (m *chaosStateMachine) run(ctx context.Context, targetPhase v1alpha1.ExperimentPhase, now time.Time) (bool, error) {
	currentPhase := m.Chaos.GetStatus().Experiment.Phase
	m.Log.Info("change phase", "current phase", currentPhase, "target phase", targetPhase)

	return phaseTransitionMap[currentPhase][targetPhase](ctx, m, targetPhase, now)
}

func (m *chaosStateMachine) Into(ctx context.Context, targetPhase v1alpha1.ExperimentPhase, now time.Time) error {
	updated, err := m.run(ctx, targetPhase, now)
	if err != nil {
		m.Log.Error(err, "error while executing state machine")
	}

	if updated {

		updateError := m.Update(ctx, m.Chaos)
		if updateError == nil {
			return err
		}

		if !strings.Contains(updateError.Error(), "the object has been modified; please apply your changes to the latest version and try again") {
			return updateError
		}

		m.Log.Error(updateError, "fail to update, and will retry on conflict")

		// avoid panic
		if m.Chaos.GetChaos() == nil || m.Chaos.GetStatus() == nil {
			return updateError
		}

		namespacedName := types.NamespacedName{
			Namespace: m.Chaos.GetChaos().Namespace,
			Name:      m.Chaos.GetChaos().Name,
		}

		updateError = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
			// Fetch the resource
			_chaos := m.Object()
			if err := m.Client.Get(ctx, namespacedName, _chaos); err != nil {
				m.Log.Error(err, "unable to get chaos")
				return err
			}
			chaos := _chaos.(v1alpha1.InnerSchedulerObject)

			// Make updates to the resource
			status := chaos.GetStatus()
			status.FailedMessage = m.Chaos.GetStatus().FailedMessage
			status.Scheduler = m.Chaos.GetStatus().Scheduler
			status.Experiment = m.Chaos.GetStatus().Experiment

			// Try to update
			return m.Update(ctx, chaos)
		})

		if updateError != nil {
			return updateError
		}
	}

	return err
}

func (m *chaosStateMachine) IterateNextTime(startTime time.Time, duration time.Duration) (*time.Time, *time.Time, error) {
	scheduler := m.Chaos.GetScheduler()
	if scheduler == nil {
		return nil, nil, errors.Errorf("misdefined scheduler")
	}
	m.Log.Info("iterate nextStart and nextRecover", "startTime", startTime, "duration", duration, "scheduler", scheduler)
	nextStart, err := nextTime(*scheduler, startTime)

	if err != nil {
		m.Log.Error(err, "failed to get the next start time")
		return nil, nil, err
	}
	nextRecover := startTime.Add(duration)

	counter := 0
	// if the duration is too long, `nextRecover` could be after `nextStart`
	// we can jump over a start to make sure `nextRecover` is before `nextStart`
	for nextRecover.After(*nextStart) {
		nextStart, err = nextTime(*scheduler, *nextStart)
		if err != nil {
			m.Log.Error(err, "failed to get the next start time")
			return nil, nil, err
		}

		counter++
		if counter > iterMax {
			err = errors.Errorf("the number of iterations exceeded with nextRecover(%s) nextStart(%s)", nextRecover, nextStart)
			return nil, nil, err
		}
	}

	return nextStart, &nextRecover, nil
}

var phaseTransitionMap = map[v1alpha1.ExperimentPhase]map[v1alpha1.ExperimentPhase]func(ctx context.Context, m *chaosStateMachine, targetPhase v1alpha1.ExperimentPhase, now time.Time) (bool, error){
	v1alpha1.ExperimentPhaseUninitialized: {
		v1alpha1.ExperimentPhaseUninitialized: noop,
		v1alpha1.ExperimentPhaseRunning:       apply,
		v1alpha1.ExperimentPhaseWaiting:       noop,
		v1alpha1.ExperimentPhasePaused:        noop,
		v1alpha1.ExperimentPhaseFailed:        unexpected,
		v1alpha1.ExperimentPhaseFinished:      noop,
	},
	v1alpha1.ExperimentPhaseRunning: {
		v1alpha1.ExperimentPhaseUninitialized: unexpected,
		v1alpha1.ExperimentPhaseRunning:       noop,
		v1alpha1.ExperimentPhaseWaiting:       recover,
		v1alpha1.ExperimentPhasePaused:        recover,
		v1alpha1.ExperimentPhaseFailed:        unexpected,
		v1alpha1.ExperimentPhaseFinished:      recover,
	},
	v1alpha1.ExperimentPhaseWaiting: {
		v1alpha1.ExperimentPhaseUninitialized: unexpected,
		v1alpha1.ExperimentPhaseRunning:       apply,
		v1alpha1.ExperimentPhaseWaiting:       noop,
		v1alpha1.ExperimentPhasePaused:        noop,
		v1alpha1.ExperimentPhaseFailed:        unexpected,
		v1alpha1.ExperimentPhaseFinished:      noop,
	},
	v1alpha1.ExperimentPhasePaused: {
		v1alpha1.ExperimentPhaseUninitialized: unexpected,
		v1alpha1.ExperimentPhaseRunning:       resume,
		v1alpha1.ExperimentPhaseWaiting:       resume,
		v1alpha1.ExperimentPhasePaused:        noop,
		v1alpha1.ExperimentPhaseFailed:        unexpected,
		v1alpha1.ExperimentPhaseFinished:      noop,
	},
	v1alpha1.ExperimentPhaseFailed: {
		v1alpha1.ExperimentPhaseUninitialized: unexpected,
		v1alpha1.ExperimentPhaseRunning:       apply,
		v1alpha1.ExperimentPhaseWaiting:       noop,
		v1alpha1.ExperimentPhasePaused:        noop,
		v1alpha1.ExperimentPhaseFailed:        noop,
		v1alpha1.ExperimentPhaseFinished:      noop,
	},
	v1alpha1.ExperimentPhaseFinished: {
		v1alpha1.ExperimentPhaseUninitialized: unexpected,
		v1alpha1.ExperimentPhaseRunning:       unexpected,
		v1alpha1.ExperimentPhaseWaiting:       unexpected,
		v1alpha1.ExperimentPhasePaused:        unexpected,
		v1alpha1.ExperimentPhaseFailed:        unexpected,
		v1alpha1.ExperimentPhaseFinished:      noop,
	},
}
