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

package v1alpha1

import "testing"

func TestPhysicalMachineProcessIsOneShot(t *testing.T) {
	cases := []struct {
		name    string
		action  PhysicalMachineChaosAction
		process *ProcessSpec
		want    bool
	}{
		{"kill", PMProcessAction, &ProcessSpec{Signal: 9}, true},
		{"terminate", PMProcessAction, &ProcessSpec{Signal: 15}, true},
		{"other signal", PMProcessAction, &ProcessSpec{Signal: 10}, true},
		{"continue", PMProcessAction, &ProcessSpec{Signal: 18}, true},
		{"linux stop", PMProcessAction, &ProcessSpec{Signal: 19}, false},
		{"kill with recovery", PMProcessAction, &ProcessSpec{Signal: 9, RecoverCmd: "start-service"}, false},
		{"terminate with recovery", PMProcessAction, &ProcessSpec{Signal: 15, RecoverCmd: "start-service"}, false},
		{"stop with recovery", PMProcessAction, &ProcessSpec{Signal: 19, RecoverCmd: "start-service"}, false},
		{"whitespace recovery", PMProcessAction, &ProcessSpec{Signal: 9, RecoverCmd: " "}, false},
		{"zero signal", PMProcessAction, &ProcessSpec{}, false},
		{"missing process", PMProcessAction, nil, false},
		{"other action", PMStressCPUAction, &ProcessSpec{Signal: 9}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &PhysicalMachineChaos{Spec: PhysicalMachineChaosSpec{Action: tc.action, ExpInfo: ExpInfo{Process: tc.process}}}
			if got := obj.IsOneShot(); got != tc.want {
				t.Fatalf("IsOneShot() = %v, want %v", got, tc.want)
			}
		})
	}
}
