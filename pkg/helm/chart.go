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

package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const ChaosMeshHelmRepo = "https://charts.chaos-mesh.org"

func FetchChaosMeshChart(ctx context.Context, version string) (*chart.Chart, error) {
	tgzPath, err := DownloadChaosMeshChartTgz(ctx, version)
	if err != nil {
		return nil, err
	}
	requestedChart, err := loader.Load(tgzPath)
	if err != nil {
		return nil, errors.Wrapf(err, "load helm chart from %s", tgzPath)
	}
	return requestedChart, nil
}

func DownloadChaosMeshChartTgz(ctx context.Context, version string) (string, error) {
	// TODO: use this context

	url := fmt.Sprintf("%s/chaos-mesh-%s.tgz", ChaosMeshHelmRepo, version)
	response, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "download helm chart from %s", url)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", errors.Wrapf(err, "download helm chart from %s", url)
	}
	target, err := os.CreateTemp("", fmt.Sprintf("chaos-mesh-%s-*.tgz", version))
	if err != nil {
		return "", errors.Wrapf(err, "download helm chart as temp file")
	}
	defer target.Close()
	_, err = io.Copy(target, response.Body)
	if err != nil {
		return "", errors.Wrapf(err, "download helm chart content")
	}
	return target.Name(), nil
}
