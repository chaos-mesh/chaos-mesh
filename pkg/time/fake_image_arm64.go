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

package time

import (
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/mapreader"
	"github.com/chaos-mesh/chaos-mesh/pkg/ptrace"
)

// one variable will use two pointers place
const varLength = 16

func (it *FakeImage) SetVarUint64(program *ptrace.TracedProgram, entry *mapreader.Entry, symbol string, value uint64) error {
	if offset, ok := it.offset[symbol]; ok {
		variableOffset := entry.StartAddress + uint64(offset) + 8

		err := program.WriteUint64ToAddr(entry.StartAddress+uint64(offset), variableOffset)
		if err != nil {
			return err
		}

		err = program.WriteUint64ToAddr(variableOffset, value)
		return err
	}

	return errors.New("symbol not found")
}
