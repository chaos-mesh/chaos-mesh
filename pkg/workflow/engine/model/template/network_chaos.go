// Copyright 2020 Chaos Mesh Authors.
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

package template

import (
	"time"

	chaosmeshv1alph1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/errors"
)

type NetworkChaosTemplate interface {
	Template
	FetchChaosNamePrefix() string
	FetchNetworkChaosSpec() chaosmeshv1alph1.NetworkChaosSpec
	GetDuration() (time.Duration, error)
}

func ParseNetworkChaosTemplate(raw interface{}) (NetworkChaosTemplate, error) {
	op := "template.NetworkChaosTemplate"
	if target, ok := raw.(NetworkChaosTemplate); ok {
		return target, nil
	}
	return nil, errors.NewParseSerialTemplateFailedError(op, raw)
}
