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

package podchaos

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pingcap/chaos-operator/pkg/apis/pingcap.com/v1alpha1"
	"github.com/pingcap/chaos-operator/pkg/client/clientset/versioned/fake"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenAnnotationKeyForImage(t *testing.T) {
	g := NewGomegaWithT(t)

	pc := newPodChaos("test")
	g.Expect(GenAnnotationKeyForImage(pc, "t")).
		To(Equal(fmt.Sprintf("%s-%s-%s-t-image", AnnotationPrefix, pc.Name, pc.Spec.Action)))
}

func TestCleanExpiredExperimentRecords(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		retentionTime  string
		expectedResult resultF
	}

	tcs := []TestCase{
		{
			name:           "invalid retention time",
			retentionTime:  "1",
			expectedResult: HaveOccurred,
		},
		{
			name:           "all is ok",
			retentionTime:  "1h",
			expectedResult: Succeed,
		},
	}

	for _, tc := range tcs {
		pc := newPodChaos("test")
		pc.Spec.StatusRetentionTime = tc.retentionTime

		cli := fake.NewSimpleClientset(pc)

		g.Expect(cleanExpiredExperimentRecords(cli, pc)).Should(tc.expectedResult(), tc.name)
	}

	cli := fake.NewSimpleClientset()
	g.Expect(cleanExpiredExperimentRecords(cli, newPodChaos("test"))).
		Should(HaveOccurred(), "podChaos not found")
}

func TestSetExperimentRecord(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		podChaos       *v1alpha1.PodChaos
		record         *v1alpha1.PodChaosExperimentStatus
		expectedStatus v1alpha1.ChaosPhase
	}

	tcs := []TestCase{
		{
			name:     "set failed record",
			podChaos: newPodChaos("t1"),
			record: &v1alpha1.PodChaosExperimentStatus{
				Phase: v1alpha1.ExperimentPhaseFailed,
				Time:  metav1.Now(),
			},
			expectedStatus: v1alpha1.ChaosPahseAbnormal,
		},
		{
			name:     "set running record",
			podChaos: newPodChaos("t1"),
			record: &v1alpha1.PodChaosExperimentStatus{
				Phase: v1alpha1.ExperimentPhaseRunning,
				Time:  metav1.Now(),
			},
			expectedStatus: v1alpha1.ChaosPhaseNormal,
		},
		{
			name:     "set finished record",
			podChaos: newPodChaos("t1"),
			record: &v1alpha1.PodChaosExperimentStatus{
				Phase: v1alpha1.ExperimentPhaseRunning,
				Time:  metav1.Now(),
			},
			expectedStatus: v1alpha1.ChaosPhaseNormal,
		},
	}

	for _, tc := range tcs {
		cli := fake.NewSimpleClientset(tc.podChaos)
		g.Expect(setExperimentRecord(cli, tc.podChaos, tc.record)).ShouldNot(HaveOccurred(), tc.name)

		tpc, err := cli.PingcapV1alpha1().PodChaoses(tc.podChaos.Namespace).
			Get(tc.podChaos.Name, metav1.GetOptions{})
		g.Expect(err).ShouldNot(HaveOccurred(), tc.name)
		g.Expect(tpc.Status.Phase).To(Equal(tc.expectedStatus), tc.name)
		g.Expect(len(tpc.Status.Experiments)).To(Equal(1), tc.name)
	}

	cli := fake.NewSimpleClientset()
	g.Expect(setExperimentRecord(cli, newPodChaos("test"), &v1alpha1.PodChaosExperimentStatus{})).
		Should(HaveOccurred(), "podChaos not found")
}

func TestSetRecordPods(t *testing.T) {
	g := NewGomegaWithT(t)

	record := &v1alpha1.PodChaosExperimentStatus{
		Phase: v1alpha1.ExperimentPhaseRunning,
		Time:  metav1.Now(),
	}

	_, pods := generateNPods("t", 5, v1.PodRunning)

	setRecordPods(record, v1alpha1.PodKillAction, podKillActionMsg, pods...)

	g.Expect(len(record.Pods)).To(Equal(5))
	for _, ps := range record.Pods {
		g.Expect(string(ps.Action)).To(Equal(string(v1alpha1.PodKillAction)), ps.Name)
		g.Expect(ps.Message).To(Equal(podKillActionMsg), ps.Name)
	}
}
