// Copyright 2020 Chaos Mesh Authors.
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

package chaosdaemon

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func applyTc(ctx context.Context, pid uint32, withoutNS bool, args ...string) error {
	// Mock point to return error in unit test
	if err := mock.On("TcApplyError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}

	processBuilder := bpm.DefaultProcessBuilder("tc", args...).SetContext(ctx)
	if !withoutNS {
		processBuilder = processBuilder.SetNS(pid, bpm.NetNS)
	}
	cmd := processBuilder.Build()
	log.Info("tc command", "command", cmd.String(), "args", args)

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Error(err, "tc command error", "command", cmd.String(), "output", string(out))
		return err
	}

	return nil
}
