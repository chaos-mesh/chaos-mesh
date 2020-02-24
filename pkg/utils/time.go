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

package utils

import "fmt"

// EncodeClkIds will convert array of clk ids into a mask
func EncodeClkIds(clkIds []string) (uint64, error) {
	mask := uint64(0)

	for _, id := range clkIds {
		// refer to `uapi/linux/time.h`
		if id == "CLOCK_REALTIME" {
			mask |= 1 << 0
		} else if id == "CLOCK_MONOTONIC" {
			mask |= 1 << 1
		} else if id == "CLOCK_PROCESS_CPUTIME_ID" {
			mask |= 1 << 2
		} else if id == "CLOCK_THREAD_CPUTIME_ID" {
			mask |= 1 << 3
		} else if id == "CLOCK_MONOTONIC_RAW" {
			mask |= 1 << 4
		} else if id == "CLOCK_REALTIME_COARSE" {
			mask |= 1 << 5
		} else if id == "CLOCK_MONOTONIC_COARSE" {
			mask |= 1 << 6
		} else if id == "CLOCK_BOOTTIME" {
			mask |= 1 << 7
		} else if id == "CLOCK_REALTIME_ALARM" {
			mask |= 1 << 8
		} else if id == "CLOCK_BOOTTIME_ALARM" {
			mask |= 1 << 9
		} else {
			return 0, fmt.Errorf("unknown clock id %s", id)
		}
	}

	return mask, nil
}
