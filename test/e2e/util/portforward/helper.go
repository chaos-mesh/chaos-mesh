// Copyright 2019 PingCAP, Inc.
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

package portforward

import (
	"context"
	"errors"
	"fmt"
)

// A helper utility to forward one port of Kubernetes resource.
func ForwardOnePort(fw PortForward, ns, resource string, port uint16) (string, uint16, context.CancelFunc, error) {
	ports := []string{fmt.Sprintf("0:%d", port)}
	forwardedPorts, cancel, err := fw.Forward(ns, resource, []string{"127.0.0.1"}, ports)
	if err != nil {
		return "", 0, nil, err
	}
	var localPort uint16
	var found bool
	for _, p := range forwardedPorts {
		if p.Remote == port {
			localPort = p.Local
			found = true
		}
	}
	if !found {
		cancel()
		return "", 0, nil, errors.New("unexpected error")
	}
	return "127.0.0.1", localPort, cancel, nil
}
