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

// ToSandboxAction convertes chaos to sandbox action
func ToSandboxAction(suid string, chaos *v1alpha1.JVMChaos) ([]byte, error) {
	if _, ok := v1alpha1.JvmSpec[chaos.Spec.Target]; !ok {
		return nil, fmt.Errorf("unknown JVM chaos target:%s",
			chaos.Spec.Target)
	}

	if _, ok := v1alpha1.JvmSpec[chaos.Spec.Target][chaos.Spec.Action]; !ok {
		return nil, fmt.Errorf("JVM target: %s does not supported action: %s",
			chaos.Spec.Target, chaos.Spec.Action)
	}

	kv := make(map[string]string)
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
					f, err := strconv.ParseBool(v)
					if err != nil {
						return nil, fmt.Errorf("can not parse Spec.Flags.%s's value:%s as boolean", k, v)
					}
					// if f is false, should not send key-value to sandbox server.
					if f {
						kv[k] = v
					}
				}
			}
		}
	}

	matchers := v1alpha1.JvmSpec[chaos.Spec.Target][chaos.Spec.Action].Matchers
	if matchers != nil {
		for k, v := range chaos.Spec.Matchers {
			for _, rule := range matchers {
				if rule.Name != k {
					continue
				}

				if rule.ParameterType != v1alpha1.BoolType {
					kv[k] = v
				} else {
					f, err := strconv.ParseBool(v)
					if err != nil {
						return nil, fmt.Errorf("can not parse Spec.Matchers.%s's value:%s as boolean", k, v)
					}
					// if f is false, should not send key-value to sandbox server.
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
