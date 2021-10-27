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

package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type IptablesTest struct {
	Exec   `exec:"iptables"`
	Port   string   `para:"-p"`
	Ports  []string `para:"--ports"`
	EPort  string   `para:"-ep"`
	EPorts []string `para:"--e_ports"`
	Match  `sub_command:""`
	Match_ `sub_command:""`
}

type Match struct {
	Exec   `exec:"-m"`
	Helper string `para:"--helper"`
}

type Match_ struct {
	Exec   `exec:"-m"`
	Helper string `para:"--helper"`
}

func TestMarshal(t *testing.T) {
	n := IptablesTest{
		NewExec(),
		"20",
		[]string{"2021", "2023"},
		"",
		[]string{"", ""},
		Match{NewExec(), "help"},
		Match_{},
	}
	path, args, err := Marshal(n)
	assert.NoError(t, err, "nil")
	assert.Equal(t, "iptables -p 20 --ports 2021 2023 -m --helper help",
		path+" "+strings.Join(args, " "))
}

type Iptables struct {
	Exec           `exec:"iptables"`
	Tables         string `para:"-t"`
	Command        string `para:""`
	Chain          string `para:""`
	JumpTarget     string `para:"-j"`
	Protocol       string `para:"--protocol"`
	MatchExtension string `para:"-m"`
	SPorts         string `para:"--source-ports"`
	DPorts         string `para:"--destination-ports"`
	SPort          string `para:"--source-port"`
	DPort          string `para:"--destination-port"`
	TcpFlags       string `para:"--tcp-flags"`
}

type TestInvalidParaType struct {
	Exec `exec:"test"`
	P    int `para:"-p"`
}

type TestInvalidParaSliceType struct {
	Exec `exec:"test"`
	P    []int `para:"-p"`
}

func TestMarshalExample(t *testing.T) {
	n := Iptables{
		Exec:           NewExec(),
		Command:        "-A",
		Chain:          "Chaos_Chain",
		JumpTarget:     "Chaos_Target",
		Protocol:       "tcp",
		MatchExtension: "multiport",
		SPorts:         "2021,2022",
		TcpFlags:       "SYN",
	}
	path, args, err := Marshal(n)
	assert.NoError(t, err, "nil")
	assert.Equal(t, "iptables -A Chaos_Chain -j Chaos_Target --protocol tcp -m multiport --source-ports 2021,2022 --tcp-flags SYN",
		path+" "+strings.Join(args, " "))

	p := TestInvalidParaType{
		Exec: NewExec(),
		P:    2,
	}
	_, _, err = Marshal(p)
	assert.EqualError(t, err, "invalid parameter type int : parameter must be string or string slice")
	ps := TestInvalidParaSliceType{
		Exec: NewExec(),
		P:    nil,
	}
	_, _, err = Marshal(ps)
	assert.EqualError(t, err, "invalid parameter slice type <[]int Value> :parameter slice must be string slice")
}
