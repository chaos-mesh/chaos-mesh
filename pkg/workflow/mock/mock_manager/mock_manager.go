// Copyright Chaos Mesh Authors.
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

// Code generated by MockGen. DO NOT EDIT.
// Source: ./manager/manager.go

// Package mock_manager is a generated GoMock package.
package mock_manager

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockWorkflowManager is a mock of WorkflowManager interface
type MockWorkflowManager struct {
	ctrl     *gomock.Controller
	recorder *MockWorkflowManagerMockRecorder
}

// MockWorkflowManagerMockRecorder is the mock recorder for MockWorkflowManager
type MockWorkflowManagerMockRecorder struct {
	mock *MockWorkflowManager
}

// NewMockWorkflowManager creates a new mock instance
func NewMockWorkflowManager(ctrl *gomock.Controller) *MockWorkflowManager {
	mock := &MockWorkflowManager{ctrl: ctrl}
	mock.recorder = &MockWorkflowManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkflowManager) EXPECT() *MockWorkflowManagerMockRecorder {
	return m.recorder
}

// GetName mocks base method
func (m *MockWorkflowManager) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName
func (mr *MockWorkflowManagerMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockWorkflowManager)(nil).GetName))
}

// Run mocks base method
func (m *MockWorkflowManager) Run(ctx context.Context) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run", ctx)
}

// Run indicates an expected call of Run
func (mr *MockWorkflowManagerMockRecorder) Run(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockWorkflowManager)(nil).Run), ctx)
}
