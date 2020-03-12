// Copyright 2020 PingCAP, Inc.
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

package v1alpha1

import (
	"encoding/json"
	"fmt"
)

// +kubebuilder:object:generate=false

// Validator describes the interface should be implemented in validation webhook
type Validator interface {
	// Validate describe the interface should be implemented in validation webhook
	Validate() (bool, string, error)
}

func ParseValidator(data []byte, kind string) (Validator, error) {
	var validator Validator

	switch kind {
	case "PodChaos":
		validator = &PodChaos{}
	case "NetworkChaos":
		validator = &NetworkChaos{}
	case "IoChaos":
		validator = &IoChaos{}
	case "TimeChaos":
		validator = &TimeChaos{}
	case "KernelChaos":
		validator = &KernelChaos{}
	default:
		return nil, fmt.Errorf("%s validator is not supported", kind)
	}

	if err := json.Unmarshal(data, validator); err != nil {
		return nil, err
	}

	return validator, nil
}
