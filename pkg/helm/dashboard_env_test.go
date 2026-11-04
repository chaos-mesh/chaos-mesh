// Copyright 2026 Chaos Mesh Authors.
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
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/chart/common"
	chartutil "helm.sh/helm/v4/pkg/chart/common/util"
	"helm.sh/helm/v4/pkg/chart/v2/loader"
	"helm.sh/helm/v4/pkg/engine"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

func TestDashboardAuthEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		values  string
		env     []corev1.EnvVar
		envFrom []corev1.EnvFromSource
	}{
		{
			name: "inline credentials",
			values: `dashboard:
  gcpSecurityMode:
    clientId: gcp-client
    clientSecret: gcp-secret
  oidcSecurityMode:
    clientId: oidc-client
    clientSecret: oidc-secret
    issuerUrl: https://issuer.example.test
`,
			env: []corev1.EnvVar{
				{Name: "GCP_CLIENT_ID", Value: "gcp-client"},
				{Name: "GCP_CLIENT_SECRET", Value: "gcp-secret"},
				{Name: "OIDC_CLIENT_ID", Value: "oidc-client"},
				{Name: "OIDC_CLIENT_SECRET", Value: "oidc-secret"},
				{Name: "OIDC_ISSUER_URL", Value: "https://issuer.example.test"},
			},
		},
		{
			name: "mixed inline and secret key",
			values: `dashboard:
  oidcSecurityMode:
    enabled: true
    clientId: dashboard-client
    issuerUrl: https://issuer.example.test
  envFollowKubernetesPattern:
    - name: GCP_CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: dashboard-auth
          key: gcp-secret
    - name: OIDC_CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: dashboard-auth
          key: client-secret
`,
			env: []corev1.EnvVar{
				{Name: "GCP_CLIENT_SECRET", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{Name: "dashboard-auth", Key: "gcp-secret"}}},
				{Name: "OIDC_CLIENT_SECRET", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{Name: "dashboard-auth", Key: "client-secret"}}},
				{Name: "GCP_CLIENT_ID"},
				{Name: "OIDC_CLIENT_ID", Value: "dashboard-client"},
				{Name: "OIDC_ISSUER_URL", Value: "https://issuer.example.test"},
			},
		},
		{
			name: "explicit scalar names are uppercased",
			values: `dashboard:
  env:
    gcp_client_id: custom-gcp
    gcp_client_secret: custom-gcp-secret
    oidc_client_id: custom-oidc
    oidc_client_secret: custom-oidc-secret
    oidc_issuer_url: https://custom.example.test
`,
			env: []corev1.EnvVar{
				{Name: "GCP_CLIENT_ID", Value: "custom-gcp"},
				{Name: "GCP_CLIENT_SECRET", Value: "custom-gcp-secret"},
				{Name: "OIDC_CLIENT_ID", Value: "custom-oidc"},
				{Name: "OIDC_CLIENT_SECRET", Value: "custom-oidc-secret"},
				{Name: "OIDC_ISSUER_URL", Value: "https://custom.example.test"},
			},
		},
		{
			name: "existing secrets",
			values: `dashboard:
  gcpSecurityMode:
    existingSecret: gcp-auth
  oidcSecurityMode:
    existingSecret: oidc-auth
`,
			envFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{Name: "gcp-auth"}},
				{SecretRef: &corev1.SecretEnvSource{Name: "oidc-auth"}},
			},
		},
		{
			name: "envFrom with existing secrets",
			values: `dashboard:
  gcpSecurityMode:
    existingSecret: gcp-auth
  oidcSecurityMode:
    existingSecret: oidc-auth
  envFrom:
    - configMapRef:
        name: dashboard-config
    - prefix: EXTRA_
      secretRef:
        name: dashboard-extra
        optional: true
`,
			envFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{Name: "gcp-auth"}},
				{SecretRef: &corev1.SecretEnvSource{Name: "oidc-auth"}},
				{ConfigMapRef: &corev1.ConfigMapEnvSource{Name: "dashboard-config"}},
				{Prefix: "EXTRA_", SecretRef: &corev1.SecretEnvSource{Name: "dashboard-extra", Optional: new(true)}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart, err := loader.Load("../../helm/chaos-mesh")
			require.NoError(t, err)
			values := map[string]any{}
			require.NoError(t, yaml.Unmarshal([]byte(tt.values), &values))
			renderValues, err := chartutil.ToRenderValues(chart, values, common.ReleaseOptions{Name: "test", Namespace: "chaos-mesh"}, nil)
			require.NoError(t, err)
			rendered, err := engine.Render(chart, renderValues)
			require.NoError(t, err)
			var deployment appsv1.Deployment
			require.NoError(t, yaml.Unmarshal([]byte(rendered["chaos-mesh/templates/chaos-dashboard-deployment.yaml"]), &deployment))
			require.NotEmpty(t, deployment.Spec.Template.Spec.Containers)
			container := deployment.Spec.Template.Spec.Containers[0]
			var authEnv []corev1.EnvVar
			seen := map[string]bool{}
			for _, env := range container.Env {
				switch env.Name {
				case "GCP_CLIENT_ID", "GCP_CLIENT_SECRET", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "OIDC_ISSUER_URL":
					require.False(t, seen[env.Name], "duplicate environment variable %s", env.Name)
					seen[env.Name] = true
					authEnv = append(authEnv, env)
				}
			}
			require.Equal(t, tt.env, authEnv)
			require.Equal(t, tt.envFrom, container.EnvFrom)
		})
	}
}
