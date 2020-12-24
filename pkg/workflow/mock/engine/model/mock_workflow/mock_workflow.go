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
// Source: ./engine/model/workflow/workflow.go

// Package mock_workflow is a generated GoMock package.
package mock_workflow

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"

	node "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	template "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
	workflow "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/workflow"
)

// MockWorkflowSpec is a mock of WorkflowSpec interface
type MockWorkflowSpec struct {
	ctrl     *gomock.Controller
	recorder *MockWorkflowSpecMockRecorder
}

// MockWorkflowSpecMockRecorder is the mock recorder for MockWorkflowSpec
type MockWorkflowSpecMockRecorder struct {
	mock *MockWorkflowSpec
}

// NewMockWorkflowSpec creates a new mock instance
func NewMockWorkflowSpec(ctrl *gomock.Controller) *MockWorkflowSpec {
	mock := &MockWorkflowSpec{ctrl: ctrl}
	mock.recorder = &MockWorkflowSpecMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkflowSpec) EXPECT() *MockWorkflowSpecMockRecorder {
	return m.recorder
}

// GetName mocks base method
func (m *MockWorkflowSpec) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName
func (mr *MockWorkflowSpecMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MockWorkflowSpec)(nil).GetName))
}

// GetEntry mocks base method
func (m *MockWorkflowSpec) GetEntry() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEntry")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetEntry indicates an expected call of GetEntry
func (mr *MockWorkflowSpecMockRecorder) GetEntry() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEntry", reflect.TypeOf((*MockWorkflowSpec)(nil).GetEntry))
}

// GetTemplates mocks base method
func (m *MockWorkflowSpec) GetTemplates() (template.Templates, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTemplates")
	ret0, _ := ret[0].(template.Templates)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTemplates indicates an expected call of GetTemplates
func (mr *MockWorkflowSpecMockRecorder) GetTemplates() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTemplates", reflect.TypeOf((*MockWorkflowSpec)(nil).GetTemplates))
}

// FetchTemplateByName mocks base method
func (m *MockWorkflowSpec) FetchTemplateByName(templateName string) (template.Template, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchTemplateByName", templateName)
	ret0, _ := ret[0].(template.Template)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchTemplateByName indicates an expected call of FetchTemplateByName
func (mr *MockWorkflowSpecMockRecorder) FetchTemplateByName(templateName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchTemplateByName", reflect.TypeOf((*MockWorkflowSpec)(nil).FetchTemplateByName), templateName)
}

// MockWorkflowStatus is a mock of WorkflowStatus interface
type MockWorkflowStatus struct {
	ctrl     *gomock.Controller
	recorder *MockWorkflowStatusMockRecorder
}

// MockWorkflowStatusMockRecorder is the mock recorder for MockWorkflowStatus
type MockWorkflowStatusMockRecorder struct {
	mock *MockWorkflowStatus
}

// NewMockWorkflowStatus creates a new mock instance
func NewMockWorkflowStatus(ctrl *gomock.Controller) *MockWorkflowStatus {
	mock := &MockWorkflowStatus{ctrl: ctrl}
	mock.recorder = &MockWorkflowStatusMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockWorkflowStatus) EXPECT() *MockWorkflowStatusMockRecorder {
	return m.recorder
}

// GetPhase mocks base method
func (m *MockWorkflowStatus) GetPhase() workflow.WorkflowPhase {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPhase")
	ret0, _ := ret[0].(workflow.WorkflowPhase)
	return ret0
}

// GetPhase indicates an expected call of GetPhase
func (mr *MockWorkflowStatusMockRecorder) GetPhase() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPhase", reflect.TypeOf((*MockWorkflowStatus)(nil).GetPhase))
}

// GetNodes mocks base method
func (m *MockWorkflowStatus) GetNodes() []node.Node {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodes")
	ret0, _ := ret[0].([]node.Node)
	return ret0
}

// GetNodes indicates an expected call of GetNodes
func (mr *MockWorkflowStatusMockRecorder) GetNodes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodes", reflect.TypeOf((*MockWorkflowStatus)(nil).GetNodes))
}

// GetWorkflowSpecName mocks base method
func (m *MockWorkflowStatus) GetWorkflowSpecName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWorkflowSpecName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetWorkflowSpecName indicates an expected call of GetWorkflowSpecName
func (mr *MockWorkflowStatusMockRecorder) GetWorkflowSpecName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWorkflowSpecName", reflect.TypeOf((*MockWorkflowStatus)(nil).GetWorkflowSpecName))
}

// GetNodesTree mocks base method
func (m *MockWorkflowStatus) GetNodesTree() node.NodeTreeNode {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodesTree")
	ret0, _ := ret[0].(node.NodeTreeNode)
	return ret0
}

// GetNodesTree indicates an expected call of GetNodesTree
func (mr *MockWorkflowStatusMockRecorder) GetNodesTree() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodesTree", reflect.TypeOf((*MockWorkflowStatus)(nil).GetNodesTree))
}

// FetchNodesMap mocks base method
func (m *MockWorkflowStatus) FetchNodesMap() map[string]node.Node {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchNodesMap")
	ret0, _ := ret[0].(map[string]node.Node)
	return ret0
}

// FetchNodesMap indicates an expected call of FetchNodesMap
func (mr *MockWorkflowStatusMockRecorder) FetchNodesMap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchNodesMap", reflect.TypeOf((*MockWorkflowStatus)(nil).FetchNodesMap))
}

// FetchNodeByName mocks base method
func (m *MockWorkflowStatus) FetchNodeByName(nodeName string) (node.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchNodeByName", nodeName)
	ret0, _ := ret[0].(node.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchNodeByName indicates an expected call of FetchNodeByName
func (mr *MockWorkflowStatusMockRecorder) FetchNodeByName(nodeName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchNodeByName", reflect.TypeOf((*MockWorkflowStatus)(nil).FetchNodeByName), nodeName)
}
