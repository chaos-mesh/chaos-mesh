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
	"os"
)

//go:embed fakeclock/*.o
var fakeclock embed.FS

// FakeImage introduce the replacement of VDSO ELF entry and customizable variables.
// FakeImage could be constructed by LoadFakeImageFromEmbedFs(), and then used by FakeClockInjector.
type FakeImage struct {
	// content presents .text section which has been "manually relocation", the address of extern variables have been calculated manually
	content []byte
	// offset stores the table with variable name, and it's address in content.
	// the key presents extern variable name, ths value is the address/offset within the content.
	offset map[string]int
}

// LoadFakeImageFromEmbedFs builds FakeImage from the embed filesystem.
// It parses the ELF file and extract the variables from the relocation section,
// reserves the space for them at the end of content,
// then calculates and saves offsets as "manually relocation"
func LoadFakeImageFromEmbedFs(filename string) (*FakeImage, error) {
	path := "fakeclock/" + filename
	object, err := fakeclock.ReadFile(path)
	if err != nil {
		log.Error(err, "read file from embedded fs", "path", path)
		os.Exit(1)
	}

	elfFile, err := elf.NewFile(bytes.NewReader(object))
	if err != nil {
		log.Error(err, "parse elf", "path", path)
		os.Exit(1)
	}

	syms, err := elfFile.Symbols()
	if err != nil {
		log.Error(err, "get symbols")
		os.Exit(1)
	}

	fakeImage := FakeImage{
		offset: make(map[string]int),
	}
	for _, r := range elfFile.Sections {
		if r.Type == elf.SHT_PROGBITS && r.Name == ".text" {
			fakeImage.content, err = r.Data()
			if err != nil {
				log.Error(err, "read text section")
				os.Exit(1)
			}

			break
		}
	}

	for _, r := range elfFile.Sections {
		if r.Type == elf.SHT_RELA && r.Name == ".rela.text" {
			rela_section, err := r.Data()
			if err != nil {
				log.Error(err, "read rela section")
				os.Exit(1)
			}
			rela_section_reader := bytes.NewReader(rela_section)

			var rela elf.Rela64
			for rela_section_reader.Len() > 0 {
				binary.Read(rela_section_reader, elfFile.ByteOrder, &rela)

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
				// len(fakeImage.content) - 4 - 0x16

				sym := &syms[symNo-1]
				fakeImage.offset[sym.Name] = len(fakeImage.content)
				targetOffset := uint32(len(fakeImage.content)) - uint32(rela.Off) + uint32(rela.Addend)
				elfFile.ByteOrder.PutUint32(fakeImage.content[rela.Off:rela.Off+4], targetOffset)

				// TODO: support other length besides uint64 (which is 8 bytes)
				fakeImage.content = append(fakeImage.content, make([]byte, 8)...)
			}

			break
		}
	}
	return &fakeImage, nil
}
