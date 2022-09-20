// Copyright 2022 Chaos Mesh Authors.
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

//go:build integration

package helm

import (
	"context"
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/cli"

	"github.com/chaos-mesh/chaos-mesh/pkg/log"
)

func ExampleHelmClient_UpgradeOrInstall() {
	chart, err := FetchChaosMeshChart(context.Background(), "2.2.0")
	if err != nil {
		panic(err)
	}

	settings := cli.New()
	restClientGetter := settings.RESTClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		panic(err)
	}
	client, err := NewHelmClient(restClientGetter, logger)
	if err != nil {
		panic(err)
	}
	_, err = client.UpgradeOrInstall(
		"chaos-mesh",
		"chaos-mesh-in-remote-cluster",
		chart,
		nil,
	)
	if err != nil {
		panic(err)
	}
	// Output:
}

func ExampleHelmClient_GetRelease() {
	settings := cli.New()
	restClientGetter := settings.RESTClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		panic(err)
	}
	client, err := NewHelmClient(restClientGetter, logger)
	if err != nil {
		panic(err)
	}
	release, err := client.GetRelease(
		"chaos-mesh",
		"chaos-mesh-in-remote-cluster",
	)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stderr, release.Name)
	// Output:
}

func ExampleHelmClient_UninstallRelease() {
	settings := cli.New()
	restClientGetter := settings.RESTClientGetter()
	logger, err := log.NewDefaultZapLogger()
	if err != nil {
		panic(err)
	}
	client, err := NewHelmClient(restClientGetter, logger)
	if err != nil {
		panic(err)
	}
	_, err = client.UninstallRelease(
		"chaos-mesh",
		"chaos-mesh-in-remote-cluster",
	)
	if err != nil {
		panic(err)
	}
	// Output:
}
