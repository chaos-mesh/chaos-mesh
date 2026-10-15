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

package clusterregistry

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestWithClientDoesNotBlockOtherClusters(t *testing.T) {
	registry := New(logr.Discard(), nil)
	registry.clusters["a"] = &remoteCluster{}
	registry.clusters["b"] = &remoteCluster{}
	entered := make(chan struct{})
	release := make(chan struct{})
	aDone := make(chan error, 1)
	go func() {
		aDone <- registry.WithClient("a", func(client.Client) error { close(entered); <-release; return nil })
	}()
	<-entered
	bDone := make(chan error, 1)
	go func() { bDone <- registry.WithClient("b", func(client.Client) error { return nil }) }()
	progressed := false
	select {
	case err := <-bDone:
		require.NoError(t, err)
		progressed = true
	case <-time.After(time.Second):
	}
	close(release)
	require.NoError(t, <-aDone)
	if !progressed {
		require.NoError(t, <-bDone)
	}
	require.True(t, progressed, "remote B must remain available while an A callback is blocked")
}
