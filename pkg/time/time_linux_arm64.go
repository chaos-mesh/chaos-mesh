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

package time

import (
	"errors"

<<<<<<< HEAD:pkg/utils/scheduler.go
	"github.com/robfig/cron/v3"

	v1alpha1 "github.com/chaos-mesh/api"
=======
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
>>>>>>> upstream/master:pkg/time/time_linux_arm64.go
)

// ModifyTime modifies time of target process
func ModifyTime(pid int, deltaSec int64, deltaNsec int64, clockIdsMask uint64) error {
	// Mock point to return error in unit test
	if err := mock.On("ModifyTimeError"); err != nil {
		if e, ok := err.(error); ok {
			return e
		}
		if ignore, ok := err.(bool); ok && ignore {
			return nil
		}
	}
	return errors.New("arm64 is not supported")
}
