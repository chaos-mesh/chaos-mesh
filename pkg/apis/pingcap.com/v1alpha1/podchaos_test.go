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
	"testing"
	"time"

	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPodChaosExperimentStatus(count int) PodChaosExperimentStatus {
	pe := PodChaosExperimentStatus{
		Phase: ExperimentPhaseFinished,
		Time:  metav1.Now(),
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

const (
	timeUnit = 1000000000 * 60
)

func newPodChaosStatus(count int, baseTime metav1.Time) PodChaosStatus {
	ps := PodChaosStatus{
		Phase: ChaosPhaseNormal,
	}

	for i := 0; i < count; i++ {
		ps.Experiments = append(ps.Experiments, PodChaosExperimentStatus{
			Phase: ExperimentPhaseFinished,
			Time:  metav1.Time{Time: baseTime.Add(time.Duration(i * timeUnit))},
		})
	}

	return ps
}

func TestPodChaosStatusSetExperimentRecord(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name                string
		ps                  PodChaosStatus
		addRecord           PodChaosExperimentStatus
		expectedRecordIndex int
	}

	baseTime := metav1.Now()

	tcs := []TestCase{
		{
			name: "add new record",
			ps:   newPodChaosStatus(3, baseTime),
			addRecord: PodChaosExperimentStatus{
				Phase: ExperimentPhaseFinished,
				Time:  metav1.Time{Time: baseTime.Add(4 * timeUnit)},
			},
			expectedRecordIndex: 3,
		},
		{
			name: "update record",
			ps:   newPodChaosStatus(2, baseTime),
			addRecord: PodChaosExperimentStatus{
				Phase: ExperimentPhaseFinished,
				Time:  metav1.Time{Time: baseTime.Add(1 * timeUnit)},
			},
			expectedRecordIndex: 1,
		},
	}

	for _, tc := range tcs {
		tc.ps.SetExperimentRecord(tc.addRecord)
		g.Expect(len(tc.ps.Experiments)).Should(BeNumerically(">=", tc.expectedRecordIndex+1), tc.name)
		g.Expect(tc.ps.Experiments[tc.expectedRecordIndex]).To(Equal(tc.addRecord), tc.name)
	}
}

func TestPodChaosCleanExpiredStatusRecords(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name              string
		ps                PodChaosStatus
		retentionTime     time.Duration
		expectedRecordLen int
	}

	baseTime := metav1.Time{Time: time.Now().Add(time.Duration(-10 * timeUnit))}

	tcs := []TestCase{
		{
			name:              "clean 0 records",
			ps:                newPodChaosStatus(5, baseTime),
			retentionTime:     20 * time.Minute,
			expectedRecordLen: 5,
		},
		{
			name:              "clean all records",
			ps:                newPodChaosStatus(5, baseTime),
			retentionTime:     1 * time.Minute,
			expectedRecordLen: 0,
		},
		{
			name:              "clean 1 records",
			ps:                newPodChaosStatus(5, baseTime),
			retentionTime:     (9*60 + 30) * time.Second,
			expectedRecordLen: 4,
		},
	}

	for _, tc := range tcs {
		tc.ps.CleanExpiredStatusRecords(tc.retentionTime)
		g.Expect(len(tc.ps.Experiments)).To(Equal(tc.expectedRecordLen), tc.name)
	}
}
