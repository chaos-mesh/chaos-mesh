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

package utils

import (
	"bufio"
	"bytes"
	"context"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

func GetTcRulesNumberByNetNS(pid uint32) (int, error) {
	return getTcRulesNumber(true, pid)
}

func getTcRulesNumber(enterNS bool, pid uint32) (int, error) {
	builder := bpm.DefaultProcessBuilder("tc", "qdisc")
	if enterNS {
		builder = builder.SetNS(pid, bpm.NetNS)
	}

	out, err := builder.Build(context.TODO()).CombinedOutput()
	if err != nil {
		return 0, err
	}

	var lines int
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines++
	}

	return lines, nil
}
