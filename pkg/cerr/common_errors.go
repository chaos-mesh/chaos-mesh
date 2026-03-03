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
	"reflect"

	"github.com/pkg/errors"
)

type errHelper struct {
	inner error
}

func FromErr(err error) *errHelper {
	return &errHelper{inner: err}
}

func (h *errHelper) Err() error {
	return h.inner
}

func NotType[expected any]() *errHelper {
	var exp expected
	return &errHelper{inner: errors.Errorf("expected type: %T", exp)}
}

func NotImpl[expected any]() *errHelper {
	var exp *expected
	return &errHelper{inner: errors.Errorf("not implement %v", reflect.TypeOf(exp).Elem())}
}

func NotFoundType[in any]() *errHelper {
	var i in
	return &errHelper{inner: errors.Errorf("not found type: %T", i)}
}

func NotInit[in any]() *errHelper {
	var i in
	return &errHelper{inner: errors.Errorf("not init %T", i)}
}

func NotFound(name string) *errHelper {
	return &errHelper{errors.Errorf("%s not found", name)}
}

func (h *errHelper) WrapInput(in any) *errHelper {
	return &errHelper{inner: errors.Wrapf(h.inner, "input type: %T, input value: %v", in, in)}
}

func (h *errHelper) WrapValue(in any) *errHelper {
	return &errHelper{inner: errors.Wrapf(h.inner, "input value: %v", in)}
}

func (h *errHelper) WrapName(name string) *errHelper {
	return &errHelper{inner: errors.Wrapf(h.inner, "%s", name)}
}

func (h *errHelper) WrapErr(err error) *errHelper {
	return &errHelper{inner: errors.Wrapf(h.inner, "err : %s", err)}
}

func (h errHelper) Wrapf(format string, args ...interface{}) *errHelper {
	return &errHelper{inner: errors.Wrapf(h.inner, format, args...)}
}

func (h *errHelper) WithStack() *errHelper {
	return &errHelper{inner: errors.WithStack(h.inner)}
}

var (
	ErrDuplicateEntity = errors.New("duplicate entity")
)
