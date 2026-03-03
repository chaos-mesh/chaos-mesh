// Copyright 2022 Chaos Mesh Authors.
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

package common

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestStatusRecordEventsChangePredicateEventsChange(t *testing.T) {
	g := NewWithT(t)
	// should skip event, if we Only the object.status.experiment.records[].events
	// notice: ResourceVersion and Generation will be changed by k8s itself
	newObj := &v1alpha1.HTTPChaos{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "1",
			Generation:      1,
		},
		Spec: v1alpha1.HTTPChaosSpec{},
		Status: v1alpha1.HTTPChaosStatus{
			ChaosStatus: v1alpha1.ChaosStatus{
				Conditions: nil,
				Experiment: v1alpha1.ExperimentStatus{
					DesiredPhase: "",
					Records: []*v1alpha1.Record{
						{
							Id:             "",
							SelectorKey:    "",
							Phase:          "",
							InjectedCount:  0,
							RecoveredCount: 0,
							Events: []v1alpha1.RecordEvent{
								{
									Type:      v1alpha1.TypeFailed,
									Operation: v1alpha1.Apply,
									Message:   "apply failed",
									Timestamp: nil,
								},
							},
						},
					},
				},
			},
			Instances: nil,
		},
	}
	oldObj := &v1alpha1.HTTPChaos{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "0",
			Generation:      0,
		},
		Spec: v1alpha1.HTTPChaosSpec{},
		Status: v1alpha1.HTTPChaosStatus{
			ChaosStatus: v1alpha1.ChaosStatus{
				Conditions: nil,
				Experiment: v1alpha1.ExperimentStatus{
					DesiredPhase: "",
					Records: []*v1alpha1.Record{
						{
							Id:             "",
							SelectorKey:    "",
							Phase:          "",
							InjectedCount:  0,
							RecoveredCount: 0,
							Events:         nil,
						},
					},
				},
			},
			Instances: nil,
		},
	}

	predicate := StatusRecordEventsChangePredicate{}
	updateEvent := event.UpdateEvent{ObjectOld: oldObj, ObjectNew: newObj}
	pick := predicate.Update(updateEvent)
	g.Expect(pick).Should(Equal(false))
}

func TestStatusRecordEventsChangePredicatePhaseChange(t *testing.T) {
	g := NewWithT(t)

	newObj := &v1alpha1.HTTPChaos{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "1",
			Generation:      1,
		},
		Spec: v1alpha1.HTTPChaosSpec{},
		Status: v1alpha1.HTTPChaosStatus{
			ChaosStatus: v1alpha1.ChaosStatus{
				Conditions: nil,
				Experiment: v1alpha1.ExperimentStatus{
					DesiredPhase: v1alpha1.StoppedPhase,
					Records:      nil,
				},
			},
			Instances: nil,
		},
	}
	oldObj := &v1alpha1.HTTPChaos{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			ResourceVersion: "0",
			Generation:      0,
		},
		Spec: v1alpha1.HTTPChaosSpec{},
		Status: v1alpha1.HTTPChaosStatus{
			ChaosStatus: v1alpha1.ChaosStatus{
				Conditions: nil,
				Experiment: v1alpha1.ExperimentStatus{
					DesiredPhase: v1alpha1.RunningPhase,
					Records:      nil,
				},
			},
			Instances: nil,
		},
	}

	predicate := StatusRecordEventsChangePredicate{}
	updateEvent := event.UpdateEvent{ObjectOld: oldObj, ObjectNew: newObj}
	pick := predicate.Update(updateEvent)
	g.Expect(pick).Should(Equal(true))
}

func TestStatusRecordEventsChangePredicateChildCRDs(t *testing.T) {
	g := NewWithT(t)

	// should allow the event about "Update" on certain Chaos Resource, v1alpha1.StatefulObject
	newObj := &v1alpha1.PodIOChaos{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.PodIOChaosSpec{},
		Status:     v1alpha1.PodIOChaosStatus{},
	}
	oldObj := &v1alpha1.PodIOChaos{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       v1alpha1.PodIOChaosSpec{},
		Status:     v1alpha1.PodIOChaosStatus{},
	}

	predicate := StatusRecordEventsChangePredicate{}
	updateEvent := event.UpdateEvent{ObjectOld: oldObj, ObjectNew: newObj}
	pick := predicate.Update(updateEvent)
	g.Expect(pick).Should(Equal(false))
}
