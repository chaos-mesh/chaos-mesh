// Copyright 2020 Chaos Mesh Authors.
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

package jvm

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	SUID   string = "suid"
	ACTION string = "action"
	TARGET string = "target"
)

func ToSandboxAction(suid string, chaos *v1alpha1.JVMChaos) ([]byte, error) {
	kv := make(map[string]string, 0)

	flags := v1alpha1.JvmSpec[chaos.Spec.Target][chaos.Spec.Action].Flags
	if flags != nil {
		for k, v := range chaos.Spec.Flags {
			for _, rule := range flags {
				if rule.Name != k {
					continue
				}

				if rule.ParameterType != v1alpha1.BoolType {
					kv[k] = v
				} else {
					f, _ := strconv.ParseBool(v)
					if f {
						kv[k] = v
					}
				}
			}
		}
	}

	matchers := v1alpha1.JvmSpec[chaos.Spec.Target][chaos.Spec.Action].Matcher
	if matchers != nil {
		for k, v := range chaos.Spec.Matchers {
			for _, rule := range matchers {
				if rule.Name != k {
					continue
				}

				if rule.ParameterType != v1alpha1.BoolType {
					kv[k] = v
				} else {
					f, _ := strconv.ParseBool(v)
					if f {
						kv[k] = v
					}
				}
			}
		}
	}

	kv[SUID] = suid
	kv[ACTION] = fmt.Sprint(chaos.Spec.Action)
	kv[TARGET] = fmt.Sprint(chaos.Spec.Target)
	return json.Marshal(kv)
}
