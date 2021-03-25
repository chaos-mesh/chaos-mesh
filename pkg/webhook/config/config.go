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

package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/ghodss/yaml"

	corev1 "k8s.io/api/core/v1"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

var (
	errMissingName         = fmt.Errorf(`name field is required for template args config`)
	errMissingTemplateName = fmt.Errorf(`template field is required for template args config`)
)

const (
	annotationNamespaceDefault = "admission-webhook.chaos-mesh.org"
)

// ExecAction describes a "run in container" action.
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command  is root ('/') in the container's filesystem. The command is simply exec'd, it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work. To use
	// a shell, you need to explicitly call out to that shell.
	// Exit status of 0 is treated as live/healthy and non-zero is unhealthy.
	// +optional
	Command []string `json:"command,omitempty"`
}

// InjectionConfig is a specific instance of an injected config, for a given annotation
type InjectionConfig struct {
	Name string
	// Selector is used to select pods that are used to inject sidecar.
	Selector *v1alpha1.PodSelectorSpec

	Containers            []corev1.Container   `json:"containers"`
	Volumes               []corev1.Volume      `json:"volumes"`
	Environment           []corev1.EnvVar      `json:"env"`
	VolumeMounts          []corev1.VolumeMount `json:"volumeMounts"`
	HostAliases           []corev1.HostAlias   `json:"hostAliases"`
	InitContainers        []corev1.Container   `json:"initContainers"`
	ShareProcessNamespace bool                 `json:"shareProcessNamespace"`
	// PostStart is called after a container is created first.
	// If the handler fails, the containers will failed.
	// Key defines for the name of deployment container.
	// Value defines for the Commands for stating container.
	// +optional
	PostStart map[string]ExecAction `json:"postStart,omitempty"`
}

// Config is a struct indicating how a given injection should be configured
type Config struct {
	sync.RWMutex
	AnnotationNamespace string
	Injections          map[string][]*InjectionConfig
}

// TemplateArgs is a set of arguments to render template
type TemplateArgs struct {
	Namespace string
	Name      string `yaml:"name"`
	// Name of the template
	Template  string            `yaml:"template"`
	Arguments map[string]string `yaml:"arguments"`
	// Selector is used to select pods that are used to inject sidecar.
	Selector *v1alpha1.PodSelectorSpec `json:"selector,omitempty"`
}

// NewConfigWatcherConf creates a configuration for watcher
func NewConfigWatcherConf() *Config {
	return &Config{
		AnnotationNamespace: annotationNamespaceDefault,
		Injections:          make(map[string][]*InjectionConfig),
	}
}

func (c *Config) RequestAnnotationKey() string {
	return c.AnnotationNamespace + "/request"
}

func (c *Config) StatusAnnotationKey() string {
	return c.AnnotationNamespace + "/status"
}

func (c *Config) RequestInitAnnotationKey() string {
	return c.AnnotationNamespace + "/init-request"
}

// GetRequestedConfig returns the InjectionConfig given a requested key
func (c *Config) GetRequestedConfig(namespace, key string) (*InjectionConfig, error) {
	c.RLock()
	defer c.RUnlock()

	if _, ok := c.Injections[namespace]; !ok {
		return nil, fmt.Errorf("no injection config at ns %s", namespace)
	}

	for _, conf := range c.Injections[namespace] {
		if key == conf.Name {
			return conf, nil
		}
	}

	return nil, fmt.Errorf("no injection config found for key %s at ns %s", key, namespace)
}

// LoadTemplateArgs takes an io.Reader and parses out an template args
func LoadTemplateArgs(reader io.Reader) (*TemplateArgs, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var cfg TemplateArgs
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Name == "" {
		return nil, errMissingName
	}

	if cfg.Template == "" {
		return nil, errMissingTemplateName
	}

	return &cfg, nil
}

// ReplaceInjectionConfigs will update the injection configs.
func (c *Config) ReplaceInjectionConfigs(updatedConfigs map[string][]*InjectionConfig) {
	c.Lock()
	defer c.Unlock()
	c.Injections = updatedConfigs
}
