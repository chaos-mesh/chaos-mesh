// Copyright 2022 Chaos Mesh Authors.
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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/helper"
)

var rootCmd = &cobra.Command{
	Use:   "cdh [command]",
	Short: "cdh is a helper to run some logic in another namespaces/cgroups",
	Long: `chaos-daemon sometimes needs to run some logic inside another namespace/cgroup.
We can write these logic inside this helper, and execute them through nsexec.`,
}

func main() {
	rootCmd.AddCommand(helper.NormalizeVolumeNameCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
