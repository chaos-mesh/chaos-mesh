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

package server

import (
	"context"
	"strings"

	"github.com/pingcap/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

// GetIpset returns result of ipset list
func (r *Resolver) GetIpset(ctx context.Context, obj *v1.Pod) (string, error) {
	cmd := "ipset list"
	return r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.NetNS)
}

// GetIpset returns result of tc qdisc list
func (r *Resolver) GetTcQdisc(ctx context.Context, obj *v1.Pod) ([]string, error) {
	cmd := "tc qdisc list"
	rules, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.NetNS)
	if err != nil {
		return nil, errors.Wrapf(err, "exec `%s`", cmd)
	}
	return strings.Split(rules, "\n"), nil
}

// GetIptables returns result of iptables --list
func (r *Resolver) GetIptables(ctx context.Context, obj *v1.Pod) ([]string, error) {
	cmd := "iptables --list"
	rules, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.NetNS)
	if err != nil {
		return nil, errors.Wrapf(err, "exec `%s`", cmd)
	}
	return strings.Split(rules, "\n"), nil
}

// cleanTcs returns actually cleaned devices
func (r *Resolver) cleanTcs(ctx context.Context, obj *v1.Pod, devices []string) ([]string, error) {
	var cleaned []string
	for _, device := range devices {
		cmd := "tc qdisc del dev " + device + " root"
		_, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.NetNS)
		if err != nil {
			return cleaned, errors.Wrapf(err, "exec `%s`", cmd)
		}
		cleaned = append(cleaned, device)
	}
	return cleaned, nil
}

// cleanIptables returns actually cleaned chains
func (r *Resolver) cleanIptables(ctx context.Context, obj *v1.Pod, chains []string) ([]string, error) {
	var cleaned []string
	for _, chain := range chains {
		cmd := "iptables -F " + chain
		_, err := r.ExecBypass(ctx, obj, cmd, bpm.PidNS, bpm.NetNS)
		if err != nil {
			return cleaned, errors.Wrapf(err, "exec `%s`", cmd)
		}
		cleaned = append(cleaned, chain)
	}
	return cleaned, nil
}
