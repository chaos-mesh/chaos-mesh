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
	"k8s.io/client-go/util/workqueue"
	"time"
)

type OperableTrigger interface {
	Trigger
	Notify(event Event) error
	NotifyDelay(event Event, delay time.Duration) error
}

type basicOperableTrigger struct {
	queue workqueue.DelayingInterface
}

func NewOperableTrigger() OperableTrigger {
	return &basicOperableTrigger{
		queue: workqueue.NewDelayingQueue(),
	}
}

func (it *basicOperableTrigger) TriggerName() string {
	return "OperableTrigger"
}

func (it *basicOperableTrigger) Acquire(ctx context.Context) (Event, error) {
	item, shutdown := it.queue.Get()

	if shutdown {
		return nil, fmt.Errorf("inner queue already shutdown")
	}
	it.queue.Done(item)

	event, ok := item.(Event)
	if !ok {
		return nil, fmt.Errorf("item could not be assert as Event")
	}
	return event, nil
}

func (it *basicOperableTrigger) Notify(event Event) error {
	it.queue.Add(event)
	return nil
}

func (it *basicOperableTrigger) NotifyDelay(event Event, delay time.Duration) error {
	it.queue.AddAfter(event, delay)
	return nil
}
