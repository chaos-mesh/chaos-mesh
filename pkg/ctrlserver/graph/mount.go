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
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

// GetMounts returns mounts info
func (r *Resolver) GetMounts(ctx context.Context, pod *v1.Pod) ([]string, error) {
	cmd := "cat /proc/mounts"
	out, err := r.ExecBypass(ctx, pod, cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "run command %s failed", cmd)
	}
	return strings.Split(string(out), "\n"), nil
}
