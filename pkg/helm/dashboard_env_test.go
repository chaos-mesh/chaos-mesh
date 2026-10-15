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
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
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
`,
			env: []corev1.EnvVar{
				{Name: "GCP_CLIENT_ID", Value: "gcp-client"},
				{Name: "GCP_CLIENT_SECRET", Value: "gcp-secret"},
			},
		},
		{
			name: "mixed inline and secret key",
			values: `dashboard:
  gcpSecurityMode:
    enabled: true
    clientId: dashboard-client
  envFollowKubernetesPattern:
    - name: GCP_CLIENT_SECRET
      valueFrom:
        secretKeyRef:
          name: dashboard-auth
          key: gcp-secret
`,
			env: []corev1.EnvVar{
				{Name: "GCP_CLIENT_SECRET", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "dashboard-auth"}, Key: "gcp-secret"}}},
				{Name: "GCP_CLIENT_ID", Value: "dashboard-client"},
			},
		},
		{
			name: "explicit scalar names are uppercased",
			values: `dashboard:
  env:
    gcp_client_id: custom-gcp
    gcp_client_secret: custom-gcp-secret
`,
			env: []corev1.EnvVar{
				{Name: "GCP_CLIENT_ID", Value: "custom-gcp"},
				{Name: "GCP_CLIENT_SECRET", Value: "custom-gcp-secret"},
			},
		},
		{
			name: "existing secrets",
			values: `dashboard:
  gcpSecurityMode:
    existingSecret: gcp-auth
`,
			envFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "gcp-auth"}}},
			},
		},
		{
			name: "envFrom with existing secrets",
			values: `dashboard:
  gcpSecurityMode:
    existingSecret: gcp-auth
  envFrom:
    - configMapRef:
        name: dashboard-config
    - prefix: EXTRA_
      secretRef:
        name: dashboard-extra
        optional: true
`,
			envFrom: []corev1.EnvFromSource{
				{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "gcp-auth"}}},
				{ConfigMapRef: &corev1.ConfigMapEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "dashboard-config"}}},
				{Prefix: "EXTRA_", SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "dashboard-extra"}, Optional: ptr.To(true)}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart, err := loader.Load("../../helm/chaos-mesh")
			require.NoError(t, err)
			values := map[string]interface{}{}
			require.NoError(t, yaml.Unmarshal([]byte(tt.values), &values))
			renderValues, err := chartutil.ToRenderValues(chart, values, chartutil.ReleaseOptions{Name: "test", Namespace: "chaos-mesh"}, nil)
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
				case "GCP_CLIENT_ID", "GCP_CLIENT_SECRET":
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
