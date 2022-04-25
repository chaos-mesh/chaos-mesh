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

package client

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/portforward"
)

const (
	CtrlServerPort = 10082
)

// CommonRestClientGetter is used for non-e2e test environment.
// It's basically do the same thing as genericclioptions.ConfigFlags, but it load rest config from incluster or .kubeconfig file
type CommonRestClientGetter struct {
	*genericclioptions.ConfigFlags
}

func NewCommonRestClientGetter() *CommonRestClientGetter {
	innerConfigFlags := genericclioptions.NewConfigFlags(false)
	return &CommonRestClientGetter{innerConfigFlags}
}

func (it *CommonRestClientGetter) ToRESTConfig() (*rest.Config, error) {
	return config.GetConfig()
}

func ForwardSvcPorts(ctx context.Context, ns, svc string, port uint16) (context.CancelFunc, uint16, error) {
	commonRestClientGetter := NewCommonRestClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to create logger")
	}
	fw, err := portforward.NewPortForwarder(ctx, commonRestClientGetter, false, logger)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to create port forwarder")
	}
	_, localPort, pfCancel, err := portforward.ForwardOnePort(fw, ns, svc, port)

	// disable error handler in k8s runtime to prevent complaining from port forwarder
	DisableRuntimeErrorHandler()
	return pfCancel, localPort, err
}

func ForwardCtrlServer(ctx context.Context, ns, managerSvc string) (context.CancelFunc, uint16, error) {
	return ForwardSvcPorts(ctx, ns, "svc/"+managerSvc, CtrlServerPort)
}
