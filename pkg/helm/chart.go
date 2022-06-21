package helm

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"io"
	"net/http"
	"os"
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
