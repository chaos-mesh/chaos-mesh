// Copyright 2022 Chaos Mesh Authors.
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

package cerr

import (
	"errors"
	"strings"
	"testing"
)

func TestFromErr(t *testing.T) {
	err := errors.New("original error")
	h := FromErr(err)
	if h.Err() != err {
		t.Errorf("expected original error, got %v", h.Err())
	}
}

func TestFromErr_Nil(t *testing.T) {
	h := FromErr(nil)
	if h.Err() != nil {
		t.Errorf("expected nil, got %v", h.Err())
	}
}

func TestNotType(t *testing.T) {
	h := NotType[string]()
	if h.Err() == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(h.Err().Error(), "string") {
		t.Errorf("expected error to mention string type, got %q", h.Err().Error())
	}
}

func TestNotImpl(t *testing.T) {
	h := NotImpl[string]()
	if h.Err() == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(h.Err().Error(), "not implement") {
		t.Errorf("expected error to contain 'not implement', got %q", h.Err().Error())
	}
}

func TestNotFoundType(t *testing.T) {
	h := NotFoundType[int]()
	if h.Err() == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(h.Err().Error(), "not found type") {
		t.Errorf("expected error to contain 'not found type', got %q", h.Err().Error())
	}
}

func TestNotInit(t *testing.T) {
	h := NotInit[string]()
	if h.Err() == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(h.Err().Error(), "not init") {
		t.Errorf("expected error to contain 'not init', got %q", h.Err().Error())
	}
}

func TestNotFound(t *testing.T) {
	h := NotFound("myresource")
	if h.Err() == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(h.Err().Error(), "myresource not found") {
		t.Errorf("expected error to contain 'myresource not found', got %q", h.Err().Error())
	}
}

func TestWrapInput(t *testing.T) {
	h := FromErr(errors.New("base")).WrapInput("testvalue")
	msg := h.Err().Error()
	if !strings.Contains(msg, "testvalue") {
		t.Errorf("expected wrapped message to contain input value, got %q", msg)
	}
}

func TestWrapValue(t *testing.T) {
	h := FromErr(errors.New("base")).WrapValue(42)
	msg := h.Err().Error()
	if !strings.Contains(msg, "42") {
		t.Errorf("expected wrapped message to contain value 42, got %q", msg)
	}
}

func TestWrapName(t *testing.T) {
	h := FromErr(errors.New("base")).WrapName("myname")
	msg := h.Err().Error()
	if !strings.Contains(msg, "myname") {
		t.Errorf("expected wrapped message to contain 'myname', got %q", msg)
	}
}

func TestWrapErr(t *testing.T) {
	inner := errors.New("inner error")
	h := FromErr(errors.New("base")).WrapErr(inner)
	msg := h.Err().Error()
	if !strings.Contains(msg, "inner error") {
		t.Errorf("expected wrapped message to contain 'inner error', got %q", msg)
	}
}

func TestWrapf(t *testing.T) {
	h := FromErr(errors.New("base")).Wrapf("context %s", "value")
	msg := h.Err().Error()
	if !strings.Contains(msg, "context value") {
		t.Errorf("expected wrapped message to contain 'context value', got %q", msg)
	}
}

func TestWithStack(t *testing.T) {
	h := FromErr(errors.New("base")).WithStack()
	if h.Err() == nil {
		t.Error("expected non-nil error after WithStack")
	}
}

func TestChaining(t *testing.T) {
	h := FromErr(errors.New("root")).WrapName("step1").WrapValue("step2").Wrapf("step3 %d", 3)
	msg := h.Err().Error()
	if !strings.Contains(msg, "root") {
		t.Errorf("expected chained error to contain 'root', got %q", msg)
	}
}

func TestErrDuplicateEntity(t *testing.T) {
	if ErrDuplicateEntity == nil {
		t.Error("expected ErrDuplicateEntity to be non-nil")
	}
	if !strings.Contains(ErrDuplicateEntity.Error(), "duplicate entity") {
		t.Errorf("expected 'duplicate entity', got %q", ErrDuplicateEntity.Error())
	}
}
