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

package netutils

import (
	"net"
	"strings"
)

// IPToCidr converts from an ip to a full mask cidr
func IPToCidr(ip string) string {
	// TODO: support IPv6
	return ip + "/32"
}

// ResolveCidrs converts multiple cidrs/ips/domains into cidr
func ResolveCidrs(names []string) ([]string, error) {
	cidrs := []string{}
	for _, target := range names {
		// TODO: resolve ip on every pods but not in controller, in case the dns server of these pods differ
		cidr, err := ResolveCidr(target)
		if err != nil {
			return nil, err
		}

		cidrs = append(cidrs, cidr...)
	}

	return cidrs, nil
}

// ResolveCidr converts cidr/ip/domain into cidr
func ResolveCidr(name string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(name)
	if err == nil {
		return []string{ipnet.String()}, nil
	}

	if net.ParseIP(name) != nil {
		return []string{IPToCidr(name)}, nil
	}

	addrs, err := net.LookupIP(name)
	if err != nil {
		return nil, err
	}

	cidrs := []string{}
	for _, addr := range addrs {
		addr := addr.String()

		// TODO: support IPv6
		if strings.Contains(addr, ".") {
			cidrs = append(cidrs, IPToCidr(addr))
		}
	}
	return cidrs, nil
}
