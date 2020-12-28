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
// Source: ./engine/model/template/template.go

// Package mock_template is a generated GoMock package.
package mock_template

import (
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"

	template "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
)

// MockTemplate is a mock of Template interface
type MockTemplate struct {
	ctrl     *gomock.Controller
	recorder *MockTemplateMockRecorder
}

// MockTemplateMockRecorder is the mock recorder for MockTemplate
type MockTemplateMockRecorder struct {
	mock *MockTemplate
}

// NewMockTemplate creates a new mock instance
func NewMockTemplate(ctrl *gomock.Controller) *MockTemplate {
	mock := &MockTemplate{ctrl: ctrl}
	mock.recorder = &MockTemplateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockTemplate) EXPECT() *MockTemplateMockRecorder {
	return m.recorder
}

// GetName mocks base method
func (m *MockTemplate) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName
func (mr *MockTemplateMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockTemplate)(nil).GetName))
}

// GetTemplateType mocks base method
func (m *MockTemplate) GetTemplateType() template.TemplateType {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTemplateType")
	ret0, _ := ret[0].(template.TemplateType)
	return ret0
}

// GetTemplateType indicates an expected call of GetTemplateType
func (mr *MockTemplateMockRecorder) GetTemplateType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTemplateType", reflect.TypeOf((*MockTemplate)(nil).GetTemplateType))
}

// GetDuration mocks base method
func (m *MockTemplate) GetDuration() (time.Duration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDuration")
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDuration indicates an expected call of GetDuration
func (mr *MockTemplateMockRecorder) GetDuration() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDuration", reflect.TypeOf((*MockTemplate)(nil).GetDuration))
}

// GetDeadline mocks base method
func (m *MockTemplate) GetDeadline() (time.Duration, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeadline")
	ret0, _ := ret[0].(time.Duration)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeadline indicates an expected call of GetDeadline
func (mr *MockTemplateMockRecorder) GetDeadline() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeadline", reflect.TypeOf((*MockTemplate)(nil).GetDeadline))
}
