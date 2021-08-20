// Copyright 2021 Chaos Mesh Authors.
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

package common

import (
	"context"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/pkg/portforward"
)

const (
	DefaultManagerNamespace = "chaos-testing"
	ManagerSvc              = "svc/chaos-mesh-controller-manager"
	CtrlServerPort          = 10082
)

func forwardPorts(ctx context.Context, pod v1.Pod, port uint16) (context.CancelFunc, uint16, error) {
	return ForwardSvcPorts(ctx, pod.Namespace, pod.Name, port)
}

func ForwardSvcPorts(ctx context.Context, ns, svc string, port uint16) (context.CancelFunc, uint16, error) {
	commonRestClientGetter := NewCommonRestClientGetter()
	fw, err := portforward.NewPortForwarder(ctx, commonRestClientGetter, false)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to create port forwarder")
	}
	_, localPort, pfCancel, err := portforward.ForwardOnePort(fw, ns, svc, port)
	return pfCancel, localPort, err
}

func ForwardCtrlServer(ctx context.Context, ns *string) (context.CancelFunc, uint16, error) {
	if ns == nil {
		ns = new(string)
		*ns = DefaultManagerNamespace
	}

	return ForwardSvcPorts(ctx, *ns, ManagerSvc, CtrlServerPort)
}
