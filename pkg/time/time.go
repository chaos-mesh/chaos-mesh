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

package time

import (
	"fmt"

	"github.com/pingcap/chaos-mesh/pkg/mapreader"
	"github.com/pingcap/chaos-mesh/pkg/ptrace"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("time")

func ModifyTime(pid int, delta_sec int64, delta_nsec int64) error {
	program, err := ptrace.Trace(pid)
	if err != nil {
		return err
	}

	var entry *mapreader.Entry
	for _, e := range *program.Entries {
		e := e
		if e.Path == "[vdso]" {
			entry = &e
		}
	}

	clockGettimeAddr, err := program.FindSymbolInEntry("clock_gettime", entry)
	if err != nil {
		return err
	}
	log.Info("get clock_gettime address", "addr", clockGettimeAddr)
	fmt.Printf("get clock_gettime addr: %x", clockGettimeAddr)

	return nil
}
