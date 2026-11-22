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

package remotecluster

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
	"helm.sh/helm/v4/pkg/storage/driver"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	controllerconfig "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

func TestRemoteHelmReleaseLifecycle(t *testing.T) {
	g := NewWithT(t)
	ctx := t.Context()
	binaryPath := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryPath == "" {
		binaryPath = "/usr/local/kubebuilder/bin"
	}
	for _, asset := range []string{"kube-apiserver", "etcd"} {
		if _, err := os.Stat(filepath.Join(binaryPath, asset)); os.IsNotExist(err) {
			t.Skipf("envtest asset %s is unavailable in %s", asset, binaryPath)
		}
	}

	env := &envtest.Environment{UseExistingCluster: new(false), BinaryAssetsDirectory: binaryPath}
	restConfig, err := env.Start()
	g.Expect(err).NotTo(HaveOccurred())
	t.Cleanup(func() {
		if err := env.Stop(); err != nil {
			t.Error(err)
		}
	})
	const namespace = "remote-helm"
	kube, err := kubernetes.NewForConfig(restConfig)
	g.Expect(err).NotTo(HaveOccurred())
	kubeconfig := clientcmd.NewDefaultClientConfig(clientcmdapi.Config{
		Clusters:       map[string]*clientcmdapi.Cluster{"test": {Server: restConfig.Host, CertificateAuthorityData: restConfig.CAData}},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{"test": {ClientCertificateData: restConfig.CertData, ClientKeyData: restConfig.KeyData}},
		Contexts:       map[string]*clientcmdapi.Context{"test": {Cluster: "test", AuthInfo: "test", Namespace: namespace}},
		CurrentContext: "test",
	}, &clientcmd.ConfigOverrides{})

	directory := t.TempDir()
	previous := controllerconfig.ControllerCfg
	controllerconfig.ControllerCfg = &config.ChaosControllerConfig{LocalHelmChartPath: directory}
	t.Cleanup(func() { controllerconfig.ControllerCfg = previous })
	saveChart := func(version, template string) {
		t.Helper()
		_, err := chartutil.Save(&chart.Chart{
			Metadata:  &chart.Metadata{Name: chaosMeshReleaseName, Version: version, APIVersion: chart.APIVersionV2, AppVersion: version},
			Templates: []*common.File{{Name: "templates/configmap.yaml", Data: []byte(template)}},
		}, directory)
		g.Expect(err).NotTo(HaveOccurred())
	}
	const template = `apiVersion: v1
kind: ConfigMap
metadata:
  name: remote-release-config
data:
  value: {{ .Values.configValue | quote }}
`
	saveChart("1.0.0", template)
	// Parse failures leave the deployed release usable and safe to retry.
	saveChart("2.0.0", "{{ .Values.configValue | missingFunction }}")
	r := Reconciler{Log: logr.Discard()}
	remote := &v1alpha1.RemoteCluster{}
	remote.Spec.Namespace = namespace
	remote.Spec.Version = "1.0.0"
	remote.Spec.ConfigOverride = []byte(`{"configValue":"original"}`)
	cms := kube.CoreV1().ConfigMaps(namespace)
	assertValue := func(want string) {
		t.Helper()
		cm, err := cms.Get(ctx, "remote-release-config", metav1.GetOptions{})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(cm.Data).To(HaveKeyWithValue("value", want))
	}

	g.Expect(r.uninstallHelmRelease(ctx, remote, kubeconfig)).To(Succeed())
	g.Expect(r.ensureHelmRelease(ctx, remote, kubeconfig)).To(Equal("1.0.0"))
	assertValue("original")

	remote.Spec.ConfigOverride = []byte("{")
	g.Expect(r.ensureHelmRelease(ctx, remote, kubeconfig)).Error().To(HaveOccurred())
	assertValue("original")
	remote.Spec.ConfigOverride = []byte(`{"configValue":"helm-value-marker"}`)
	remote.Spec.Version = "missing"
	g.Expect(r.ensureHelmRelease(ctx, remote, kubeconfig)).Error().To(MatchError(os.ErrNotExist))
	remote.Spec.Version = "2.0.0"
	_, err = r.ensureHelmRelease(ctx, remote, kubeconfig)
	g.Expect(err).To(MatchError(ContainSubstring("upgrade release chaos-mesh")))
	g.Expect(err.Error()).NotTo(ContainSubstring("helm-value-marker"))
	assertValue("original")

	saveChart("2.0.0", template)
	remote.Spec.ConfigOverride = []byte(`{"configValue":"upgraded"}`)
	g.Expect(r.ensureHelmRelease(ctx, remote, kubeconfig)).To(Equal("2.0.0"))
	assertValue("upgraded")
	helmClient, err := r.getHelmClient(ctx, kubeconfig)
	g.Expect(err).NotTo(HaveOccurred())
	release, err := helmClient.GetRelease(namespace, chaosMeshReleaseName)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(release.Version).To(Equal(2))
	g.Expect(r.uninstallHelmRelease(ctx, remote, kubeconfig)).To(Succeed())
	_, err = cms.Get(ctx, "remote-release-config", metav1.GetOptions{})
	g.Expect(err).To(Satisfy(apierrors.IsNotFound))
	g.Expect(helmClient.GetRelease(namespace, chaosMeshReleaseName)).Error().To(MatchError(driver.ErrReleaseNotFound))
	g.Expect(r.uninstallHelmRelease(ctx, remote, kubeconfig)).To(Succeed())
}
