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
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type HelmClient struct {
	restClientGetter genericclioptions.RESTClientGetter
	logger           logr.Logger
}

func NewHelmClient(restClientGetter genericclioptions.RESTClientGetter, logger logr.Logger) (*HelmClient, error) {
	return &HelmClient{restClientGetter: restClientGetter, logger: logger}, nil
}

func (h HelmClient) spawnConfigurationWithNamespace(namespace string) (*action.Configuration, error) {
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "create helm registry client")
	}
	kubeclient := kube.New(h.restClientGetter)
	if err != nil {
		return nil, errors.Wrap(err, "create kubernetes client set")
	}
	clientset, err := kubeclient.Factory.KubernetesClientSet()
	if err != nil {
		return nil, errors.Wrap(err, "create kubernetes client set")
	}
	secretInterface := clientset.CoreV1().Secrets(namespace)
	helmConfiguration := action.Configuration{
		Releases:         storage.Init(driver.NewSecrets(secretInterface)),
		KubeClient:       kubeclient,
		Capabilities:     chartutil.DefaultCapabilities,
		RegistryClient:   registryClient,
		RESTClientGetter: h.restClientGetter,
		Log: func(format string, v ...interface{}) {
			h.logger.Info(fmt.Sprintf(format, v...))
		},
	}
	return &helmConfiguration, nil
}

func (h *HelmClient) GetRelease(namespace string, releaseName string) (*release.Release, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	getAction := action.NewGet(configurationWithNamespace)
	result, err := getAction.Run(releaseName)
	if err != nil {
		return nil, errors.Wrapf(err, "get release %s, in namespace %s", releaseName, namespace)
	}
	return result, nil
}

func (h *HelmClient) InstallRelease(namespace string, releaseName string, chart *chart.Chart, values map[string]interface{}) (*release.Release, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	actionInstall := action.NewInstall(configurationWithNamespace)
	actionInstall.ReleaseName = releaseName
	actionInstall.Namespace = namespace
	actionInstall.CreateNamespace = true
	result, err := actionInstall.Run(chart, values)
	if err != nil {
		return nil, errors.Wrapf(err, "install release %s, with chart %s, with values %v", releaseName, chart.Metadata.Name, values)
	}
	return result, nil
}

func (h *HelmClient) UpgradeRelease(namespace string, releaseName string, chart *chart.Chart, values map[string]interface{}) (*release.Release, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	actionUpgrade := action.NewUpgrade(configurationWithNamespace)
	actionUpgrade.Namespace = namespace
	result, err := actionUpgrade.Run(releaseName, chart, values)
	if err != nil {
		return nil, errors.Wrapf(err, "upgrade release %s, with chart %s, with values %v", releaseName, chart.Metadata.Name, values)
	}
	return result, nil
}

func (h *HelmClient) UninstallRelease(namespace string, releaseName string) (*release.UninstallReleaseResponse, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	uninstallAction := action.NewUninstall(configurationWithNamespace)
	response, err := uninstallAction.Run(releaseName)
	if err != nil {
		return nil, errors.Wrapf(err, "uninstall release %s, in namespace %s", releaseName, namespace)
	}
	return response, nil
}
