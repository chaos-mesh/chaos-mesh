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

package manager

import (
	"fmt"

	"github.com/robfig/cron/v3"
)

type Job interface {
	Run()
}

// Runner is the base unit for performing chaos action.
type Runner struct {
	Name    string
	Rule    string
	EntryID int
	Job     Job
}

func (r *Runner) Validate() error {
	if len(r.Name) == 0 {
		return fmt.Errorf("runner name is empty")
	}

	if len(r.Rule) == 0 {
		return fmt.Errorf("runner rule is empty")
	}

	if _, err := cron.ParseStandard(r.Rule); err != nil {
		return fmt.Errorf("fail to parse runner rule %s, %v", r.Rule, err)
	}

	if r.Job == nil {
		return fmt.Errorf("runner job is empty")
	}

	return nil
}
