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

package config

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/webhook/config/watcher"
)

func TestValidations(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Namespace scoped",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = Describe("Namespace-scoped Chaos", func() {
	Context("Namespace-scoped Chaos Configuration Validation", func() {
		It("Validation", func() {
			type TestCase struct {
				name        string
				config      config.ChaosControllerConfig
				expectValid bool
			}

			testCases := []TestCase{
				{
					name: "target namespace should not be empty with namespaced scope",
					config: config.ChaosControllerConfig{
						WatcherConfig: &watcher.Config{
							ClusterScoped:     false,
							TemplateNamespace: "",
							TargetNamespace:   "",
							TemplateLabels:    nil,
							ConfigLabels:      nil,
						},
						ClusterScoped:   false,
						TargetNamespace: "",
					},
					expectValid: false,
				},
				{
					name: "Watcher Config is always required",
					config: config.ChaosControllerConfig{
						WatcherConfig:   nil,
						ClusterScoped:   true,
						TargetNamespace: "",
					},
					expectValid: false,
				},
				{
					name: "clusterScope should keep constant",
					config: config.ChaosControllerConfig{
						WatcherConfig: &watcher.Config{
							ClusterScoped:     false,
							TemplateNamespace: "",
							TargetNamespace:   "",
							TemplateLabels:    nil,
							ConfigLabels:      nil,
						},
						ClusterScoped:   true,
						TargetNamespace: "",
					},
					expectValid: false,
				},
				{
					name: "ns should keep constant",
					config: config.ChaosControllerConfig{
						WatcherConfig: &watcher.Config{
							ClusterScoped:   false,
							TargetNamespace: "ns1",
						},
						ClusterScoped:   false,
						TargetNamespace: "ns2",
					},
					expectValid: false,
				},
			}

			for _, testCase := range testCases {
				By(testCase.name)
				err := validate(&testCase.config)
				if testCase.expectValid {
					Expect(err).NotTo(HaveOccurred())
				} else {
					Expect(err).To(HaveOccurred())
				}
			}
		})
	})
})

func TestIsAllowedNamespaces(t *testing.T) {
	g := NewGomegaWithT(t)
	type TestCase struct {
		name   string
		pods   []v1.Pod
		ret    []bool
		allow  string
		ignore string
	}
	pods := []v1.Pod{
		newPod("p1", v1.PodRunning, "allow", nil, nil, ""),
		newPod("p1", v1.PodRunning, "allow-app", nil, nil, ""),
		newPod("p1", v1.PodRunning, "app-allow", nil, nil, ""),
		newPod("p1", v1.PodRunning, "ignore", nil, nil, ""),
		newPod("p1", v1.PodRunning, "ignore-app", nil, nil, ""),
		newPod("p1", v1.PodRunning, "app-ignore", nil, nil, ""),
	}

	allowRet := []bool{true, true, true, false, false, false}

	var tcs []TestCase
	tcs = append(tcs, TestCase{
		name:  "only set allow",
		pods:  pods,
		ret:   allowRet,
		allow: "allow",
	})

	tcs = append(tcs, TestCase{
		name:   "only set ignore",
		pods:   pods,
		ret:    allowRet,
		ignore: "ignore",
	})

	tcs = append(tcs, TestCase{
		name:   "only set allow",
		pods:   pods,
		ret:    allowRet,
		allow:  "allow",
		ignore: "ignore",
	})

	for _, tc := range tcs {
		for index, pod := range tc.pods {
			g.Expect(IsAllowedNamespaces(pod.Namespace, tc.allow, tc.ignore)).Should(Equal(tc.ret[index]))
		}
	}
}

// TODO: reuse this function
func newPod(
	name string,
	status v1.PodPhase,
	namespace string,
	ans map[string]string,
	ls map[string]string,
	nodename string,
) v1.Pod {
	return v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Labels:      ls,
			Annotations: ans,
		},
		Spec: v1.PodSpec{
			NodeName: nodename,
		},
		Status: v1.PodStatus{
			Phase: status,
		},
	}
}
