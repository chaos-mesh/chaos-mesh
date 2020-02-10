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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pingcap/chaos-mesh/pkg/version"

	"github.com/pingcap/chaos-mesh/pkg/time"
)

var (
	pid             int
	sec_delta       int64
	nsec_delta      int64
	fake_image_path string
	printVersion    bool
)

func initFlag() {
	flag.IntVar(&pid, "pid", 0, "pid of target program")
	flag.Int64Var(&sec_delta, "sec_delta", 0, "delta time of sec field")
	flag.Int64Var(&nsec_delta, "nsec_delta", 0, "delta time of nsec field")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")

	flag.Parse()
}

func main() {
	initFlag()

	version.PrintVersionInfo("watchmaker")

	if printVersion {
		os.Exit(0)
	}

	err := time.ModifyTime(pid, sec_delta, nsec_delta, fake_image_path)

	if err != nil {
		fmt.Printf("error while modifying time, pid: %d, sec_delta: %d, nsec_delta: %d\n Error: %s", pid, sec_delta, nsec_delta, err.Error())
	}
}
