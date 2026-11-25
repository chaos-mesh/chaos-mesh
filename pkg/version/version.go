// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package version reports Chaos Mesh build, source revision, and Go runtime version information.
package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

var (
	gitVersion string
	buildDate  = "unknown"
)

// PrintVersionInfo show version info to Stdout
func PrintVersionInfo(name string) {
	fmt.Printf("%s Version: %#v\n", name, Get())
}

// Get reports the running executable's version and native build metadata.
func Get() Info {
	info := Info{
		GitVersion: "(devel)",
		GitCommit:  "unknown",
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		if buildInfo.Main.Version != "" {
			info.GitVersion = buildInfo.Main.Version
		}
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				info.GitCommit = setting.Value
			}
		}
	}
	if gitVersion != "" {
		info.GitVersion = gitVersion
	}
	return info
}
