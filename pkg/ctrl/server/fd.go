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

package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/model"
)

// GetFdsOfProcess returns fd-target pairs.
// The output looks like:
// ```
// total 0
// lrwx------ 1 docker docker 64 Mar  3 16:11 0 -> /dev/pts/0
// lrwx------ 1 docker docker 64 Mar  3 16:11 1 -> /dev/pts/0
// lrwx------ 1 docker docker 64 Mar  3 16:11 2 -> /dev/pts/0
// lr-x------ 1 docker docker 64 Mar  3 16:11 3 -> /proc/642108/fd
// ```
func (r *Resolver) GetFdsOfProcess(ctx context.Context, process *model.Process) []*model.Fd {
	cmd := fmt.Sprintf("ls -l /proc/%s/fd", process.Pid)
	out, err := r.ExecBypass(ctx, process.Pod, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		// errors often occur on some short-life process, ignored
		r.Log.Error(err, "get fds of process", "pid", process.Pid)
		return nil
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

	return fds
}
