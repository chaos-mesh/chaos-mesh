// Copyright 2019 PingCAP, Inc.
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
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ghodss/yaml"

	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
)

var log = ctrl.Log.WithName("inject-webhook")

var (
	// ErrMissingName ..
	ErrMissingName = fmt.Errorf(`name field is required for an injection config`)
	// ErrNoConfigurationLoaded ..
	ErrNoConfigurationLoaded = fmt.Errorf(`at least one config must be present in the --config-directory`)
)

const (
	annotationNamespaceDefault = "admission-webhook.pingcap.com"
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

// InjectionConfig is a specific instance of a injected config, for a given annotation
type InjectionConfig struct {
	Name                  string               `json:"name"`
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
	AnnotationNamespace string                      `yaml:"annotationNamespace"`
	Injections          map[string]*InjectionConfig `yaml:"injections"`
}

// LoadConfigDirectory loads all configs in a directory and returns the Config
func LoadConfigDirectory(path string) (*Config, error) {
	cfg := Config{
		Injections: map[string]*InjectionConfig{},
	}
	glob := filepath.Join(path, "*.yaml")
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}
	for _, p := range matches {
		c, err := LoadInjectionConfigFromFilePath(p)
		if err != nil {
			log.Error(err, "Error reading injection config", "from", p)
			return nil, err
		}
		cfg.Injections[c.Name] = c
	}

	if len(cfg.Injections) == 0 {
		return nil, ErrNoConfigurationLoaded
	}

	if cfg.AnnotationNamespace == "" {
		cfg.AnnotationNamespace = annotationNamespaceDefault
	}

	log.V(2).Info("Loaded injection configs", "len", len(cfg.Injections), "from", glob)

	return &cfg, nil
}

func (c *Config) RequestAnnotationKey() string {
	return c.AnnotationNamespace + "/request"
}

func (c *Config) StatusAnnotationKey() string {
	return c.AnnotationNamespace + "/status"
}

// GetRequestedConfig returns the InjectionConfig given a requested key
func (c *Config) GetRequestedConfig(key string) (*InjectionConfig, error) {
	c.RLock()
	defer c.RUnlock()
	k := strings.ToLower(key)
	i, ok := c.Injections[k]
	if !ok {
		return nil, fmt.Errorf("no required config found for annotation %s", key)
	}
	return i, nil
}

// LoadInjectionConfigFromFilePath returns a InjectionConfig given a yaml file on disk
func LoadInjectionConfigFromFilePath(configFile string) (*InjectionConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("error loading injection config from file %s: %s", configFile, err.Error())
	}
	defer f.Close()

	log.V(3).Info("Loading injection config", "file", configFile)
	return LoadInjectionConfig(f)
}

// LoadInjectionConfig takes an io.Reader and parses out an injectionconfig
func LoadInjectionConfig(reader io.Reader) (*InjectionConfig, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var cfg InjectionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Name == "" {
		return nil, ErrMissingName
	}

	log.V(3).Info("Loaded injection configx", "name", cfg.Name, "sha256sum", sha256.Sum256(data))

	return &cfg, nil
}
