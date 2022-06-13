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
				err := binary.Read(rela_section_reader, elfFile.ByteOrder, &rela)
				if err != nil {
					return nil, errors.Wrapf(err, "read rela section rela64 entry %s", path)
				}

				symNo := rela.Info >> 32
				if symNo == 0 || symNo > uint64(len(syms)) {
					continue
				}

				sym := syms[symNo-1]
				byteorder := elfFile.ByteOrder
				if elfFile.Machine == elf.EM_X86_64 || elfFile.Machine == elf.EM_AARCH64 {
					AssetLD(rela, imageOffset, &imageContent, sym, byteorder)
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
