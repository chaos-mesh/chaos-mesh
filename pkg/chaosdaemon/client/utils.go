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

package client

import (
	"context"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	chaosdaemon "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	grpcUtils "github.com/chaos-mesh/chaos-mesh/pkg/grpc"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var log = ctrl.Log.WithName("chaos-daemon-client-utils")

// ChaosDaemonClientInterface represents the ChaosDaemonClient, it's used to simply unit test
type ChaosDaemonClientInterface interface {
	chaosdaemon.ChaosDaemonClient
	Close() error
}

// GrpcChaosDaemonClient would act like chaosdaemon.ChaosDaemonClient with a Close method
type GrpcChaosDaemonClient struct {
	chaosdaemon.ChaosDaemonClient
	conn *grpc.ClientConn
}

func (c *GrpcChaosDaemonClient) Close() error {
	return c.conn.Close()
}

func FindDaemonIP(ctx context.Context, c client.Client, pod *v1.Pod) (string, error) {
	nodeName := pod.Spec.NodeName
	log.Info("Creating client to chaos-daemon", "node", nodeName)

	ns := config.ControllerCfg.Namespace
	var endpoints v1.Endpoints
	err := c.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      "chaos-daemon",
	}, &endpoints)
	if err != nil {
		return "", err
	}

	daemonIP := findIPOnEndpoints(&endpoints, nodeName)
	if len(daemonIP) == 0 {
		return "", errors.Errorf("cannot find daemonIP on node %s in related Endpoints %v", nodeName, endpoints)
	}

	return daemonIP, nil
}

func findIPOnEndpoints(e *v1.Endpoints, nodeName string) string {
	for _, subset := range e.Subsets {
		for _, addr := range subset.Addresses {
			if addr.NodeName != nil && *addr.NodeName == nodeName {
				return addr.IP
			}
		}
	}

	return ""
}

// NewChaosDaemonClient would create ChaosDaemonClient
func NewChaosDaemonClient(ctx context.Context, c client.Client, pod *v1.Pod) (ChaosDaemonClientInterface, error) {
	if cli := mock.On("MockChaosDaemonClient"); cli != nil {
		return cli.(ChaosDaemonClientInterface), nil
	}
	if err := mock.On("NewChaosDaemonClientError"); err != nil {
		return nil, err.(error)
	}

	daemonIP, err := FindDaemonIP(ctx, c, pod)
	if err != nil {
		return nil, err
	}

	cc, err := grpcUtils.CreateGrpcConnectionWithAddress(daemonIP, config.ControllerCfg.ChaosDaemonPort, config.ControllerCfg.TLSConfig.ChaosMeshCACert, config.ControllerCfg.TLSConfig.ChaosDaemonClientCert, config.ControllerCfg.TLSConfig.ChaosDaemonClientKey)
	if err != nil {
		return nil, err
	}
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(cc),
		conn:              cc,
	}, nil
}

// NewChaosDaemonClientLocally would create ChaosDaemonClient in localhost
func NewChaosDaemonClientLocally(port int, caCert string, cert string, key string) (ChaosDaemonClientInterface, error) {
	if cli := mock.On("MockChaosDaemonClient"); cli != nil {
		return cli.(ChaosDaemonClientInterface), nil
	}
	if err := mock.On("NewChaosDaemonClientError"); err != nil {
		return nil, err.(error)
	}

	cc, err := grpcUtils.CreateGrpcConnectionWithAddress("localhost", port, caCert, cert, key)
	if err != nil {
		return nil, err
	}
	return &GrpcChaosDaemonClient{
		ChaosDaemonClient: chaosdaemon.NewChaosDaemonClient(cc),
		conn:              cc,
	}, nil
}
