// Copyright 2019 PingCAP, Inc.
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
	"fmt"
	. "github.com/onsi/gomega"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPodChaosExperimentStatus(count int) PodChaosExperimentStatus {
	pe := PodChaosExperimentStatus{
		Phase:     ExperimentPhaseFinished,
		StartTime: metav1.Now(),
	}

	for i := 0; i < count; i++ {
		pe.Pods = append(pe.Pods, PodStatus{
			Namespace: metav1.NamespaceDefault,
			Name:      fmt.Sprintf("%d", i),
		})
	}

	return pe
}

func TestPodChaosExperimentStatusSetPods(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name             string
		pe               PodChaosExperimentStatus
		addPod           PodStatus
		expectedPodIndex int
	}

	tcs := []TestCase{
		{
			name: "add new pod",
			pe:   newPodChaosExperimentStatus(2),
			addPod: PodStatus{
				Namespace: "t1",
				Name:      "t",
			},
			expectedPodIndex: 2,
		},
		{
			name: "empty",
			pe:   newPodChaosExperimentStatus(0),
			addPod: PodStatus{
				Namespace: "t1",
				Name:      "t",
			},
			expectedPodIndex: 0,
		},
		{
			name: "update pod",
			pe:   newPodChaosExperimentStatus(2),
			addPod: PodStatus{
				Namespace: metav1.NamespaceDefault,
				Name:      "1",
			},
			expectedPodIndex: 1,
		},
	}

	for _, tc := range tcs {
		tc.pe.SetPods(tc.addPod)
		g.Expect(len(tc.pe.Pods)).Should(BeNumerically(">=", tc.expectedPodIndex+1), tc.name)
		g.Expect(tc.pe.Pods[tc.expectedPodIndex]).To(Equal(tc.addPod), tc.name)
	}
}
