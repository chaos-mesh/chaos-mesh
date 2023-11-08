// Copyright 2023 Chaos Mesh Authors.
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

package cloudstackvm

import (
	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/action"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/cloudstackvm/vmrestart"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/cloudstackvm/vmstop"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
)

type Impl struct {
	fx.In

	VMStop    *vmstop.Impl    `action:"vm-stop"`
	VMRestart *vmrestart.Impl `action:"vm-restart"`
}

func NewImpl(impl Impl) *types.ChaosImplPair {
	delegate := action.NewMultiplexer(&impl)
	return &types.ChaosImplPair{
		Name:   "cloudstackvmchaos",
		Object: &v1alpha1.CloudStackVMChaos{},
		Impl:   &delegate,
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	vmstop.NewImpl,
	vmrestart.NewImpl,
)
