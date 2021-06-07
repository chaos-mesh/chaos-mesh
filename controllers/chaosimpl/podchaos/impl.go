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

package podchaos

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/action"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/containerkill"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/podfailure"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/podchaos/podkill"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

type Impl struct {
	fx.In

	PodKill       *podkill.Impl       `action:"pod-kill"`
	PodFailure    *podfailure.Impl    `action:"pod-failure"`
	ContainerKill *containerkill.Impl `action:"container-kill"`
}

func NewImpl(impl Impl) *common.ChaosImplPair {
	delegate := action.New(&impl)
	return &common.ChaosImplPair{
		Name:   "podchaos",
		Object: &v1alpha1.PodChaos{},
		Impl:   &delegate,
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	podkill.NewImpl,
	podfailure.NewImpl,
	containerkill.NewImpl,
)
