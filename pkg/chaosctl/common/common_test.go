// Copyright 2020 Chaos Mesh Authors.
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

package common

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubectlscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

func TestGetChaosList(t *testing.T) {
	logger, _, _ := NewStderrLogger()
	SetupGlobalLogger(logger)

	g := NewWithT(t)

	chaos1 := v1alpha1.NetworkChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkChaos",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "fakechaos-1",
		},
	}

	chaos2 := v1alpha1.NetworkChaos{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NetworkChaos",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: metav1.NamespaceDefault,
			Name:      "fakechaos-2",
		},
	}

	g.Expect(v1alpha1.SchemeBuilder.AddToScheme(kubectlscheme.Scheme)).To(BeNil())

	client := fake.NewFakeClientWithScheme(kubectlscheme.Scheme, &chaos1, &chaos2)

	tests := []struct {
		name        string
		chaosType   string
		chaosName   string
		ns          string
		expectedNum int
		expectedErr bool
	}{
		{
			name:        "Only specify chaosType",
			chaosType:   "networkchaos",
			chaosName:   "",
			ns:          "default",
			expectedNum: 2,
			expectedErr: false,
		},
		{
			name:        "Specify chaos type, chaos name and namespace",
			chaosType:   "networkchaos",
			chaosName:   "fakechaos-1",
			ns:          "default",
			expectedNum: 1,
			expectedErr: false,
		},
		{
			name:        "Specify non-exist chaos name",
			chaosType:   "networkchaos",
			chaosName:   "fakechaos-oops",
			ns:          "default",
			expectedErr: true,
		},
		{
			name:        "Specify non-exist chaos types",
			chaosType:   "stresschaos",
			chaosName:   "fakechaos-1",
			ns:          "default",
			expectedErr: true,
		},
		{
			name:        "Specify non-exist namespace",
			chaosType:   "networkchaos",
			chaosName:   "fakechaos-1",
			ns:          "oops",
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chaos, _, err := GetChaosList(context.Background(), test.chaosType, test.chaosName, test.ns, client)
			if test.expectedErr {
				g.Expect(err).NotTo(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(len(chaos)).To(Equal(test.expectedNum))
			}
		})
	}
}

func TestGetPods(t *testing.T) {
	logger, _, _ := NewStderrLogger()
	SetupGlobalLogger(logger)

	g := NewWithT(t)

	_ = v1alpha1.ChaosStatus{
		Experiment: v1alpha1.ExperimentStatus{
			DesiredPhase: v1alpha1.RunningPhase,
		},
	}

	nodeObjects, _ := utils.GenerateNNodes("node", 2, nil)
	podObjects0, _ := utils.GenerateNPods("pod-node0", 1, utils.PodArg{Labels: map[string]string{"app": "pod"}, Nodename: "node0"})
	podObjects1, _ := utils.GenerateNPods("pod-node1", 1, utils.PodArg{Labels: map[string]string{"app": "pod"}, Nodename: "node1"})
	daemonObjects0, _ := utils.GenerateNPods("daemon-node0", 1, utils.PodArg{Labels: map[string]string{"app.kubernetes.io/component": "chaos-daemon"}, Nodename: "node0"})
	daemonObjects1, _ := utils.GenerateNPods("daemon-node1", 1, utils.PodArg{Labels: map[string]string{"app.kubernetes.io/component": "chaos-daemon"}, Nodename: "node1"})

	allObjects := append(nodeObjects, daemonObjects0[0], podObjects0[0], daemonObjects1[0], podObjects1[0])
	g.Expect(v1alpha1.SchemeBuilder.AddToScheme(kubectlscheme.Scheme)).To(BeNil())
	client := fake.NewFakeClientWithScheme(kubectlscheme.Scheme, allObjects...)

	tests := []struct {
		name              string
		chaosSelector     v1alpha1.PodSelectorSpec
		chaosStatus       v1alpha1.ChaosStatus
		wait              bool
		expectedPodNum    int
		expectedDaemonNum int
		expectedErr       bool
	}{
		{
			name:              "chaos on two pods",
			chaosSelector:     v1alpha1.PodSelectorSpec{LabelSelectors: map[string]string{"app": "pod"}},
			chaosStatus:       v1alpha1.ChaosStatus{},
			expectedPodNum:    2,
			expectedDaemonNum: 2,
			expectedErr:       false,
		},
		{
			name: "chaos on one pod",
			chaosSelector: v1alpha1.PodSelectorSpec{
				Nodes:          []string{"node0"},
				LabelSelectors: map[string]string{"app": "pod"},
			},
			chaosStatus:       v1alpha1.ChaosStatus{},
			expectedPodNum:    1,
			expectedDaemonNum: 1,
			expectedErr:       false,
		},
		{
			name:          "wrong selector to get pod",
			chaosSelector: v1alpha1.PodSelectorSpec{LabelSelectors: map[string]string{"app": "oops"}},
			chaosStatus:   v1alpha1.ChaosStatus{},
			expectedErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			timeBefore := time.Now()
			pods, daemons, err := GetPods(context.Background(), test.name, test.chaosStatus, test.chaosSelector, client)
			if test.wait {
				g.Expect(time.Now().Add(time.Millisecond * -50).Before(timeBefore)).To(BeTrue())
			}
			if test.expectedErr {
				g.Expect(err).NotTo(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(len(pods)).To(Equal(test.expectedPodNum))
				g.Expect(len(daemons)).To(Equal(test.expectedDaemonNum))
			}
		})
	}
}
