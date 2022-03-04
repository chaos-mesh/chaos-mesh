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
	"github.com/spf13/cobra"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/create"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-multiversion-helper/migrate"
)

var rootCmd = &cobra.Command{
	Use:   "chaos-multiversion-helper",
	Short: "chaos-multiversion-helper is a command to help developer to create a new version of chaos api",
}

func main() {
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(migrate.MigrateCmd)

	rootCmd.Execute()
}
