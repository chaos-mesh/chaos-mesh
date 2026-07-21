// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Validate ensures an IstioChaos injects at least one fault.
func (in *IstioFault) Validate(_ interface{}, path *field.Path) field.ErrorList {
	if in.Delay == nil && in.Abort == nil {
		return field.ErrorList{
			field.Required(path, "at least one of delay or abort must be configured"),
		}
	}
	return nil
}
