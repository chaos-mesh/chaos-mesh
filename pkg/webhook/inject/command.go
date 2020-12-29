// Copyright 2020 Chaos Mesh Authors.
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

package inject

import "strings"

// MergeCommands merges injected commands and original commands for injecting commands to containers,
// eg: inject: []string{"bash", "-c", "/check.sh"}, origin: []string{"bash", "-c", "/run.sh"}
// merged commands: []string{"/bin/sh", "-ec", "/check.sh\n/run.sh"}
func MergeCommands(inject []string, origin []string, args []string) []string {
	// merge injected commands
	scripts := mergeCommandsAction(inject)

	// merge original commands
	scripts += mergeOriginCommandsAndArgs(origin, args)

	return []string{"/bin/sh", "-ec", scripts}
}

func mergeCommandsAction(commands []string) string {
	scripts := ""

	for i := 0; i < len(commands); i++ {
		cmd := commands[i]
		if isCommonScripts(cmd) {
			if len(commands) <= i+1 {
				scripts += cmd
				scripts += "\n"
				break
			}

			if strings.HasPrefix(commands[i+1], "-") {
				i++
				continue
			}

			tempScripts := cmd + " "
			for j := i + 1; j < len(commands); j++ {
				c := commands[j]
				if j < len(commands)-1 {
					c += " "
				}

				tempScripts += c
			}

			scripts += tempScripts
			scripts += "\n"
			break
		}

		if len(commands) <= i+1 {
			scripts += cmd
			scripts += "\n"
			break
		}

		if strings.HasPrefix(commands[i+1], "-") {
			tempScripts := cmd + " "
			for j := i + 1; j < len(commands); j++ {
				if strings.HasPrefix(commands[j], "-") {
					c := commands[j]
					if j < len(commands)-1 {
						c += " "
					}
					tempScripts += c
					continue
				}
			}
			scripts += tempScripts
			scripts += "\n"
			break
		}

		scripts += cmd
		scripts += "\n"
	}

	return scripts
}

func mergeOriginCommandsAndArgs(origin []string, args []string) string {
	commands := append(origin, args...)

	return mergeCommandsAction(commands)
}

func isCommonScripts(cmd string) bool {
	if isShellScripts(cmd) || isPythonScripts(cmd) {
		return true
	}

	return false
}

func isShellScripts(cmd string) bool {
	if cmd == "bash" || cmd == "sh" ||
		strings.HasPrefix(cmd, "bash ") || strings.HasPrefix(cmd, "sh ") ||
		strings.HasPrefix(cmd, "/bin/sh") || strings.HasPrefix(cmd, "/bin/bash") ||
		strings.HasPrefix(cmd, "/usr/bin/sh") || strings.HasPrefix(cmd, "/usr/bin/bash") ||
		strings.HasPrefix(cmd, "/usr/share/bin/sh") || strings.HasPrefix(cmd, "/usr/share/bin/bash") {
		return true
	}

	return false
}

func isPythonScripts(cmd string) bool {
	if cmd == "python" || cmd == "python3" ||
		strings.HasPrefix(cmd, "python ") || strings.HasPrefix(cmd, "python3 ") ||
		strings.HasPrefix(cmd, "/bin/python") || strings.HasPrefix(cmd, "/bin/python3") ||
		strings.HasPrefix(cmd, "/usr/bin/python") || strings.HasPrefix(cmd, "/usr/bin/python3") ||
		strings.HasPrefix(cmd, "/usr/share/bin/python") || strings.HasPrefix(cmd, "/usr/share/bin/python3") {
		return true
	}

	return false
}
