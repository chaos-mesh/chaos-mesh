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

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

$ source <(chaosctl completion bash)

# To load completions for each session, execute once:
Linux:
  $ ./bin/chaosctl completion bash > /etc/bash_completion.d/chaosctl
MacOS:
  $ ./bin/chaosctl completion bash > /usr/local/etc/bash_completion.d/chaosctl

Zsh:

$ compdef _chaosctl ./bin/chaosctl

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ ./bin/chaosctl completion zsh > "${fpath[1]}/_chaosctl"

# You will need to start a new shell for this setup to take effect.

Fish:

$ ./bin/chaosctl completion fish | source

# To load completions for each session, execute once:
$ ./bin/chaosctl completion fish > ~/.config/fish/completions/chaosctl.fish

Powershell:

PS> ./bin/chaosctl completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> ./bin/chaosctl completion powershell > chaosctl.ps1
# and source this file from your powershell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		// TODO: powershell completion is still not fully supported, see https://github.com/spf13/cobra/pull/1208
		// Need to update cobra version when this PR is merged
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
