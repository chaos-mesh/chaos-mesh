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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	genericnamespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
)

// SelectImpl selects the explicit route target from an EnvoyGatewayChaos.
type SelectImpl struct {
	c client.Client
	generic.Option
	logger logr.Logger
}

// Select returns the configured route target.
func (impl *SelectImpl) Select(ctx context.Context, target *v1alpha1.EnvoyGatewayTarget) ([]*v1alpha1.EnvoyGatewayTarget, error) {
	if target == nil {
		return []*v1alpha1.EnvoyGatewayTarget{}, nil
	}

	if !impl.ClusterScoped && target.Namespace != impl.TargetNamespace {
		return nil, errors.Errorf("could NOT select an Envoy Gateway route from out of scoped namespace: %s", target.Namespace)
	}

	if impl.EnableFilterNamespace && !genericnamespace.CheckNamespace(ctx, impl.c, target.Namespace, impl.logger) {
		return []*v1alpha1.EnvoyGatewayTarget{}, nil
	}

	return []*v1alpha1.EnvoyGatewayTarget{target}, nil
}

// New creates an Envoy Gateway target selector.
func New(c client.Client, logger logr.Logger) *SelectImpl {
	return &SelectImpl{
		c: c,
		Option: generic.Option{
			ClusterScoped:         config.ControllerCfg.ClusterScoped,
			TargetNamespace:       config.ControllerCfg.TargetNamespace,
			EnableFilterNamespace: config.ControllerCfg.EnableFilterNamespace,
		},
		logger: logger.WithName("envoy-gateway-selector"),
	}
}
