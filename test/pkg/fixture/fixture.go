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

package fixture

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
)

func NewDefaultPodChaos() *v1alpha1.PodChaos {

	return &v1alpha1.PodChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-chaos-example",
			Namespace: "chaos-testing",
		},
		Spec: v1alpha1.PodChaosSpec{
			Selector: v1alpha1.SelectorSpec{
				Namespaces: []string{
					"chaos-testing",
				},
			},
			Action: v1alpha1.PodFailureAction,
			Mode:   v1alpha1.OnePodMode,
		},
	}
}
