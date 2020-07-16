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

package netem

import (
	"net"

	"github.com/pingcap/chaos-mesh/controllers/networkchaos/netutils"
)

func resolveCidrs(name string) (string, error) {
	_, ipnet, err := net.ParseCIDR(name)
	if err == nil {
		return ipnet.String(), nil
	}

	if net.ParseIP(name) != nil {
		return netutils.IPToCidr(name), nil
	}

	addrs, err := net.LookupIP(name)
	if err != nil {
		return "", err
	}

	// TODO: support IPv6
	return addrs[0].String(), nil
}
