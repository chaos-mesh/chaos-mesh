// Copyright 2019 Chaos Mesh Authors.
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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

func TestTargetNamespaceShouldSetWhileClusterScopedIsFalse(t *testing.T) {
	c := config.ChaosControllerConfig{
		WatcherConfig: &watcher.Config{
			ClusterScoped:     false,
			TemplateNamespace: "",
			TargetNamespace:   "",
			TemplateLabels:    nil,
			ConfigLabels:      nil,
		},
		ClusterScoped:   false,
		TargetNamespace: "",
	}
	err := validate(&c)
	assert.NotNil(t, err)
}

func TestWatcherConfigIsAlwaysRequired(t *testing.T) {
	c := config.ChaosControllerConfig{
		WatcherConfig:   nil,
		ClusterScoped:   true,
		TargetNamespace: "",
	}
	err := validate(&c)
	assert.NotNil(t, err)
}

func TestClusterScopedInConfigAndClusterScopedInWatcherConfigShouldKeepConstant(t *testing.T) {
	c := config.ChaosControllerConfig{
		WatcherConfig: &watcher.Config{
			ClusterScoped:     false,
			TemplateNamespace: "",
			TargetNamespace:   "",
			TemplateLabels:    nil,
			ConfigLabels:      nil,
		},
		ClusterScoped:   true,
		TargetNamespace: "",
	}
	err := validate(&c)
	assert.NotNil(t, err)
}

func TestNsInConfigAndNsInWatcherConfigShouldKeepConstant(t *testing.T) {
	c := config.ChaosControllerConfig{
		WatcherConfig: &watcher.Config{
			ClusterScoped:   false,
			TargetNamespace: "ns1",
		},
		ClusterScoped:   false,
		TargetNamespace: "ns2",
	}
	err := validate(&c)
	assert.NotNil(t, err)
}
