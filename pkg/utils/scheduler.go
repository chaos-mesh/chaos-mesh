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

package utils

import (
	"fmt"
	"time"

	cron "github.com/robfig/cron/v3"

	"github.com/pingcap/chaos-operator/api/v1alpha1"
)

func NextTime(spec v1alpha1.SchedulerSpec, now time.Time) (*time.Time, error) {
	scheduler, err := cron.ParseStandard(spec.Cron)
	if err != nil {
		return nil, fmt.Errorf("fail to parse runner rule %s, %v", spec.Cron, err)
	}

	next := scheduler.Next(now)
	return &next, nil
}
