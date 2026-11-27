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

// Package cloud encodes cloud selectors as target IDs for experiment records.
package cloud

import (
	"encoding/json"
	"fmt"
)

type Target string

//revive:disable-next-line:var-naming Preserve the selector.Target identity contract.
func (t Target) Id() string {
	return string(t)
}

func Select(spec any) ([]Target, error) {
	id, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("encode cloud selector %T: %w", spec, err)
	}
	return []Target{Target(id)}, nil
}
