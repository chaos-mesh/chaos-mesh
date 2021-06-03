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

type TemplateType string

const (
	Task         TemplateType = "Task"
	Serial       TemplateType = "Serial"
	Parallel     TemplateType = "Parallel"
	Suspend      TemplateType = "Suspend"
	IOChaos      TemplateType = "IOChaos"
	NetworkChaos TemplateType = "NetworkChaos"
	StressChaos  TemplateType = "StressChaos"
	PodChaos     TemplateType = "PodChaos"
	TimeChaos    TemplateType = "TimeChaos"
	KernelChaos  TemplateType = "KernelChaos"
	DnsChaos     TemplateType = "DnsChaos"
	HttpChaos    TemplateType = "HttpChaos"
	JvmChaos     TemplateType = "JvmChaos"
)

type Template interface {
	Name() string
	TemplateType() TemplateType
}

// func IsCompositeType. CompositeType means this Template could have children Templates.
func (it TemplateType) IsCompositeType() bool {
	return it == Serial || it == Parallel || it == Task
}
