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

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestIsCommonScripts(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		cmd           string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "command scripts: bash",
			cmd:           "bash",
			expectedValue: true,
		},
		{
			name:          "command scripts: bash -c echo 1",
			cmd:           "bash -c echo 1",
			expectedValue: true,
		},
		{
			name:          "command scripts: sh",
			cmd:           "sh",
			expectedValue: true,
		},
		{
			name:          "command scripts: /bin/sh",
			cmd:           "/bin/sh",
			expectedValue: true,
		},
		{
			name:          "command scripts: /bin/bash",
			cmd:           "/bin/bash",
			expectedValue: true,
		},
		{
			name:          "command scripts: /usr/bin/bash",
			cmd:           "/usr/bin/bash",
			expectedValue: true,
		},
		{
			name:          "not command scripts: /usr/bin/echo",
			cmd:           "/usr/bin/echo",
			expectedValue: false,
		},
		{
			name:          "not command scripts: /chaos-mesh",
			cmd:           "/chaos-mesh",
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		g.Expect(isCommonScripts(tc.cmd)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestIsShellScripts(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		cmd           string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "bash",
			cmd:           "bash",
			expectedValue: true,
		},
		{
			name:          "bash -c echo 1",
			cmd:           "bash -c echo 1",
			expectedValue: true,
		},
		{
			name:          "/usr/bin/bash",
			cmd:           "/usr/bin/bash",
			expectedValue: true,
		},
		{
			name:          "/usr/bin/echo",
			cmd:           "/usr/bin/echo",
			expectedValue: false,
		},
		{
			name:          "/chaos-mesh",
			cmd:           "/chaos-mesh",
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		g.Expect(isShellScripts(tc.cmd)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestIsPythonScripts(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		cmd           string
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name:          "python",
			cmd:           "python",
			expectedValue: true,
		},
		{
			name:          "bash -c echo 1",
			cmd:           "bash -c echo 1",
			expectedValue: false,
		},
		{
			name:          "/bin/python",
			cmd:           "/bin/python",
			expectedValue: true,
		},
		{
			name:          "/usr/bin/bash",
			cmd:           "/usr/bin/bash",
			expectedValue: false,
		},
		{
			name:          "/usr/bin/echo",
			cmd:           "/usr/bin/echo",
			expectedValue: false,
		},
		{
			name:          "/chaos-mesh",
			cmd:           "/chaos-mesh",
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		g.Expect(isPythonScripts(tc.cmd)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestMergeCommandsAction(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		commands      []string
		expectedValue string
	}

	tcs := []TestCase{
		{
			name: "scripts files",
			commands: []string{
				"/bin/sh",
				"start_chaos_mesh.sh",
			},
			expectedValue: "/bin/sh start_chaos_mesh.sh\n",
		},
		{
			name: "common scripts",
			commands: []string{
				"bash",
				"-ec",
				"echo $HOSENAME\n /start_chaos_mesh.sh -s t1 \n -v t2",
			},
			expectedValue: "echo $HOSENAME\n /start_chaos_mesh.sh -s t1 \n -v t2\n",
		},
		{
			name: "common start scripts",
			commands: []string{
				"/chaos-mesh",
				"--c t",
				"--v 2",
				"--log stdout",
			},
			expectedValue: "/chaos-mesh --c t --v 2 --log stdout\n",
		},
		{
			name: "one line",
			commands: []string{
				"/chaos-mesh --c t --v 2 --log stdout",
			},
			expectedValue: "/chaos-mesh --c t --v 2 --log stdout\n",
		},
	}

	for _, tc := range tcs {
		g.Expect(mergeCommandsAction(tc.commands)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestMergeOriginCommandsAndArgs(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		commands      []string
		args          []string
		expectedValue string
	}

	tcs := []TestCase{
		{
			name: "only commands",
			commands: []string{
				"bash",
				"-ec",
				"echo $HOSENAME\n /start_chaos_mesh.sh -s t1 \n -v t2",
			},
			expectedValue: "echo $HOSENAME\n /start_chaos_mesh.sh -s t1 \n -v t2\n",
		},
		{
			name: "only args",
			args: []string{
				"bash",
				"-ec",
				"echo $HOSENAME\n /start_chaos_mesh.sh -s t1 \n -v t2",
			},
			expectedValue: "echo $HOSENAME\n /start_chaos_mesh.sh -s t1 \n -v t2\n",
		},
		{
			name: "commands and args",
			commands: []string{
				"/chaos-mesh",
			},
			args: []string{
				"--c t",
				"--v 2",
				"--log stdout",
			},
			expectedValue: "/chaos-mesh --c t --v 2 --log stdout\n",
		},
	}

	for _, tc := range tcs {
		g.Expect(mergeOriginCommandsAndArgs(tc.commands, tc.args)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestMergeCommands(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name          string
		inject        []string
		origin        []string
		args          []string
		expectedValue []string
	}

	tcs := []TestCase{
		{
			name: "scripts file",
			inject: []string{
				"/check.sh -v 1",
			},
			origin: []string{
				"/bin/sh",
				"/start.sh",
			},
			expectedValue: []string{
				"/bin/sh",
				"-ec",
				"/check.sh -v 1\n/bin/sh /start.sh\n",
			},
		},
		{
			name: "common scripts",
			inject: []string{
				"/check.sh -v 1",
			},
			origin: []string{
				"bash",
				"-c",
				"set -ex\n[[ `hostname` =~ -([0-9]+)$ ]] || exit 1\n/tiflash server --config-file /data/config.toml",
			},
			expectedValue: []string{
				"/bin/sh",
				"-ec",
				"/check.sh -v 1\nset -ex\n[[ `hostname` =~ -([0-9]+)$ ]] || exit 1\n/tiflash server --config-file /data/config.toml\n",
			},
		},
	}

	for _, tc := range tcs {
		g.Expect(MergeCommands(tc.inject, tc.origin, tc.args)).To(Equal(tc.expectedValue), tc.name)
	}
}
