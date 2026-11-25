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

package mapreader

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestReadProcessMaps(t *testing.T) {
	g := NewWithT(t)
	entries, err := Read(os.Getpid())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(entries).NotTo(BeEmpty())
	g.Expect(Read(-1)).Error().To(MatchError(os.ErrNotExist))
}

func TestParseMappings(t *testing.T) {
	g := NewWithT(t)
	input := "10-20 r--p 0008 00:00 0\t /tmp/path with spaces (deleted)\n" +
		"20-30\trw-p\t0000 00:00 0  \n\n" +
		"30-40 r-xp 0000 00:00 0   [vdso]\n"
	g.Expect(parseMaps(input)).To(Equal([]Entry{
		{StartAddress: 0x10, EndAddress: 0x20, Privilege: "r--p", PaddingSize: 0x8, Path: "/tmp/path with spaces (deleted)"},
		{StartAddress: 0x20, EndAddress: 0x30, Privilege: "rw-p"},
		{StartAddress: 0x30, EndAddress: 0x40, Privilege: "r-xp", Path: "[vdso]"},
	}))
}

func TestParseMalformedMappings(t *testing.T) {
	for _, test := range []struct {
		line, message string
	}{
		{"10-20 r--p 0000", "malformed mapping"},
		{"10 r--p 0000 00:00 0", "missing address range"},
		{"xx-20 r--p 0000 00:00 0", "invalid start address"},
		{"10-xx r--p 0000 00:00 0", "invalid end address"},
		{"20-10 r--p 0000 00:00 0", "invalid address range"},
		{"10-20 r--p xxxx 00:00 0", "invalid offset"},
	} {
		t.Run(test.message, func(t *testing.T) {
			g := NewWithT(t)
			entries, err := parseMaps("01-02 r--p 0000 00:00 0\n" + test.line)
			g.Expect(entries).To(BeNil())
			g.Expect(err).To(MatchError(ContainSubstring(test.message + " on line 2")))
		})
	}
}
