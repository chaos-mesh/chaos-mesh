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

package utils

import (
	"bytes"
	"context"

	"github.com/retailnext/iptables_exporter/iptables"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

func GetIptablesContentByNetNS(pid uint32) (tables iptables.Tables, err error) {
	return getIptablesContent(true, pid)
}

func getIptablesContent(enterNS bool, pid uint32) (tables iptables.Tables, err error) {
	builder := bpm.DefaultProcessBuilder("iptables-save", "-c")
	if enterNS {
		builder = builder.SetNS(pid, bpm.NetNS)
	}

	out, err := builder.Build(context.TODO()).CombinedOutput()
	if err != nil {
		return nil, err
	}

	return iptables.ParseIptablesSave(bytes.NewReader(out))
}
