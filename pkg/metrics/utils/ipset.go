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
	"context"
	"encoding/xml"

	"github.com/romana/ipset"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

func GetIPSetRulesNumberByNetNS(pid uint32) (int, error) {
	return getIPSetRulesNumber(true, pid)
}

func getIPSetRulesNumber(enterNS bool, pid uint32) (int, error) {
	builder := bpm.DefaultProcessBuilder("ipset", "save", "-o", "xml")
	if enterNS {
		builder = builder.SetNS(pid, bpm.NetNS)
	}

	out, err := builder.Build(context.TODO()).CombinedOutput()
	if err != nil {
		return 0, err
	}

	var sets ipset.Ipset
	if err = xml.Unmarshal(out, &sets); err != nil {
		return 0, err
	}

	var members int
	for _, set := range sets.Sets {
		members += len(set.Members)
	}
	return members, nil
}
