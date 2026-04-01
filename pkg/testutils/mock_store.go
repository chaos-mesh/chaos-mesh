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

package testutils

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

type MockExperimentStore struct {
	mock.Mock
}

func (m *MockExperimentStore) ListMeta(ctx context.Context, kind, namespace, name string, archived bool) ([]*core.ExperimentMeta, error) {
	args := m.Called(ctx, kind, namespace, name, archived)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ExperimentMeta), args.Error(1)
}

func (m *MockExperimentStore) FindByUID(ctx context.Context, UID string) (*core.Experiment, error) {
	args := m.Called(ctx, UID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Experiment), args.Error(1)
}

func (m *MockExperimentStore) FindManagedByNamespaceName(ctx context.Context, namespace, name string) ([]*core.Experiment, error) {
	args := m.Called(ctx, namespace, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Experiment), args.Error(1)
}

func (m *MockExperimentStore) FindMetaByUID(ctx context.Context, UID string) (*core.ExperimentMeta, error) {
	args := m.Called(ctx, UID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ExperimentMeta), args.Error(1)
}

func (m *MockExperimentStore) Set(ctx context.Context, exp *core.Experiment) error {
	args := m.Called(ctx, exp)
	return args.Error(0)
}

func (m *MockExperimentStore) Archive(ctx context.Context, namespace, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func (m *MockExperimentStore) Delete(ctx context.Context, exp *core.Experiment) error {
	args := m.Called(ctx, exp)
	return args.Error(0)
}

func (m *MockExperimentStore) DeleteByFinishTime(ctx context.Context, duration time.Duration) error {
	args := m.Called(ctx, duration)
	return args.Error(0)
}

func (m *MockExperimentStore) DeleteIncompleteExperiments(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockExperimentStore) DeleteByUIDs(ctx context.Context, uids []string) error {
	args := m.Called(ctx, uids)
	return args.Error(0)
}

type MockScheduleStore struct {
	mock.Mock
}

func (m *MockScheduleStore) ListMeta(ctx context.Context, namespace, name string, archived bool) ([]*core.ScheduleMeta, error) {
	args := m.Called(ctx, namespace, name, archived)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.ScheduleMeta), args.Error(1)
}

func (m *MockScheduleStore) FindByUID(ctx context.Context, UID string) (*core.Schedule, error) {
	args := m.Called(ctx, UID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Schedule), args.Error(1)
}

func (m *MockScheduleStore) FindMetaByUID(ctx context.Context, UID string) (*core.ScheduleMeta, error) {
	args := m.Called(ctx, UID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.ScheduleMeta), args.Error(1)
}

func (m *MockScheduleStore) Set(ctx context.Context, sch *core.Schedule) error {
	args := m.Called(ctx, sch)
	return args.Error(0)
}

func (m *MockScheduleStore) Archive(ctx context.Context, namespace, name string) error {
	args := m.Called(ctx, namespace, name)
	return args.Error(0)
}

func (m *MockScheduleStore) Delete(ctx context.Context, sch *core.Schedule) error {
	args := m.Called(ctx, sch)
	return args.Error(0)
}

func (m *MockScheduleStore) DeleteByFinishTime(ctx context.Context, duration time.Duration) error {
	args := m.Called(ctx, duration)
	return args.Error(0)
}

func (m *MockScheduleStore) DeleteByUIDs(ctx context.Context, uids []string) error {
	args := m.Called(ctx, uids)
	return args.Error(0)
}

func (m *MockScheduleStore) DeleteIncompleteSchedules(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockEventStore struct {
	mock.Mock
}

func (m *MockEventStore) List(ctx context.Context) ([]*core.Event, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Event), args.Error(1)
}

func (m *MockEventStore) ListByUID(ctx context.Context, uid string) ([]*core.Event, error) {
	args := m.Called(ctx, uid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Event), args.Error(1)
}

func (m *MockEventStore) ListByUIDs(ctx context.Context, uids []string) ([]*core.Event, error) {
	args := m.Called(ctx, uids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Event), args.Error(1)
}

func (m *MockEventStore) ListByExperiment(ctx context.Context, namespace string, name string, kind string) ([]*core.Event, error) {
	args := m.Called(ctx, namespace, name, kind)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Event), args.Error(1)
}

func (m *MockEventStore) ListByFilter(ctx context.Context, filter core.Filter) ([]*core.Event, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*core.Event), args.Error(1)
}

func (m *MockEventStore) Find(ctx context.Context, id uint) (*core.Event, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*core.Event), args.Error(1)
}

func (m *MockEventStore) Create(ctx context.Context, event *core.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockEventStore) DeleteByUID(ctx context.Context, uid string) error {
	args := m.Called(ctx, uid)
	return args.Error(0)
}

func (m *MockEventStore) DeleteByUIDs(ctx context.Context, uids []string) error {
	args := m.Called(ctx, uids)
	return args.Error(0)
}

func (m *MockEventStore) DeleteByTime(ctx context.Context, startTime string, endTime string) error {
	args := m.Called(ctx, startTime, endTime)
	return args.Error(0)
}

func (m *MockEventStore) DeleteByDuration(ctx context.Context, duration time.Duration) error {
	args := m.Called(ctx, duration)
	return args.Error(0)
}
