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

package netutils

import (
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// IPToCidr converts from an ip to a full mask cidr
func IPToCidr(ip string) string {
	// TODO: support IPv6
	return ip + "/32"
}

// ResolveCidrs converts multiple cidrs/ips/domains into cidr
func ResolveCidrs(names []string) ([]v1alpha1.CidrAndPort, error) {
	cidrs := []v1alpha1.CidrAndPort{}
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
func ResolveCidr(name string) ([]v1alpha1.CidrAndPort, error) {
	var toResolve string
	var port uint16

	if host, portStr, err := net.SplitHostPort(name); err == nil {
		toResolve = host
		port64, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			return nil, errors.Errorf("parse port %s", err)
		}
		port = uint16(port64)
	} else {
		toResolve = name
	}

	_, ipnet, err := net.ParseCIDR(toResolve)
	if err == nil {
		return []v1alpha1.CidrAndPort{{Cidr: ipnet.String(), Port: port}}, nil
	}

	if net.ParseIP(toResolve) != nil {
		return []v1alpha1.CidrAndPort{{Cidr: IPToCidr(toResolve), Port: port}}, nil
	}

	addrs, err := LookupIP(toResolve)
	if err != nil {
		return nil, err
	}

	cidrs := []v1alpha1.CidrAndPort{}
	for _, addr := range addrs {
		addr := addr.String()

		// TODO: support IPv6
		if strings.Contains(addr, ".") {
			cidrs = append(cidrs, v1alpha1.CidrAndPort{Cidr: IPToCidr(addr), Port: port})
		}
	}
	return cidrs, nil
}

func LookupIP(host string) ([]net.IP, error) {
	if result := mock.On("LookupIP"); result != nil {
		return result.([]net.IP), nil
	}
	return net.LookupIP(host)
}
