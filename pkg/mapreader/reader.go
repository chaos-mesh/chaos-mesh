// Copyright 2020 PingCAP, Inc.
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

package mapreader

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type Entry struct {
	StatAddress uint64
	EndAddress  uint64
	Privilege   string
	PaddingSize uint64
	Path        string
}

func Read(pid int) (error, *[]Entry) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/maps", pid))
	if err != nil {
		return err, nil
	}

	lines := strings.Split(string(data), "\n")

	var entries []Entry
	for _, line := range lines {
		sections := strings.Split(line, " ")
		if len(sections) < 3 {
			continue
		}

		var path string

		if len(sections) > 5 {
			path = sections[len(sections)-1]
		} else {
			path = ""
		}

		addresses := strings.Split(sections[0], "-")
		startAddress, err := strconv.ParseUint(addresses[0], 16, 64)
		if err != nil {
			return err, nil
		}
		endAddresses, err := strconv.ParseUint(addresses[1], 16, 64)
		if err != nil {
			return err, nil
		}

		privilege := sections[1]

		paddingSize, err := strconv.ParseUint(sections[2], 16, 64)
		if err != nil {
			return err, nil
		}

		entries = append(entries, Entry{
			startAddress,
			endAddresses,
			privilege,
			paddingSize,
			path,
		})
	}

	return nil, &entries
}
