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
	// The relocation of a X86 image is like:
	// Relocation section '.rela.text' at offset 0x288 contains 3 entries:
	// Offset          Info           Type           Sym. Value    Sym. Name + Addend
	// 000000000016  000900000002 R_X86_64_PC32     0000000000000000 CLOCK_IDS_MASK - 4
	// 00000000001f  000a00000002 R_X86_64_PC32     0000000000000008 TV_NSEC_DELTA - 4
	// 00000000002a  000b00000002 R_X86_64_PC32     0000000000000010 TV_SEC_DELTA - 4
	//
	// For example, we need to write the offset of `CLOCK_IDS_MASK` - 4 in 0x16 of the section
	// If we want to put the `CLOCK_IDS_MASK` at the end of the section, it will be
	// len(imageContent) - 4 - 0x16

	imageOffset[sym.Name] = len(*imageContent)
	targetOffset := uint32(len(*imageContent)) - uint32(rela.Off) + uint32(rela.Addend)
	byteorder.PutUint32((*imageContent)[rela.Off:rela.Off+4], targetOffset)

	// TODO: support other length besides uint64 (which is 8 bytes)
	*imageContent = append(*imageContent, make([]byte, varLength)...)
}
