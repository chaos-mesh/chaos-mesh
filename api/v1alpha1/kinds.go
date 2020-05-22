// Copyright 2020 PingCAP, Inc.
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

import "k8s.io/apimachinery/pkg/runtime"

const (
	// KindPodChaos is the kind for pod chaos
	KindPodChaos = "PodChaos"
	// KindNetworkChaos is the kind for network chaos
	KindNetworkChaos = "NetworkChaos"
	// KindIOChaos is the kind for io chaos
	KindIOChaos = "IoChaos"
	// KindKernelChaos is the kind for kernel chaos
	KindKernelChaos = "KernelChaos"
	// KindStressChaos is the kind for stress chaos
	KindStressChaos = "StressChaos"
	// KindTimeChaos is the kind for time chaos
	KindTimeChaos = "TimeChaos"
)

var (
	// Kinds is a collection to keep chaos kind name and its type
	Kinds = map[string]*ChaosKind{
		KindPodChaos:     {Chaos: &PodChaos{}, ChaosList: &PodChaosList{}},
		KindNetworkChaos: {Chaos: &NetworkChaos{}, ChaosList: &NetworkChaosList{}},
		KindIOChaos:      {Chaos: &IoChaos{}, ChaosList: &IoChaosList{}},
		KindKernelChaos:  {Chaos: &KernelChaos{}, ChaosList: &KernelChaosList{}},
		KindTimeChaos:    {Chaos: &TimeChaos{}, ChaosList: &TimeChaosList{}},
		KindStressChaos:  {Chaos: &StressChaos{}, ChaosList: &StressChaosList{}},
	}
)

// +kubebuilder:object:generate=false

// ChaosKind includes one kind of chaos and its list type
type ChaosKind struct {
	Chaos runtime.Object
	ChaosList
}
