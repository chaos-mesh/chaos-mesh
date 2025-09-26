// Copyright 2021 Chaos Mesh Authors.
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

package ycchaos

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/action"
	impltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/ycchaos/computerestart"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/ycchaos/computestop"
)

type Impl struct {
	fx.In

	ComputeStop    *computestop.Impl    `action:"compute-stop"`
	ComputeRestart *computerestart.Impl `action:"compute-restart"`
}

func NewImpl(impl Impl) *impltypes.ChaosImplPair {
	delegate := action.NewMultiplexer(&impl)
	return &impltypes.ChaosImplPair{
		Name:   "ycchaos",
		Object: &v1alpha1.YCChaos{},
		Impl:   &delegate,
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	computestop.NewImpl,
	computerestart.NewImpl,
)
