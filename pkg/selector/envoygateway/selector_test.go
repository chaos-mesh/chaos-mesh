// Copyright 2026 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package envoygateway

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
)

func TestSelectHonorsControllerNamespaceScope(t *testing.T) {
	selector := &SelectImpl{
		Option: generic.Option{ClusterScoped: false, TargetNamespace: "chaos-testing"},
		logger: logr.Discard(),
	}

	_, err := selector.Select(context.Background(), &v1alpha1.EnvoyGatewayTarget{Namespace: "production"})
	require.ErrorContains(t, err, "out of scoped namespace")
}

func TestSelectHonorsNamespaceFilter(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "enabled",
				Annotations: map[string]string{generic.InjectAnnotationKey: "enabled"},
			},
		},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "disabled"}},
	).Build()
	selector := &SelectImpl{
		c:      client,
		Option: generic.Option{ClusterScoped: true, EnableFilterNamespace: true},
		logger: logr.Discard(),
	}

	target := &v1alpha1.EnvoyGatewayTarget{Namespace: "enabled"}
	selected, err := selector.Select(context.Background(), target)
	require.NoError(t, err)
	require.Equal(t, []*v1alpha1.EnvoyGatewayTarget{target}, selected)

	selected, err = selector.Select(context.Background(), &v1alpha1.EnvoyGatewayTarget{Namespace: "disabled"})
	require.NoError(t, err)
	require.Empty(t, selected)
}
