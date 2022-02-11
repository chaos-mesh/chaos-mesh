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

package chaoserr

import (
	"github.com/pkg/errors"
)

type ErrNotFound struct {
	Name string
}

func (e ErrNotFound) Error() string {
	return e.Name + " not found"
}

func NotFound(name string) error {
	return ErrNotFound{Name: name}
}

type ErrNotImplemented struct {
	Name string
}

func (e ErrNotImplemented) Error() string {
	return e.Name + " not implement"
}

func NotImplemented(name string) error {
	return ErrNotImplemented{Name: name}
}

var (
	ErrDuplicateEntity = errors.New("duplicate entity")
)
