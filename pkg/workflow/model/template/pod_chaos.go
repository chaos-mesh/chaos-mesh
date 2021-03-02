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
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/errors"
)

type PodChaosTemplate interface {
	Template
	ChaosNamePrefix() string
	PodChaosSpec() chaosmeshv1alph1.PodChaosSpec
	Duration() (time.Duration, error)
}

func ParsePodChaosTemplate(raw interface{}) (PodChaosTemplate, error) {
	op := "template.ParsePodChaosTemplate"
	if target, ok := raw.(PodChaosTemplate); ok {
		return target, nil
	}
	return nil, errors.NewParseSerialTemplateFailedError(op, raw)
}
