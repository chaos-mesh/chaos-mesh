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

package watcher

import (
	"fmt"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("webhook config watcher", func() {
	Context("New", func() {
		It("should return InClusterConfig error", func() {
			old := restClusterConfig
			defer func() { restClusterConfig = old }()

			restClusterConfig = func() (*rest.Config, error) {
				return nil, fmt.Errorf("InClusterConfig error")
			}
			config := NewConfig()
			config.TemplateNamespace = "testNamespace"
			configWatcher, err := New(*config, nil)
			Expect(configWatcher).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("InClusterConfig"))
		})

		It("should return NewForConfig error", func() {
			restClusterConfig = MockClusterConfig
			old := kubernetesNewForConfig
			defer func() { kubernetesNewForConfig = old }()

			kubernetesNewForConfig = func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, fmt.Errorf("NewForConfig error")
			}
			config := NewConfig()
			config.TemplateNamespace = "testNamespace"
			configWatcher, err := New(*config, nil)
			Expect(configWatcher).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("NewForConfig"))
		})

		It("should return no error", func() {
			restClusterConfig = MockClusterConfig

			config := NewConfig()
			config.TemplateNamespace = "testNamespace"
			configWatcher, err := New(*config, nil)
			Expect(configWatcher).ToNot(BeNil())
			Expect(err).To(BeNil())
		})
	})

	Context("validate", func() {
		It("should return configmap watcher was nil", func() {
			err := validate(nil)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("configmap watcher was nil"))
		})

		It("should return namespace is empty", func() {
			var cmw K8sConfigMapWatcher
			err := validate(&cmw)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("namespace is empty"))
		})

		It("should return template labels was an uninitialized map", func() {
			var cmw K8sConfigMapWatcher
			cmw.TemplateNamespace = "testNamespace"
			err := validate(&cmw)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("template labels was an uninitialized map"))
		})

		It("should return config labels was an uninitialized map", func() {
			var cmw K8sConfigMapWatcher
			cmw.TemplateNamespace = "testNamespace"
			cmw.TemplateLabels = map[string]string{"test": "test"}
			err := validate(&cmw)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("config labels was an uninitialized map"))
		})
	})

	Context("Watch error", func() {
		It("should return unable to create template watcher", func() {
			var cmw K8sConfigMapWatcher
			cmw.Config = *NewConfig()
			cmw.TemplateNamespace = "testNamespace"
			k8sConfig, _ := MockClusterConfig()
			clientset, _ := kubernetesNewForConfig(k8sConfig)
			cmw.client = clientset.CoreV1()
			sigChan := make(chan interface{}, 10)
			stopCh := ctrl.SetupSignalHandler()
			err := cmw.Watch(sigChan, stopCh)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("unable to create template watcher"))
		})
	})

	Context("get", func() {
		It("should return error when ConfigMaps.List", func() {
			var cmw K8sConfigMapWatcher
			cmw.Config = *NewConfig()
			cmw.TemplateNamespace = "testNamespace"
			k8sConfig, _ := MockClusterConfig()
			clientset, _ := kubernetesNewForConfig(k8sConfig)
			cmw.client = clientset.CoreV1()
			_, err := cmw.GetConfigs()
			Expect(err).ToNot(BeNil())
		})
	})
})
