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

package cerr

import (
	"errors"
	"strings"
	"testing"
)

// TestFromErr_Err verifies that FromErr wraps the provided error and Err()
// returns it unchanged.
func TestFromErr_Err(t *testing.T) {
	original := errors.New("original error")
	h := FromErr(original)
	if h.Err() != original {
		t.Errorf("expected Err() to return the original error, got %v", h.Err())
	}
}

// TestNotType verifies that NotType produces an error message containing the
// expected type name.
func TestNotType(t *testing.T) {
	h := NotType[int]()
	msg := h.Err().Error()
	if !strings.Contains(msg, "int") {
		t.Errorf("expected error message to contain 'int', got %q", msg)
	}
}

// TestNotImpl verifies that NotImpl produces an error message containing the
// expected interface name.
func TestNotImpl(t *testing.T) {
	type MyInterface interface{ Foo() }
	h := NotImpl[MyInterface]()
	msg := h.Err().Error()
	if !strings.Contains(msg, "MyInterface") {
		t.Errorf("expected error message to contain 'MyInterface', got %q", msg)
	}
}

// TestNotFoundType verifies that NotFoundType produces an error message
// containing the expected type name.
func TestNotFoundType(t *testing.T) {
	h := NotFoundType[string]()
	msg := h.Err().Error()
	if !strings.Contains(msg, "string") {
		t.Errorf("expected error message to contain 'string', got %q", msg)
	}
}

// TestNotInit verifies that NotInit produces an error message containing the
// expected type name.
func TestNotInit(t *testing.T) {
	h := NotInit[float64]()
	msg := h.Err().Error()
	if !strings.Contains(msg, "float64") {
		t.Errorf("expected error message to contain 'float64', got %q", msg)
	}
}

// TestNotFound verifies that NotFound produces an error message containing the
// provided name.
func TestNotFound(t *testing.T) {
	h := NotFound("myresource")
	msg := h.Err().Error()
	if !strings.Contains(msg, "myresource") {
		t.Errorf("expected error message to contain 'myresource', got %q", msg)
	}
	if !strings.Contains(msg, "not found") {
		t.Errorf("expected error message to contain 'not found', got %q", msg)
	}
}

// TestWrapInput verifies that WrapInput includes both the type and value of the
// input in the wrapped error message.
func TestWrapInput(t *testing.T) {
	base := NotFound("x")
	h := base.WrapInput(42)
	msg := h.Err().Error()
	if !strings.Contains(msg, "int") {
		t.Errorf("expected wrapped message to contain input type 'int', got %q", msg)
	}
	if !strings.Contains(msg, "42") {
		t.Errorf("expected wrapped message to contain input value '42', got %q", msg)
	}
}

// TestWrapValue verifies that WrapValue includes the value of the input in the
// wrapped error message.
func TestWrapValue(t *testing.T) {
	base := NotFound("y")
	h := base.WrapValue("hello")
	msg := h.Err().Error()
	if !strings.Contains(msg, "hello") {
		t.Errorf("expected wrapped message to contain value 'hello', got %q", msg)
	}
}

// TestWrapName verifies that WrapName includes the provided name in the
// wrapped error message.
func TestWrapName(t *testing.T) {
	base := NotFound("z")
	h := base.WrapName("context-name")
	msg := h.Err().Error()
	if !strings.Contains(msg, "context-name") {
		t.Errorf("expected wrapped message to contain 'context-name', got %q", msg)
	}
}

// TestWrapErr verifies that WrapErr includes the cause error's message in the
// wrapped error message.
func TestWrapErr(t *testing.T) {
	base := NotFound("a")
	cause := errors.New("cause error")
	h := base.WrapErr(cause)
	msg := h.Err().Error()
	if !strings.Contains(msg, "cause error") {
		t.Errorf("expected wrapped message to contain 'cause error', got %q", msg)
	}
}

// TestWrapf verifies that Wrapf formats the additional context correctly.
func TestWrapf(t *testing.T) {
	base := NotFound("b")
	h := base.Wrapf("context: %s=%d", "key", 99)
	msg := h.Err().Error()
	if !strings.Contains(msg, "key=99") {
		t.Errorf("expected wrapped message to contain 'key=99', got %q", msg)
	}
}

// TestWithStack verifies that WithStack wraps the error with a stack trace
// (the error itself is preserved).
func TestWithStack(t *testing.T) {
	base := NotFound("c")
	h := base.WithStack()
	if h.Err() == nil {
		t.Error("expected WithStack() to return a non-nil error")
	}
	if !strings.Contains(h.Err().Error(), "c") {
		t.Errorf("expected WithStack error to preserve original message, got %q", h.Err().Error())
	}
}

// TestErrDuplicateEntity verifies that the package-level sentinel error is
// non-nil and has a meaningful message.
func TestErrDuplicateEntity(t *testing.T) {
	if ErrDuplicateEntity == nil {
		t.Fatal("expected ErrDuplicateEntity to be non-nil")
	}
	if !strings.Contains(ErrDuplicateEntity.Error(), "duplicate") {
		t.Errorf("expected ErrDuplicateEntity message to contain 'duplicate', got %q", ErrDuplicateEntity.Error())
	}
}

// TestChaining verifies that multiple Wrap calls can be chained and the final
// error message contains all intermediate context strings.
func TestChaining(t *testing.T) {
	h := NotFound("resource").
		WrapName("step-one").
		WrapValue("val").
		Wrapf("step-three key=%s", "abc")

	msg := h.Err().Error()
	for _, want := range []string{"resource", "step-one", "val", "abc"} {
		if !strings.Contains(msg, want) {
			t.Errorf("expected chained message to contain %q, got %q", want, msg)
		}
	}
}
