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
	"fmt"
	"sync"
	"time"

	"go.uber.org/atomic"

	"k8s.io/client-go/util/workqueue"
)

type OperableTrigger interface {
	Trigger
	Notify(event Event) error
	NotifyDelay(event Event, delay time.Duration) error
}

type basicOperableTrigger struct {
	queue  workqueue.DelayingInterface
	missed *syncQueue
}

func NewOperableTrigger() OperableTrigger {
	return &basicOperableTrigger{
		queue:  workqueue.NewDelayingQueue(),
		missed: newSyncQueue(),
	}
}

func (it *basicOperableTrigger) TriggerName() string {
	return "OperableTrigger"
}

func (it *basicOperableTrigger) Acquire(ctx context.Context) (Event, bool, error) {

	deq := it.missed.deq()
	if deq != nil {
		return deq.event, false, deq.err
	}

	canceled := atomic.NewBool(false)

	result := make(chan EventAndError, 1)
	go func() {
		defer close(result)
		item, shutdown := it.queue.Get()

		if shutdown {
			result <- EventAndError{
				event: nil,
				err:   fmt.Errorf("inner queue already shutdown"),
			}
			return
		}
		defer it.queue.Done(item)

		event, ok := item.(Event)
		if !ok {

			result <- EventAndError{
				event: nil,
				err:   fmt.Errorf("item could not be assert as Event"),
			}
			return
		}

		if canceled.Load() {
			it.missed.enq(&EventAndError{
				event: event,
				err:   nil,
			})
		} else {
			result <- EventAndError{
				event: event,
				err:   nil,
			}
		}

	}()

	select {
	case item := <-result:
		return item.event, false, item.err
	case <-ctx.Done():
		canceled.Store(true)
		return nil, true, ctx.Err()
	}
}

func (it *basicOperableTrigger) Notify(event Event) error {
	it.queue.Add(event)
	return nil
}

func (it *basicOperableTrigger) NotifyDelay(event Event, delay time.Duration) error {
	it.queue.AddAfter(event, delay)
	return nil
}

type syncQueue struct {
	sync.Mutex
	store []*EventAndError
}

func newSyncQueue() *syncQueue {
	return &syncQueue{
		Mutex: sync.Mutex{},
		store: nil,
	}
}

func (it *syncQueue) deq() *EventAndError {
	it.Lock()
	defer it.Unlock()
	if len(it.store) > 0 {
		result := it.store[0]
		it.store = it.store[1:]
		return result
	}
	return nil
}

func (it *syncQueue) enq(event *EventAndError) {
	it.Lock()
	defer it.Unlock()
	it.store = append(it.store, event)
}
func (it *syncQueue) isEmpty() bool {
	it.Lock()
	defer it.Unlock()
	return len(it.store) == 0
}

type EventAndError struct {
	event Event
	err   error
}
