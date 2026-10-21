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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"
)

func TestDNSServerConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		values   map[string]interface{}
		grpcPort int32
	}{
		{name: "defaults", grpcPort: 9288},
		{
			name: "custom grpc port",
			values: map[string]interface{}{
				"dnsServer": map[string]interface{}{"grpcPort": 19288},
			},
			grpcPort: 19288,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart, err := loader.Load("../../helm/chaos-mesh")
			require.NoError(t, err)
			renderValues, err := chartutil.ToRenderValues(chart, tt.values, chartutil.ReleaseOptions{Name: "test", Namespace: "chaos-mesh"}, nil)
			require.NoError(t, err)
			rendered, err := engine.Render(chart, renderValues)
			require.NoError(t, err)

			var config corev1.ConfigMap
			require.NoError(t, yaml.Unmarshal([]byte(rendered["chaos-mesh/templates/dns-configmap.yaml"]), &config))
			corefile := strings.Join(strings.Fields(config.Data["Corefile"]), " ")
			require.Contains(t, corefile, ".:5353 {")
			require.Contains(t, corefile, fmt.Sprintf("dns_chaos { grpcport %d }", tt.grpcPort))
			require.Contains(t, corefile, "kubernetes cluster.local in-addr.arpa ip6.arpa { pods insecure fallthrough in-addr.arpa ip6.arpa ttl 30 }")
			require.Contains(t, corefile, "cache 30")
			require.Contains(t, corefile, "forward . /etc/resolv.conf { max_concurrent 1000 }")
			require.NotContains(t, corefile, "k8s_dns_chaos")

			var deployment appsv1.Deployment
			require.NoError(t, yaml.Unmarshal([]byte(rendered["chaos-mesh/templates/dns-deployment.yaml"]), &deployment))
			require.Len(t, deployment.Spec.Template.Spec.Containers, 1)
			container := deployment.Spec.Template.Spec.Containers[0]
			require.Equal(t, "ghcr.io/chaos-mesh/chaos-coredns:v0.3.0", container.Image)
			require.ElementsMatch(t, []corev1.ContainerPort{
				{Name: "dns", ContainerPort: 5353, Protocol: corev1.ProtocolUDP},
				{Name: "dns-tcp", ContainerPort: 5353, Protocol: corev1.ProtocolTCP},
				{Name: "metrics", ContainerPort: 9153, Protocol: corev1.ProtocolTCP},
				{Name: "grpc", ContainerPort: tt.grpcPort, Protocol: corev1.ProtocolTCP},
			}, container.Ports)

			var service corev1.Service
			require.NoError(t, yaml.Unmarshal([]byte(rendered["chaos-mesh/templates/dns-service.yaml"]), &service))
			require.ElementsMatch(t, []corev1.ServicePort{
				{Name: "dns", Port: 53, TargetPort: intstr.FromInt32(5353), Protocol: corev1.ProtocolUDP},
				{Name: "dns-tcp", Port: 53, TargetPort: intstr.FromInt32(5353), Protocol: corev1.ProtocolTCP},
				{Name: "metrics", Port: 9153, Protocol: corev1.ProtocolTCP},
				{Name: "grpc", Port: tt.grpcPort, Protocol: corev1.ProtocolTCP},
			}, service.Spec.Ports)

			var controller appsv1.Deployment
			require.NoError(t, yaml.Unmarshal([]byte(rendered["chaos-mesh/templates/controller-manager-deployment.yaml"]), &controller))
			require.NotEmpty(t, controller.Spec.Template.Spec.Containers)
			require.Contains(t, controller.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{
				Name: "CHAOS_DNS_SERVICE_PORT", Value: fmt.Sprint(tt.grpcPort),
			})

			var clusterRole rbacv1.ClusterRole
			for _, document := range strings.Split(rendered["chaos-mesh/templates/dns-rbac.yaml"], "---") {
				var role rbacv1.ClusterRole
				require.NoError(t, yaml.Unmarshal([]byte(document), &role))
				if role.Kind == "ClusterRole" && role.Name == "test-chaos-dns-server-cluster-level" {
					clusterRole = role
				}
			}
			require.NotEmpty(t, clusterRole.Name)
			require.Contains(t, clusterRole.Rules, rbacv1.PolicyRule{
				APIGroups: []string{"discovery.k8s.io"},
				Resources: []string{"endpointslices"},
				Verbs:     []string{"get", "list", "watch"},
			})
		})
	}
}
