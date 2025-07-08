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

package chaosdaemon

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var log = ctrl.Log.WithName("controller-chaos-daemon-client-utils")

func findIPOnEndpointSlice(e *discoveryv1.EndpointSliceList, nodeName string) string {
	for _, endpointSlice := range e.Items {
		// Only get the endpoint slice of the current chaos-daemon pod.
		if strings.HasPrefix(endpointSlice.ObjectMeta.Name, "chaos-daemon-") {
			for _, ep := range endpointSlice.Endpoints {
				if ep.NodeName != nil && *ep.NodeName == nodeName {
					return ep.Addresses[0]
				}
			}
		}
	}

	return ""
}

type ChaosDaemonClientBuilder struct {
	client.Reader
}

func (b *ChaosDaemonClientBuilder) FindDaemonIP(ctx context.Context, pod *v1.Pod) (string, error) {
	nodeName := pod.Spec.NodeName
	log.Info("Creating client to chaos-daemon", "node", nodeName)

	ns := config.ControllerCfg.Namespace
	var endpointSliceList discoveryv1.EndpointSliceList
	err := b.Reader.List(ctx, &endpointSliceList, client.InNamespace(ns))
	if err != nil {
		return "", err
	}

	daemonIP := findIPOnEndpointSlice(&endpointSliceList, nodeName)
	if len(daemonIP) == 0 {
		return "", errors.Errorf("cannot find daemonIP on node %s", nodeName)
	}

	return daemonIP, nil
}

// Build will construct a ChaosDaemonClient
// The `id` parameter is the namespacedName of current handling resource,
// which will be printed in the log of the chaos-daemon
func (b *ChaosDaemonClientBuilder) Build(ctx context.Context, pod *v1.Pod, id *types.NamespacedName) (chaosdaemonclient.ChaosDaemonClientInterface, error) {
	if cli := mock.On("MockChaosDaemonClient"); cli != nil {
		return cli.(chaosdaemonclient.ChaosDaemonClientInterface), nil
	}
	if err := mock.On("NewChaosDaemonClientError"); err != nil {
		return nil, err.(error)
	}

	daemonIP, err := b.FindDaemonIP(ctx, pod)
	if err != nil {
		return nil, err
	}
	builder := grpcUtils.Builder(daemonIP, config.ControllerCfg.ChaosDaemonPort).WithDefaultTimeout()
	if config.ControllerCfg.TLSConfig.ChaosMeshCACert != "" {
		builder.TLSFromFile(config.ControllerCfg.TLSConfig.ChaosMeshCACert, config.ControllerCfg.TLSConfig.ChaosDaemonClientCert, config.ControllerCfg.TLSConfig.ChaosDaemonClientKey)
	} else {
		builder.Insecure()
	}

	if id != nil {
		builder = builder.WithNamespacedName(*id)
	}

	cc, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return chaosdaemonclient.New(cc), nil
}

type ChaosDaemonClientBuilderParams struct {
	fx.In

	NoCacheReader           client.Reader `name:"no-cache"`
	ControlPlaneCacheReader client.Reader `name:"control-plane-cache" optional:"true"`
}

func New(params ChaosDaemonClientBuilderParams) *ChaosDaemonClientBuilder {
	var reader client.Reader
	if params.ControlPlaneCacheReader != nil {
		reader = params.ControlPlaneCacheReader
	} else {
		reader = params.NoCacheReader
	}
	return &ChaosDaemonClientBuilder{
		Reader: reader,
	}
}
