// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package gcp

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type SelectImpl struct{}

func (impl *SelectImpl) Select(ctx context.Context, gcpSelector *v1alpha1.GcpSelector) ([]*v1alpha1.GcpSelector, error) {
	return []*v1alpha1.GcpSelector{gcpSelector}, nil
}

func New() *SelectImpl {
	return &SelectImpl{}
}
