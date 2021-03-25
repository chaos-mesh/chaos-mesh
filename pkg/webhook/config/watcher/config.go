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

package watcher

import (
	"github.com/pingcap/errors"
)

// Config is a configuration struct for the Watcher type
type Config struct {
	// ClusterScoped means control Chaos Object in cluster level(all namespace);
	ClusterScoped bool `envconfig:"CLUSTER_SCOPED" default:"true"`
	// TemplateNamespace is the namespace which holds the template configmap.
	// If controller-manager is running with in-cluster mode. If is set to empty string, it will be overwrite to namespace which the pod belongs.
	TemplateNamespace string `envconfig:"TEMPLATE_NAMESPACE" default:""`
	// TargetNamespace means configmaps in this namespace will be controlled by this controller.
	// It SHOULD be the same with TargetNamespace in config.ChaosControllerConfig while clusterScoped is false.
	TargetNamespace string `envconfig:"TARGET_NAMESPACE" default:""`
	// TemplateLabels is label pairs used to discover common templates in Kubernetes. These should be key1:value[,key2:val2,...]
	TemplateLabels map[string]string `envconfig:"TEMPLATE_LABELS"`
	// ConfigLabels is label pairs used to discover ConfigMaps in Kubernetes. These should be key1:value[,key2:val2,...]
	ConfigLabels map[string]string `envconfig:"CONFIGMAP_LABELS"`
}

// NewConfig returns a new initialized Config
func NewConfig() *Config {
	return &Config{
		ClusterScoped:     true,
		TemplateNamespace: "",
		TargetNamespace:   "",
		TemplateLabels:    map[string]string{},
		ConfigLabels:      map[string]string{},
	}
}

// Verify will verify the parameter configuration is correct
func (c *Config) Verify() error {
	if len(c.TemplateLabels) == 0 {
		return errors.New("envconfig:\"TEMPLATE_LABELS\" template labels must be set")
	}
	if len(c.ConfigLabels) == 0 {
		return errors.New("envconfig:\"CONFIGMAP_LABELS\" conf labels must be set")
	}
	if !c.ClusterScoped && len(c.TargetNamespace) == 0 {
		return errors.New("envconfig:\"TARGET_NAMESPACE\" conf labels must be set while CLUSTER_SCOPED is false")
	}
	return nil
}
