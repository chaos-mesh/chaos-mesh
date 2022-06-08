// Copyright 2022 Chaos Mesh Authors.
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
	"debug/elf"
	"encoding/binary"
)

func AssetLD(rela elf.Rela64, imageOffset map[string]int, imageContent *[]byte, sym elf.Symbol, byteorder binary.ByteOrder) {
	imageOffset[sym.Name] = len(*imageContent)

	targetOffset := uint32(len(*imageContent)) - uint32(rela.Off) + uint32(rela.Addend)

	// The relocation of a aarch64 image is like:
	// Offset          Info           Type           Sym. Value    Sym. Name + Addend
	// 000000000010  000b00000135 R_AARCH64_GOT_LD_ 0000000000000000 CLOCK_IDS_MASK + 0
	// 00000000002c  000c00000135 R_AARCH64_GOT_LD_ 0000000000000000 TV_NSEC_DELTA + 0
	// 000000000034  000d00000135 R_AARCH64_GOT_LD_ 0000000000000000 TV_SEC_DELTA + 0

	// we assume the type is always R_AARCH64_GOT_LD_PREL19, with `-mcmodel=tiny`

	// In this situation, we need to push two uint64 to the end:
	// One for the location of variable, and one for the variable

	// For example, if the entry starts at 0x00, and we have two variables whose value are
	// 0xFF and 0xFE. We will have 32 bytes after the content:
	// | 0x00 | 0x08 | 0x10 | 0x18 |
	// | 0x08 | 0xFF | 0x18 | 0xFE |

	// See the manual of LDR (literal) and LDR (register) to understand the
	// relocation based on the following assemblies:
	//
	// ldr x1, #OFFSET_OF_ADDR ; in this step, the address of variable is loaded
	//                           into the x1 register
	// ldr x1, [x1]            ; in this step, the variable itself is loaded into
	//                           the register

	targetOffset >>= 2
	instr := byteorder.Uint32((*imageContent)[rela.Off : rela.Off+4])

	// See the document of instruction
	// [ldr](https://developer.arm.com/documentation/ddi0596/2021-12/Base-Instructions/LDR--literal---Load-Register--literal--?lang=en)
	// the offset is saved in `imm19`, and the max length is 19 (bit)
	//
	// 1. cut `instr` at [0:4] and [23:31]
	// 2. cut the little 19 bit of `targetOffset`, and shift it to [5:23]
	// 3. concat them
	instr = uint32(int(instr) & ^((1<<19-1)<<5)) | ((targetOffset & (1<<19 - 1)) << 5)
	byteorder.PutUint32((*imageContent)[rela.Off:rela.Off+4], instr)

	// TODO: support other length besides uint64 (which is 8 bytes)
	*imageContent = append(*imageContent, make([]byte, varLength)...)
}
