// Copyright 2021 Chaos Mesh Authors.
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

package flags

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// MapStringStringFlag is a flag struct for key=value pairs
type MapStringStringFlag struct {
	Values map[string]string
}

// String implements the flag.Var interface
func (s *MapStringStringFlag) String() string {
	z := []string{}
	for x, y := range s.Values {
		z = append(z, fmt.Sprintf("%s=%s", x, y))
	}
	return strings.Join(z, ",")
}

// Set implements the flag.Var interface
func (s *MapStringStringFlag) Set(value string) error {
	if s.Values == nil {
		s.Values = map[string]string{}
	}
	for _, p := range strings.Split(value, ",") {
		fields := strings.Split(p, "=")
		if len(fields) != 2 {
			return errors.Errorf("%s is incorrectly formatted! should be key=value[,key2=value2]", p)
		}
		s.Values[fields[0]] = fields[1]
	}
	return nil
}

// ToMapStringString returns the underlying representation of the map of key=value pairs
func (s *MapStringStringFlag) ToMapStringString() map[string]string {
	return s.Values
}

// NewMapStringStringFlag creates a new flag var for storing key=value pairs
func NewMapStringStringFlag() MapStringStringFlag {
	return MapStringStringFlag{Values: map[string]string{}}
}
