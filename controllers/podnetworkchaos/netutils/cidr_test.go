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
	"reflect"
	"testing"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

func TestIPToCidr(t *testing.T) {
	tests := []struct {
		ip   string
		want string
	}{
		{"1.2.3.4", "1.2.3.4/32"},
		{"2001:db8::1", "2001:db8::1/128"},
		{"::1", "::1/128"},
	}
	for _, tt := range tests {
		if got := IPToCidr(tt.ip); got != tt.want {
			t.Errorf("IPToCidr(%q) = %q, want %q", tt.ip, got, tt.want)
		}
	}
}

func TestResolveCidr(t *testing.T) {
	// Mock returns one IPv4 and one IPv6 address to cover dual-stack resolution.
	ipv6 := net.ParseIP("2001:db8::1")
	defer mock.With("LookupIP", []net.IP{{1, 1, 1, 1}, {2, 2, 2, 2}, ipv6})()

	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []v1alpha1.CidrAndPort
		wantErr bool
	}{
		{
			name: "ip address",
			args: args{name: "1.1.1.1"},
			want: []v1alpha1.CidrAndPort{{Cidr: "1.1.1.1/32"}},
		},
		{
			name: "ip address and port",
			args: args{name: "1.1.1.1:80"},
			want: []v1alpha1.CidrAndPort{{Cidr: "1.1.1.1/32", Port: 80}},
		},
		{
			name: "ipv6 address",
			args: args{name: "2001:db8::1"},
			want: []v1alpha1.CidrAndPort{{Cidr: "2001:db8::1/128"}},
		},
		{
			name: "ipv6 address and port",
			args: args{name: "[2001:db8::1]:80"},
			want: []v1alpha1.CidrAndPort{{Cidr: "2001:db8::1/128", Port: 80}},
		},
		{
			name: "ipv6 loopback and port",
			args: args{name: "[::1]:443"},
			want: []v1alpha1.CidrAndPort{{Cidr: "::1/128", Port: 443}},
		},
		{
			name: "ipv6 subnet",
			args: args{name: "2001:db8::/32"},
			want: []v1alpha1.CidrAndPort{{Cidr: "2001:db8::/32"}},
		},
		{
			name: "subnet",
			args: args{name: "0.0.0.0/24"},
			want: []v1alpha1.CidrAndPort{{Cidr: "0.0.0.0/24"}},
		},
		{
			name: "subnet and port",
			args: args{name: "0.0.0.0/24:443"},
			want: []v1alpha1.CidrAndPort{{Cidr: "0.0.0.0/24", Port: 443}},
		},
		{
			name: "hostname",
			args: args{name: "example.com"},
			want: []v1alpha1.CidrAndPort{{Cidr: "1.1.1.1/32"}, {Cidr: "2.2.2.2/32"}, {Cidr: "2001:db8::1/128"}},
		},
		{
			name: "hostname and port",
			args: args{name: "example.com:80"},
			want: []v1alpha1.CidrAndPort{{Cidr: "1.1.1.1/32", Port: 80}, {Cidr: "2.2.2.2/32", Port: 80}, {Cidr: "2001:db8::1/128", Port: 80}},
		},
		{
			name:    "missing port",
			args:    args{name: "1.1.1.1:"},
			wantErr: true,
		},
		{
			name:    "port out of range",
			args:    args{name: "1.1.1.1:65536"},
			wantErr: true,
		},
		{
			name:    "not a port",
			args:    args{name: "1.1.1.1:foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveCidr(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveCidr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResolveCidr() got = %v, want %v", got, tt.want)
			}
		})
	}
}
