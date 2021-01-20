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

package common

import (
	"os"
	"path/filepath"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// CommonRestClientGetter is used for non-e2e test environment.
// It's basically do the same thing as genericclioptions.ConfigFlags, but it load rest config from incluster or .kubeconfig file
type CommonRestClientGetter struct {
	*genericclioptions.ConfigFlags
}

func NewCommonRestClientGetter() *CommonRestClientGetter {
	innerConfigFlags := genericclioptions.NewConfigFlags(false)
	return &CommonRestClientGetter{innerConfigFlags}
}

func (it *CommonRestClientGetter) ToRESTConfig() (*rest.Config, error) {
	return loadRestConfig()
}

func loadRestConfig() (*rest.Config, error) {
	var kubeconfig string
	inClusterConfig, err := rest.InClusterConfig()
	if err == nil {
		return inClusterConfig, nil
	}
	kubeconfig = os.Getenv("KUBECONFIG")
	if len(kubeconfig) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
