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
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotifyThenAcquire(t *testing.T) {
	const namespace = "mock-ns"
	const workflowName = "mock-workflow"
	const nodeName = "mock-node"
	const eventType = NodeCreated
	trigger := NewOperableTrigger()
	expected := NewEvent(namespace, workflowName, nodeName, eventType)
	err := trigger.Notify(expected)
	assert.NoError(t, err)
	actual, canceled, err := trigger.Acquire(context.TODO())
	assert.NoError(t, err)
	assert.False(t, canceled)
	assert.Equal(t, expected, actual)
}

func TestAcquireThenNotify(t *testing.T) {
	const namespace = "mock-ns"
	const workflowName = "mock-workflow"
	const nodeName = "mock-node"
	const eventType = NodeCreated
	trigger := NewOperableTrigger()
	expected := NewEvent(namespace, workflowName, nodeName, eventType)
	go func() {
		time.Sleep(time.Second)
		err := trigger.Notify(expected)
		assert.NoError(t, err)
	}()
	actual, canceled, err := trigger.Acquire(context.TODO())
	assert.NoError(t, err)
	assert.False(t, canceled)
	assert.Equal(t, expected, actual)
}

func TestNotifyAndAcquireParallel(t *testing.T) {
	const namespace = "mock-ns"
	const workflowName = "mock-workflow"
	const eventType = NodeCreated
	trigger := NewOperableTrigger()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			expected := NewEvent(namespace, workflowName, fmt.Sprintf("%d", i), eventType)
			err := trigger.Notify(expected)
			assert.NoError(t, err)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			actual, canceled, err := trigger.Acquire(context.TODO())
			assert.NoError(t, err)
			assert.False(t, canceled)
			assert.Equal(t, fmt.Sprintf("%d", i), actual.GetNodeName())
		}
	}()
	wg.Wait()
}

func TestFirstAcquireCanceledThenReturnInNextAcquire(t *testing.T) {
	const namespace = "mock-ns"
	const workflowName = "mock-workflow"
	const nodeName = "mock-node"
	const eventType = NodeCreated
	trigger := NewOperableTrigger()
	expected := NewEvent(namespace, workflowName, nodeName, eventType)

	ctx, cancelFunc := context.WithCancel(context.TODO())
	cancelFunc()
	// first acquire
	actual, canceled, err := trigger.Acquire(ctx)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.True(t, canceled)
	assert.Nil(t, actual)

	err = trigger.Notify(expected)
	assert.NoError(t, err)

	// second acquire
	actual, canceled, err = trigger.Acquire(context.TODO())
	assert.NoError(t, err)
	assert.False(t, canceled)
	assert.Equal(t, expected, actual)
}

func BenchmarkNotifyAndAcquireParallel(b *testing.B) {
	const namespace = "mock-ns"
	const workflowName = "mock-workflow"
	const eventType = NodeCreated
	trigger := NewOperableTrigger()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			expected := NewEvent(namespace, workflowName, fmt.Sprintf("%d", i), eventType)
			err := trigger.Notify(expected)
			assert.NoError(b, err)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < b.N; i++ {
			actual, canceled, err := trigger.Acquire(context.TODO())
			assert.NoError(b, err)
			assert.False(b, canceled)
			assert.Equal(b, fmt.Sprintf("%d", i), actual.GetNodeName())
		}
	}()
	wg.Wait()
}
