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

type Applied struct {
	Id string
}

func (a Applied) Type() string {
	return "Normal"
}

func (a Applied) Reason() string {
	return "Applied"
}

func (a Applied) Message() string {
	return fmt.Sprintf("Successfully apply chaos for %s", a.Id)
}

type Recovered struct {
	Id string
}

func (r Recovered) Type() string {
	return "Normal"
}

func (r Recovered) Reason() string {
	return "Recovered"
}

func (r Recovered) Message() string {
	return fmt.Sprintf("Successfully recover chaos for %s", r.Id)
}

type NotSupported struct {
	Activity string
}

func (r NotSupported) Type() string {
	return "Warning"
}

func (r NotSupported) Reason() string {
	return "NotSupported"
}

func (r NotSupported) Message() string {
	return fmt.Sprintf("%s is not supported", r.Activity)
}

func init() {
	register(Applied{}, Recovered{}, NotSupported{})
}
