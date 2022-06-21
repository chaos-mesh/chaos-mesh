package helm

import (
	"github.com/go-logr/logr"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type HelmClient struct {
	restClientGetter genericclioptions.RESTClientGetter
	logger           logr.Logger

	helmConfiguration *action.Configuration
}

func (h *HelmClient) GetRelease(namespace string, releaseName string) (release.Release, error) {
	//TODO implement me
	panic("implement me")
}

func (h *HelmClient) UpgradeOrInstall(namespace string, releaseName string, chart *chart.Chart, values map[string]interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (h *HelmClient) UninstallRelease(namespace string, releaseName string) error {
	//TODO implement me
	panic("implement me")
}
