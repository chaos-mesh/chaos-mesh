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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// SelectImpl selects the explicit route target from an EnvoyGatewayChaos.
type SelectImpl struct{}

// Select returns the configured route target.
func (impl *SelectImpl) Select(_ context.Context, target *v1alpha1.EnvoyGatewayTarget) ([]*v1alpha1.EnvoyGatewayTarget, error) {
	return []*v1alpha1.EnvoyGatewayTarget{target}, nil
}

// New creates an Envoy Gateway target selector.
func New() *SelectImpl {
	return &SelectImpl{}
}
