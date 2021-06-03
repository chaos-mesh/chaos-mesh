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

package stresschaos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func makeMemoryStressChaos(
	namespace, name string,
	podNs, podAppName string, memorySize string, worker int,
) *v1alpha1.StressChaos {
	return &v1alpha1.StressChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.StressChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode: v1alpha1.AllPodMode,
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces: []string{podNs},
						LabelSelectors: map[string]string{
							"app": podAppName,
						},
					},
				},
			},
			Stressors: &v1alpha1.Stressors{
				MemoryStressor: &v1alpha1.MemoryStressor{
					Size:     memorySize,
					Stressor: v1alpha1.Stressor{Workers: worker},
				},
			},
		},
	}
}

func makeCPUStressChaos(
	namespace, name string,
	podNs, podAppName string, worker int, load int,
) *v1alpha1.StressChaos {
	return &v1alpha1.StressChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.StressChaosSpec{
			ContainerSelector: v1alpha1.ContainerSelector{
				PodSelector: v1alpha1.PodSelector{
					Mode: v1alpha1.AllPodMode,
					Selector: v1alpha1.PodSelectorSpec{
						Namespaces: []string{podNs},
						LabelSelectors: map[string]string{
							"app": podAppName,
						},
					},
				},
			},
			Stressors: &v1alpha1.Stressors{
				CPUStressor: &v1alpha1.CPUStressor{
					Load:     &load,
					Stressor: v1alpha1.Stressor{Workers: worker},
				},
			},
		},
	}
}

type StressCondition struct {
	CpuTime     uint64 `json:"cpuTime"`
	MemoryUsage uint64 `json:"memoryUsage"`
}

func getStressCondition(c http.Client, port uint16) (*StressCondition, error) {
	klog.Infof("sending request to http://localhost:%d/stress", port)

	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/stress", port))
	if err != nil {
		return nil, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	condition := &StressCondition{}
	err = json.Unmarshal(out, condition)
	if err != nil {
		return nil, err
	}

	return condition, nil
}

func probeStressCondition(
	c http.Client, peers []*corev1.Pod, ports []uint16,
) (map[int]*StressCondition, error) {
	stressConditions := make(map[int]*StressCondition)

	for index, port := range ports {
		stressCondition, err := getStressCondition(c, port)
		if err != nil {
			return nil, err
		}

		stressConditions[index] = stressCondition
	}

	return stressConditions, nil
}
