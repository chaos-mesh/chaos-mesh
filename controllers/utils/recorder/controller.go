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
	"strings"
)

type Updated struct {
	Field string
}

func (u Updated) Type() string {
	return "Normal"
}

func (u Updated) Reason() string {
	return "Updated"
}

func (u Updated) Message() string {
	return fmt.Sprintf("Successfully update %s of resource", u.Field)
}

func (u Updated) Parse(message string) ChaosEvent {
	prefix := "Successfully update "
	suffix := " of resource"
	if strings.HasPrefix(message, prefix) && strings.HasSuffix(message, suffix) {
		return Updated{
			Field: strings.TrimSuffix(strings.TrimPrefix(message, prefix), suffix),
		}
	}

	return nil
}

func init() {
	register(Updated{})
}
