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

package test

// Config describe the basic config for the operator test
type Config struct {
	ChartDir         string
	ManifestDir      string
	Tag              string
	ManagerImage     string
	ManagerTag       string
	DaemonImage      string
	DaemonTag        string
	E2EImage         string
	ChaosDNSImage    string
	InstallChaosMesh bool
	EnableDashboard  bool
}

// NewDefaultConfig describe the default configuration for operator test
func NewDefaultConfig() *Config {
	return &Config{
		ChartDir:         "/charts",
		ManifestDir:      "/manifests",
		Tag:              "e2e",
		ManagerImage:     "localhost:5000/pingcap/chaos-mesh",
		ManagerTag:       "latest",
		DaemonImage:      "localhost:5000/pingcap/chaos-daemon",
		DaemonTag:        "latest",
		E2EImage:         "localhost:5000/pingcap/e2e-helper:latest",
		ChaosDNSImage:    "localhost:5000/pingcap/chaos-dns:latest",
		InstallChaosMesh: false,
		EnableDashboard:  false,
	}
}
