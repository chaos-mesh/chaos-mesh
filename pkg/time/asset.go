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
	"bytes"
	"debug/elf"
	"embed"
	"encoding/binary"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

//go:embed fakeclock/*.o
var fakeclock embed.FS

const textSection = ".text"
const relocationSection = ".rela.text"

// LoadFakeImageFromEmbedFs builds FakeImage from the embed filesystem. It parses the ELF file and extract the variables from the relocation section, reserves the space for them at the end of content, then calculates and saves offsets as "manually relocation"
func LoadFakeImageFromEmbedFs(filename string, symbolName string, logger logr.Logger) (*FakeImage, error) {
	path := "fakeclock/" + filename
	object, err := fakeclock.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "read file from embedded fs %s", path)
	}

	elfFile, err := elf.NewFile(bytes.NewReader(object))
	if err != nil {
		return nil, errors.Wrapf(err, "parse elf file %s", path)
	}

	syms, err := elfFile.Symbols()
	if err != nil {
		return nil, errors.Wrapf(err, "get symbols %s", path)
	}

	var imageContent []byte
	imageOffset := make(map[string]int)

	for _, r := range elfFile.Sections {

		if r.Type == elf.SHT_PROGBITS && r.Name == textSection {
			imageContent, err = r.Data()
			if err != nil {
				return nil, errors.Wrapf(err, "read text section data %s", path)
			}
			break
		}
	}

	for _, r := range elfFile.Sections {
		if r.Type == elf.SHT_RELA && r.Name == relocationSection {
			rela_section, err := r.Data()
			if err != nil {
				return nil, errors.Wrapf(err, "read rela section data %s", path)
			}
			rela_section_reader := bytes.NewReader(rela_section)

			var rela elf.Rela64
			for rela_section_reader.Len() > 0 {
				if elfFile.Machine == elf.EM_X86_64 {
					err := binary.Read(rela_section_reader, elfFile.ByteOrder, &rela)
					if err != nil {
						return nil, errors.Wrapf(err, "read rela section rela64 entry %s", path)
					}

					symNo := rela.Info >> 32
					if symNo == 0 || symNo > uint64(len(syms)) {
						continue
					}

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

					sym := &syms[symNo-1]
					imageOffset[sym.Name] = len(imageContent)
					targetOffset := uint32(len(imageContent)) - uint32(rela.Off) + uint32(rela.Addend)
					elfFile.ByteOrder.PutUint32(imageContent[rela.Off:rela.Off+4], targetOffset)

					// TODO: support other length besides uint64 (which is 8 bytes)
					imageContent = append(imageContent, make([]byte, 8)...)

				} else if elfFile.Machine == elf.EM_AARCH64 {
					err := binary.Read(rela_section_reader, elfFile.ByteOrder, &rela)
					if err != nil {
						return nil, errors.Wrapf(err, "read rela section rela64 entry %s", path)
					}

					symNo := rela.Info >> 32
					if symNo == 0 || symNo > uint64(len(syms)) {
						continue
					}
					sym := &syms[symNo-1]
					imageOffset[sym.Name] = len(imageContent)

					targetOffset := uint32(len(imageContent)) - uint32(rela.Off) + uint32(rela.Addend)

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
					instr := elfFile.ByteOrder.Uint32(imageContent[rela.Off : rela.Off+4])
					instr = uint32(int(instr) & ^((1<<19-1)<<5)) | ((targetOffset & (1<<19 - 1)) << 5)
					elfFile.ByteOrder.PutUint32(imageContent[rela.Off:rela.Off+4], instr)

					// TODO: support other length besides uint64 (which is 8 bytes)
					imageContent = append(imageContent, make([]byte, 16)...)
				} else {
					return nil, errors.Errorf("unsupported architecture")
				}
			}

			break
		}
	}
	return NewFakeImage(
		symbolName,
		imageContent,
		imageOffset,
		logger,
	), nil
}
