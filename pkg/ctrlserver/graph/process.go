// Copyright 2021 Chaos Mesh Authors.
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

package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// GetPidFromPS returns pid-command pairs
func (r *Resolver) GetPidFromPS(ctx context.Context, pod *v1.Pod) ([]string, []string, error) {
	cmd := "ps"
	out, err := r.ExecBypass(ctx, pod, cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "run command %s failed", cmd)
	}
	outLines := strings.Split(string(out), "\n")
	if len(outLines) < 2 {
		return nil, nil, fmt.Errorf("ps returns empty")
	}
	titles := strings.Fields(outLines[0])
	var pidColumn, cmdColumn int
	for i, t := range titles {
		if t == "PID" {
			pidColumn = i
		}
		if t == "COMMAND" || t == "CMD" {
			cmdColumn = i
		}
	}
	if pidColumn == 0 && cmdColumn == 0 {
		return nil, nil, fmt.Errorf("parsing ps error: could not get PID and COMMAND column")
	}
	var pids, commands []string
	for _, line := range outLines[1:] {
		item := strings.Fields(line)
		// break when got empty line
		if len(item) == 0 {
			break
		}
		pids = append(pids, item[pidColumn])
		commands = append(commands, item[cmdColumn])
	}
	return pids, commands, nil
}
