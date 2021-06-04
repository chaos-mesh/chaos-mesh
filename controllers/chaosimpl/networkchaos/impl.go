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

package networkchaos

import (
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/action"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/partition"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/trafficcontrol"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
)

type Impl struct {
	fx.In

	TrafficControl *trafficcontrol.Impl `action:"bandwidth,netem,delay,loss,duplicate,corrupt"`
	Partition      *partition.Impl      `action:"partition"`
}

func NewImpl(impl Impl) *common.ChaosImplPair {
	delegate := action.New(&impl)
	return &common.ChaosImplPair{
		Name:       "networkchaos",
		Object:     &v1alpha1.NetworkChaos{},
		Impl:       &delegate,
		ObjectList: &v1alpha1.NetworkChaosList{},
		Controlls:  []runtime.Object{&v1alpha1.PodNetworkChaos{}},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	trafficcontrol.NewImpl,
	partition.NewImpl,
	podnetworkchaosmanager.NewBuilder,
)
