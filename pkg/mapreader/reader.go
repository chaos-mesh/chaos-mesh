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

// Package mapreader parses Linux process memory mappings from /proc into address ranges and file paths.
package mapreader

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Entry is one line in /proc/pid/maps
type Entry struct {
	StartAddress uint64
	EndAddress   uint64
	Privilege    string
	PaddingSize  uint64
	Path         string
}

// Read parse /proc/[pid]/maps and return a list of entry
// The format of /proc/[pid]/maps can be found in `man proc`.
func Read(pid int) ([]Entry, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return parseMaps(string(data))
}

// The pathname extends to the end of the line and may contain spaces.
var mapLine = regexp.MustCompile(`^([^ 	]+)[ 	]+([^ 	]+)[ 	]+([^ 	]+)[ 	]+[^ 	]+[ 	]+[^ 	]+(?:[ 	]+(.*))?$`)

func parseMaps(data string) ([]Entry, error) {
	var entries []Entry
	for index, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := mapLine.FindStringSubmatch(line)
		if fields == nil {
			return nil, fmt.Errorf("malformed mapping on line %d", index+1)
		}
		start, end, ok := strings.Cut(fields[1], "-")
		if !ok {
			return nil, fmt.Errorf("missing address range on line %d", index+1)
		}
		startAddress, err := strconv.ParseUint(start, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid start address on line %d: %w", index+1, err)
		}
		endAddress, err := strconv.ParseUint(end, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end address on line %d: %w", index+1, err)
		}
		if endAddress <= startAddress {
			return nil, fmt.Errorf("invalid address range on line %d", index+1)
		}
		paddingSize, err := strconv.ParseUint(fields[3], 16, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid offset on line %d: %w", index+1, err)
		}
		entries = append(entries, Entry{
			StartAddress: startAddress,
			EndAddress:   endAddress,
			Privilege:    fields[2],
			PaddingSize:  paddingSize,
			Path:         fields[4],
		})
	}
	return entries, nil
}
