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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/kube"
	"helm.sh/helm/v4/pkg/registry"
	releaseapi "helm.sh/helm/v4/pkg/release"
	release "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	// Share field ownership with Helm CLI releases when applying their upgrades.
	kube.ManagedFieldsManager = "helm"
}

type Client struct {
	restClientGetter genericclioptions.RESTClientGetter
	logger           logr.Logger
}

func NewHelmClient(restClientGetter genericclioptions.RESTClientGetter, logger logr.Logger) (*Client, error) {
	return &Client{restClientGetter: restClientGetter, logger: logger}, nil
}

func (h Client) spawnConfigurationWithNamespace(namespace string) (*action.Configuration, error) {
	registryClient, err := registry.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "create helm registry client")
	}
	kubeclient := kube.New(h.restClientGetter)
	clientset, err := kubeclient.Factory.KubernetesClientSet()
	if err != nil {
		return nil, errors.Wrap(err, "create kubernetes client set")
	}
	secretInterface := clientset.CoreV1().Secrets(namespace)
	helmConfiguration := action.NewConfiguration(action.ConfigurationSetLogger(logr.ToSlogHandler(h.logger)))
	helmConfiguration.Releases = storage.Init(driver.NewSecrets(secretInterface))
	helmConfiguration.KubeClient = kubeclient
	helmConfiguration.Capabilities = common.DefaultCapabilities
	helmConfiguration.RegistryClient = registryClient
	helmConfiguration.RESTClientGetter = h.restClientGetter
	return helmConfiguration, nil
}

func (h *Client) GetRelease(namespace string, releaseName string) (*release.Release, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	getAction := action.NewGet(configurationWithNamespace)
	result, err := getAction.Run(releaseName)
	if err != nil {
		return nil, errors.Wrapf(err, "get release %s, in namespace %s", releaseName, namespace)
	}
	return releaseV1(result)
}

func (h *Client) InstallRelease(namespace string, releaseName string, chart *chart.Chart, values map[string]any) (*release.Release, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	actionInstall := action.NewInstall(configurationWithNamespace)
	actionInstall.ReleaseName = releaseName
	actionInstall.Namespace = namespace
	actionInstall.CreateNamespace = true
	actionInstall.WaitStrategy = kube.HookOnlyStrategy
	actionInstall.ServerSideApply = false
	result, err := actionInstall.Run(chart, values)
	if err != nil {
		return nil, errors.Wrapf(err, "install release %s, with chart %s, with values %v", releaseName, chart.Metadata.Name, values)
	}
	return releaseV1(result)
}

func (h *Client) UpgradeRelease(namespace string, releaseName string, chart *chart.Chart, values map[string]any) (*release.Release, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	actionUpgrade := action.NewUpgrade(configurationWithNamespace)
	actionUpgrade.Namespace = namespace
	actionUpgrade.WaitStrategy = kube.HookOnlyStrategy
	result, err := actionUpgrade.Run(releaseName, chart, values)
	if err != nil {
		return nil, errors.Wrapf(err, "upgrade release %s, with chart %s, with values %v", releaseName, chart.Metadata.Name, values)
	}
	return releaseV1(result)
}

func (h *Client) UninstallRelease(namespace string, releaseName string) (*releaseapi.UninstallReleaseResponse, error) {
	configurationWithNamespace, err := h.spawnConfigurationWithNamespace(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "create helm configuration")
	}
	uninstallAction := action.NewUninstall(configurationWithNamespace)
	uninstallAction.WaitStrategy = kube.HookOnlyStrategy
	response, err := uninstallAction.Run(releaseName)
	if err != nil {
		return nil, errors.Wrapf(err, "uninstall release %s, in namespace %s", releaseName, namespace)
	}
	return response, nil
}

func releaseV1(result releaseapi.Releaser) (*release.Release, error) {
	rel, ok := result.(*release.Release)
	if !ok || rel == nil {
		return nil, errors.Errorf("unexpected Helm release type %T", result)
	}
	return rel, nil
}
