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

package trigger

import (
	"context"
	"sync"

	"github.com/go-logr/logr"

	"go.uber.org/atomic"
)

const defaultBufferSize = 10

// Multiplexer for trigger. Notice that it will change the context for the Acquire.
type CompositeTrigger struct {
	backends   []Trigger
	subscribed *atomic.Bool
	queue      chan EventOrError
	Logger     logr.Logger
}

func NewCompositeTrigger(backends ...Trigger) *CompositeTrigger {
	return &CompositeTrigger{
		backends:   backends,
		subscribed: atomic.NewBool(false),
		queue:      make(chan EventOrError, defaultBufferSize),
	}
}

func (it *CompositeTrigger) TriggerName() string {
	return "CompositeTrigger"
}

func (it *CompositeTrigger) Acquire(ctx context.Context) (Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case item := <-it.queue:
		return item.Event, item.Err
	}
}

func (it *CompositeTrigger) RunAndPending(ctx context.Context) error {
	working := atomic.NewBool(true)

	go func() {
		<-ctx.Done()
		working.Store(false)
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(it.backends))
	for _, item := range it.backends {
		eachBackends := item
		go func() {
			for working.Load() {
				event, err := eachBackends.Acquire(ctx)
				it.queue <- EventOrError{
					Event: event,
					Err:   err,
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
	if len(it.queue) > 0 {
		it.Logger.Info("there are still unconsumed events in buffering queue", "size", len(it.queue))
	}

	return nil
}

type EventOrError struct {
	Event Event
	Err   error
}
