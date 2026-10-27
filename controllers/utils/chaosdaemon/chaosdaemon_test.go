// Copyright 2026 Chaos Mesh Authors.
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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	controllerconfig "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

func TestFindDaemonIPByNodePreservesPodLookup(t *testing.T) {
	previousConfig := controllerconfig.ControllerCfg
	controllerconfig.ControllerCfg = &config.ChaosControllerConfig{Namespace: "chaos-mesh"}
	t.Cleanup(func() { controllerconfig.ControllerCfg = previousConfig })
	scheme := runtime.NewScheme()
	require.NoError(t, discoveryv1.AddToScheme(scheme))
	nodeA, nodeB := "node-a", "node-b"
	kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&discoveryv1.EndpointSlice{
		ObjectMeta:  metav1.ObjectMeta{Namespace: "chaos-mesh", Name: "chaos-daemon-test"},
		AddressType: discoveryv1.AddressTypeIPv4,
		Endpoints: []discoveryv1.Endpoint{
			{NodeName: &nodeB, Addresses: []string{"10.0.0.2"}},
			{NodeName: &nodeA, Addresses: []string{"10.0.0.1"}},
			{NodeName: &nodeA, Addresses: []string{"10.0.0.3"}},
		},
	}).Build()
	builder := &ClientBuilder{Reader: kube}
	address, err := builder.FindDaemonIPByNode(context.Background(), nodeA)
	require.NoError(t, err)
	require.Equal(t, "10.0.0.1", address)
	address, err = builder.FindDaemonIP(context.Background(), &v1.Pod{Spec: v1.PodSpec{NodeName: nodeA}})
	require.NoError(t, err)
	require.Equal(t, "10.0.0.1", address)
	_, err = builder.FindDaemonIPByNode(context.Background(), "missing-node")
	require.Error(t, err)
}

func TestFindDaemonIPByNodePreservesListError(t *testing.T) {
	previousConfig := controllerconfig.ControllerCfg
	controllerconfig.ControllerCfg = &config.ChaosControllerConfig{Namespace: "chaos-mesh"}
	t.Cleanup(func() { controllerconfig.ControllerCfg = previousConfig })
	want := errors.New("endpoint access denied")
	builder := &ClientBuilder{Reader: &endpointListErrorReader{err: want}}
	_, err := builder.FindDaemonIPByNode(context.Background(), "node-a")
	require.ErrorIs(t, err, want)
}

type endpointListErrorReader struct {
	client.Reader
	err error
}

func (r *endpointListErrorReader) List(context.Context, client.ObjectList, ...client.ListOption) error {
	return r.err
}
