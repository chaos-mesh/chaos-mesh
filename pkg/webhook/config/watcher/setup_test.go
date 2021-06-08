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

package watcher

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"webhook config watcher",
		[]Reporter{printer.NewlineReporter{}})
}

func MockClusterConfig() (*rest.Config, error) {
	return &rest.Config{
		Host:            "https://testhost:9527",
		TLSClientConfig: rest.TLSClientConfig{},
		BearerToken:     "testToken",
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
	}, nil
}
