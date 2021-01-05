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

package v1alpha1

import (
	"time"

	chaosmeshv1alph1 "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"
)

type Template struct {
	Name         string                            `json:"name"`
	TemplateType template.TemplateType             `json:"template_type"`
	Duration     string                            `json:"duration"`
	Deadline     string                            `json:"deadline"`
	NetworkChaos chaosmeshv1alph1.NetworkChaosSpec `json:"network_chaos"`
}

func (it *Template) GetName() string {
	return it.Name
}

func (it *Template) GetTemplateType() template.TemplateType {
	return it.TemplateType
}

func (it *Template) GetDuration() (time.Duration, error) {
	return time.ParseDuration(it.Duration)
}

func (it *Template) GetDeadline() (time.Duration, error) {
	return time.ParseDuration(it.Deadline)
}

func (it *Template) FetchChaosNamePrefix() string {
	return it.Name
}

func (it *Template) FetchNetworkChaosSpec() chaosmeshv1alph1.NetworkChaosSpec {
	return it.NetworkChaos
}
