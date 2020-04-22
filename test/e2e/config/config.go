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

package config

import (
	"flag"

	"github.com/pingcap/chaos-mesh/test"
)

// TestConfig for the test config
var TestConfig *test.Config = test.NewDefaultConfig()

// RegisterOperatorFlags registers flags for chaos-mesh.
func RegisterOperatorFlags(flags *flag.FlagSet) {
	flags.StringVar(&TestConfig.ManagerImage, "manager-image", "pingcap/chaos-mesh", "chaos-mesh image")
	flags.StringVar(&TestConfig.ManagerTag, "manager-image-tag", "latest", "chaos-mesh image tag")
	flags.StringVar(&TestConfig.DaemonImage, "daemon-image", "pingcap/chaos-daemon", "chaos-daemon image")
	flags.StringVar(&TestConfig.DaemonTag, "daemon-image-tag", "latest", "chaos-daemon image tag")
	flags.StringVar(&TestConfig.E2EImage, "e2e-image", "pingcap/e2e-helper:latest", "e2e helper image")
}
