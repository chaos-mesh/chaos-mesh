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

package tproxyconfig

import (
	"encoding/json"
	"testing"
)

func TestPodHttpChaosPatchBody_UnmarshalJSON_Valid(t *testing.T) {
	raw := `{"type":"JSON","value":"patch-value"}`
	var b PodHttpChaosPatchBody
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Contents.Type != "JSON" {
		t.Errorf("expected Type=JSON, got %q", b.Contents.Type)
	}
	if b.Contents.Value != "patch-value" {
		t.Errorf("expected Value=patch-value, got %q", b.Contents.Value)
	}
}

func TestPodHttpChaosPatchBody_UnmarshalJSON_Invalid(t *testing.T) {
	raw := `not-json`
	var b PodHttpChaosPatchBody
	if err := json.Unmarshal([]byte(raw), &b); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestPodHttpChaosPatchBody_UnmarshalJSON_EmptyValue(t *testing.T) {
	raw := `{"type":"JSON","value":""}`
	var b PodHttpChaosPatchBody
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Contents.Type != "JSON" {
		t.Errorf("expected Type=JSON, got %q", b.Contents.Type)
	}
	if b.Contents.Value != "" {
		t.Errorf("expected empty Value, got %q", b.Contents.Value)
	}
}

func TestPodHttpChaosReplaceBody_UnmarshalJSON_AsStruct(t *testing.T) {
	raw := `{"type":"JSON","value":"replace-value"}`
	var b PodHttpChaosReplaceBody
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Contents.Type != "JSON" {
		t.Errorf("expected Type=JSON, got %q", b.Contents.Type)
	}
	if b.Contents.Value != "replace-value" {
		t.Errorf("expected Value=replace-value, got %q", b.Contents.Value)
	}
}

func TestPodHttpChaosReplaceBody_UnmarshalJSON_FallbackToText(t *testing.T) {
	raw := `"aGVsbG8="`
	var b PodHttpChaosReplaceBody
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Contents.Type != "TEXT" {
		t.Errorf("expected Type=TEXT, got %q", b.Contents.Type)
	}
	if b.Contents.Value != "hello" {
		t.Errorf("expected Value=hello, got %q", b.Contents.Value)
	}
}

func TestPodHttpChaosReplaceBody_UnmarshalJSON_Invalid(t *testing.T) {
	raw := `not-json`
	var b PodHttpChaosReplaceBody
	if err := json.Unmarshal([]byte(raw), &b); err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}
