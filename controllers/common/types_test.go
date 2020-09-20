package common

import (
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
	"github.com/stretchr/testify/assert"
	"testing"
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
