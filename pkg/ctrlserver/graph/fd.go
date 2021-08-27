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

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

// GetFdsOfProcess returns fd-target pairs
func (r *Resolver) GetFdsOfProcess(ctx context.Context, process *model.Process) ([]*model.Fd, error) {
	cmd := fmt.Sprintf("ls -l /proc/%s/fd", process.Pid)
	out, err := r.ExecBypass(ctx, process.Pod, cmd)
	if err != nil {
		return nil, err
	}
	var fds []*model.Fd
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		length := len(fields)
		if length < 3 {
			// skip
			continue
		}
		fd := &model.Fd{
			Fd:     fields[length-3],
			Target: fields[length-1],
		}
		fds = append(fds, fd)
	}

	return fds, nil
}
