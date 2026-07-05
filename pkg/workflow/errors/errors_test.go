// Copyright 2026 Chaos Mesh Authors.
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

package errors

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---------- WorkflowError ----------

func TestWorkflowError_Error(t *testing.T) {
	err := New("something went wrong")
	if err.Error() != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", err.Error())
	}
}

func TestSentinelErrors(t *testing.T) {
	cases := []struct {
		name string
		err  *WorkflowError
		want string
	}{
		{"ErrNoSuchNode", ErrNoSuchNode, "no such node"},
		{"ErrNoSuchTemplate", ErrNoSuchTemplate, "no such template"},
		{"ErrParseTemplateFailed", ErrParseTemplateFailed, "failed to parse certain type of template"},
		{"ErrNoMoreTemplateInSerialTemplate", ErrNoMoreTemplateInSerialTemplate, "no more template could schedule in serial template"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == nil {
				t.Fatal("expected non-nil sentinel error")
			}
			if tc.err.Error() != tc.want {
				t.Errorf("expected %q, got %q", tc.want, tc.err.Error())
			}
		})
	}
}

// ---------- NoSuchTemplateError ----------

func TestNoSuchTemplateError_Fields(t *testing.T) {
	err := NewNoSuchTemplateError("get", "my-workflow", "my-template")
	if err.Op != "get" {
		t.Errorf("expected Op=get, got %q", err.Op)
	}
	if err.WorkflowName != "my-workflow" {
		t.Errorf("expected WorkflowName=my-workflow, got %q", err.WorkflowName)
	}
	if err.TemplateName != "my-template" {
		t.Errorf("expected TemplateName=my-template, got %q", err.TemplateName)
	}
	if err.Unwrap() != ErrNoSuchTemplate {
		t.Errorf("expected Unwrap to return ErrNoSuchTemplate")
	}
}

func TestNoSuchTemplateError_ErrorIsJSON(t *testing.T) {
	err := NewNoSuchTemplateError("get", "my-workflow", "my-template")
	msg := err.Error()
	if !json.Valid([]byte(msg)) {
		t.Errorf("expected Error() to return valid JSON, got %q", msg)
	}
}

func TestNewNoSuchTemplateErrorInTemplates(t *testing.T) {
	available := []string{"tmpl-a", "tmpl-b"}
	err := NewNoSuchTemplateErrorInTemplates("list", "missing-tmpl", available)
	if err.TemplateName != "missing-tmpl" {
		t.Errorf("expected TemplateName=missing-tmpl, got %q", err.TemplateName)
	}
	if len(err.AllAvailableTemplates) != 2 {
		t.Errorf("expected 2 available templates, got %d", len(err.AllAvailableTemplates))
	}
}

// ---------- NoSuchTreeNodeError ----------

func TestNoSuchTreeNodeError_Fields(t *testing.T) {
	err := NewNoSuchTreeNodeError("find", "parent-node", "my-workflow")
	if err.Op != "find" {
		t.Errorf("expected Op=find, got %q", err.Op)
	}
	if err.ParentNodeName != "parent-node" {
		t.Errorf("expected ParentNodeName=parent-node, got %q", err.ParentNodeName)
	}
	if err.WorkflowName != "my-workflow" {
		t.Errorf("expected WorkflowName=my-workflow, got %q", err.WorkflowName)
	}
	if err.Unwrap() != ErrNoSuchNode {
		t.Errorf("expected Unwrap to return ErrNoSuchNode")
	}
}

func TestNoSuchTreeNodeError_ErrorIsJSON(t *testing.T) {
	err := NewNoSuchTreeNodeError("find", "parent-node", "my-workflow")
	msg := err.Error()
	if !json.Valid([]byte(msg)) {
		t.Errorf("expected Error() to return valid JSON, got %q", msg)
	}
}

// ---------- ParseTemplateFailedError ----------

func TestParseTemplateFailedError_Fields(t *testing.T) {
	err := NewParseSerialTemplateFailedError("parse", "some-raw-value")
	if err.Op != "parse" {
		t.Errorf("expected Op=parse, got %q", err.Op)
	}
	if err.Unwrap() != ErrParseTemplateFailed {
		t.Errorf("expected Unwrap to return ErrParseTemplateFailed")
	}
	if !strings.Contains(err.RawType, "string") {
		t.Errorf("expected RawType to contain 'string', got %q", err.RawType)
	}
}

func TestParseTemplateFailedError_ErrorIsJSON(t *testing.T) {
	err := NewParseSerialTemplateFailedError("parse", "some-raw-value")
	msg := err.Error()
	if !json.Valid([]byte(msg)) {
		t.Errorf("expected Error() to return valid JSON, got %q", msg)
	}
}

// ---------- NoMoreTemplateInSerialTemplateError ----------

func TestNoMoreTemplateInSerialTemplateError_Fields(t *testing.T) {
	err := NewNoMoreTemplateInSerialTemplateError("schedule", "my-workflow", "serial-tmpl", "node-1")
	if err.Op != "schedule" {
		t.Errorf("expected Op=schedule, got %q", err.Op)
	}
	if err.WorkflowName != "my-workflow" {
		t.Errorf("expected WorkflowName=my-workflow, got %q", err.WorkflowName)
	}
	if err.TemplateName != "serial-tmpl" {
		t.Errorf("expected TemplateName=serial-tmpl, got %q", err.TemplateName)
	}
	if err.NodeName != "node-1" {
		t.Errorf("expected NodeName=node-1, got %q", err.NodeName)
	}
	if err.Unwrap() != ErrNoMoreTemplateInSerialTemplate {
		t.Errorf("expected Unwrap to return ErrNoMoreTemplateInSerialTemplate")
	}
}

func TestNoMoreTemplateInSerialTemplateError_ErrorIsJSON(t *testing.T) {
	err := NewNoMoreTemplateInSerialTemplateError("schedule", "my-workflow", "serial-tmpl", "node-1")
	msg := err.Error()
	if !json.Valid([]byte(msg)) {
		t.Errorf("expected Error() to return valid JSON, got %q", msg)
	}
}
