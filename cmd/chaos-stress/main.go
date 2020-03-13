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
	"math/rand"
	"os"
	"time"

	"github.com/pingcap/chaos-mesh/pkg/chaosstress"
	"github.com/pingcap/chaos-mesh/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	addr         string
	printVersion bool
	log          = ctrl.Log.WithName("chaos-stress")
)

func initFlag() {
	flag.StringVar(&addr, "addr", ":65533", "RPC server address")
	flag.BoolVar(&printVersion, "version", false, "Print version information")
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
}

func main() {
	initFlag()
	version.PrintVersionInfo("chaos-stress")
	if printVersion {
		os.Exit(0)
	}
	if err := chaosstress.StartServer(addr); err != nil {
		log.Error(err, "Server exited")
		os.Exit(1)
	}
}
