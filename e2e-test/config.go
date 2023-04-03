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

package test

// Config describe the basic config for the operator test
type Config struct {
	ChartDir             string
	ManifestDir          string
	Tag                  string
	ManagerImageRegistry string
	ManagerImage         string
	ManagerTag           string
	DaemonImageRegistry  string
	DaemonImage          string
	DaemonTag            string
	E2EImage             string
	PauseImage           string
	ChaosCoreDNSImage    string
	InstallChaosMesh     bool
	EnableDashboard      bool
}

// NewDefaultConfig describe the default configuration for operator test
func NewDefaultConfig() *Config {
	return &Config{
		ChartDir:             "/charts",
		ManifestDir:          "/manifests",
		Tag:                  "e2e",
		ManagerImageRegistry: "ghcr.io",
		ManagerImage:         "chaos-mesh/chaos-mesh",
		ManagerTag:           "latest",
		DaemonImageRegistry:  "ghcr.io",
		DaemonImage:          "chaos-mesh/chaos-daemon",
		DaemonTag:            "latest",
		E2EImage:             "ghcr.io/chaos-mesh/e2e-helper:latest",
		ChaosCoreDNSImage:    "ghcr.io/chaos-mesh/chaos-coredns:latest",
		PauseImage:           "gcr.io/google-containers/pause:latest",
		InstallChaosMesh:     false,
		EnableDashboard:      false,
	}
}
