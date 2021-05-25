// Copyright 2021 Chaos Mesh Authors.
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

package recorder

import (
	"fmt"
)

type Failed struct {
	Activity string

	Err string
}

func (f Failed) Type() string {
	return "Warning"
}

func (f Failed) Reason() string {
	return "Failed"
}

func (f Failed) Message() string {
	return fmt.Sprintf("Failed to %s: %s", f.Activity, f.Err)
}

func init() {
	register(Failed{})
}
