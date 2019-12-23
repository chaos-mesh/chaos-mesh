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

package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/pingcap/chaos-operator/pkg/chaosdaemon"
	"github.com/pingcap/chaos-operator/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var log = ctrl.Log.WithName("chaos-daemon")

var (
	printVersion bool
	rawPort      string
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.StringVar(&rawPort, "port", "", "the port which server listens on")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo("Chaos-daemon")

	if printVersion {
		os.Exit(0)
	}

	ctrl.SetLogger(zap.Logger(true))

	if rawPort == "" {
		rawPort = os.Getenv("PORT")
	}

	if rawPort == "" {
		rawPort = "8080"
	}

	port, err := strconv.Atoi(rawPort)
	if err != nil {
		log.Error(err, "Error while parsing PORT environment variable", "port", rawPort)
		os.Exit(1)
	}
	log.Info("starting server")
	chaosdaemon.StartServer("0.0.0.0", port)
}
