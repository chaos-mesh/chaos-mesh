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

package collector

import (
	"encoding/json"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type ValuedCollector struct {
	status v1alpha1.ConditionalBranchesStatus
}

func NewValuedCollector(status v1alpha1.ConditionalBranchesStatus) *ValuedCollector {
	return &ValuedCollector{status: status}
}

func (it *ValuedCollector) CollectContext() (env map[string]interface{}, err error) {
	if len(it.status.Context) == 0 {
		return nil, nil
	}
	result := make(map[string]interface{})
	for _, jsonString := range it.status.Context {
		var tmp map[string]interface{}
		err := json.Unmarshal([]byte(jsonString), &tmp)
		if err != nil {
			return nil, err
		}
		mapExtend(result, tmp)
	}
	return result, err
}
