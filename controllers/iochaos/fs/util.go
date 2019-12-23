// Copyright 2019 PingCAP, Inc.
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

package fs

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	fspb "github.com/pingcap/chaos-mesh/pkg/chaosfs/pb"
)

const (
	ioChaosDelayActionMsg = "inject file system io delay for %s"
	ioChaosErrnoActionMsg = "inject file system errno delay for %s"
	ioChaosMixedChaosMsg  = "inject file system mixed chaos for %s"
)

func genMessage(iochaos *v1alpha1.IoChaos) string {
	switch iochaos.Spec.Action {
	case v1alpha1.IODelayAction:
		return fmt.Sprintf(ioChaosDelayActionMsg, iochaos.Spec.Duration)
	case v1alpha1.IOErrnoAction:
		return fmt.Sprintf(ioChaosErrnoActionMsg, iochaos.Spec.Duration)
	case v1alpha1.IOMixedAction:
		return fmt.Sprintf(ioChaosMixedChaosMsg, iochaos.Spec.Duration)
	default:
		return ""
	}
}

func genChaosfsRequest(iochaos *v1alpha1.IoChaos) (*fspb.Request, error) {
	req := &fspb.Request{
		Pct:  100,
		Path: iochaos.Spec.Path,
	}

	if iochaos.Spec.Percent != "" {
		percent, err := strconv.Atoi(iochaos.Spec.Percent)
		if err != nil {
			return nil, err
		}

		if percent <= 0 || percent > 100 {
			return nil, fmt.Errorf("iochaos percentage value of %d is invalid, Must be (0,100]", percent)
		}
		req.Pct = uint32(percent)
	}

	if len(iochaos.Spec.Methods) > 0 {
		req.Methods = iochaos.Spec.Methods
	}

	if iochaos.Spec.Action == v1alpha1.IODelayAction || iochaos.Spec.Action == v1alpha1.IOMixedAction {
		delay, err := time.ParseDuration(iochaos.Spec.Delay)
		if err != nil {
			return nil, err
		}
		req.Delay = uint32(delay.Nanoseconds() / 1000)
	}

	if iochaos.Spec.Action == v1alpha1.IOErrnoAction || iochaos.Spec.Action == v1alpha1.IOMixedAction {
		req.Random = true
		if iochaos.Spec.Errno != "" {
			errno, err := strconv.Atoi(iochaos.Spec.Errno)
			if err != nil {
				return nil, err
			}
			req.Random = false
			req.Errno = uint32(errno)
		}
	}

	return req, nil
}
