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

	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

// Target is a resolved network and an optional port.
type Target struct {
	Network *net.IPNet
	Port    uint16
	FromDNS bool
}

// IPToCidr returns the single-host network for ip.
func IPToCidr(ip net.IP) *net.IPNet {
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	bits := len(ip) * 8
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)}
}

// ResolveCidrs converts multiple cidrs/ips/domains into cidr
func ResolveCidrs(names []string) ([]Target, error) {
	cidrs := []Target{}
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
func ResolveCidr(name string) ([]Target, error) {
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
		return []Target{{Network: ipnet, Port: port}}, nil
	}

	if ip := net.ParseIP(toResolve); ip != nil {
		return []Target{{Network: IPToCidr(ip), Port: port}}, nil
	}

	addrs, err := LookupIP(toResolve)
	if err != nil {
		return nil, err
	}

	cidrs := []Target{}
	for _, addr := range addrs {
		cidrs = append(cidrs, Target{Network: IPToCidr(addr), Port: port, FromDNS: true})
	}
	return cidrs, nil
}

func LookupIP(host string) ([]net.IP, error) {
	if result := mock.On("LookupIP"); result != nil {
		return result.([]net.IP), nil
	}
	return net.LookupIP(host)
}
