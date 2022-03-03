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
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

// GetMounts returns mounts info
// The output looks like:
// ```
// proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
// sys /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
// dev /dev devtmpfs rw,nosuid,relatime,size=16283300k,nr_inodes=4070825,mode=755,inode64 0 0
// run /run tmpfs rw,nosuid,nodev,relatime,mode=755,inode64 0 0
// tmpfs /dev/shm tmpfs rw,nosuid,nodev,inode64 0 0
// cgroup2 /sys/fs/cgroup cgroup2 rw,nosuid,nodev,noexec,relatime,nsdelegate,memory_recursiveprot 0 0
// tmpfs /run/user/1000 tmpfs rw,nosuid,nodev,relatime,size=3258252k,nr_inodes=814563,mode=700,uid=1000,gid=1000,inode64 0 0
// ```
func (r *Resolver) GetMounts(ctx context.Context, pod *v1.Pod) ([]string, error) {
	cmd := "cat /proc/mounts"
	out, err := r.ExecBypass(ctx, pod, cmd, bpm.PidNS, bpm.MountNS)
	if err != nil {
		return nil, errors.Wrapf(err, "run command %s failed", cmd)
	}
	return strings.Split(string(out), "\n"), nil
}
