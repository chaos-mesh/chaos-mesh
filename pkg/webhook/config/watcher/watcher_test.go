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

package watcher

import (
	"fmt"
	v1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("webhook config watcher", func() {
	Context("New", func() {
		It("should return maybe you should specify --configmap-namespace", func() {
			config := NewConfig()
			configWatcher, err := New(*config)
			Expect(configWatcher).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("maybe you should specify --configmap-namespace"))
		})

		It("should return InClusterConfig error", func() {
			old := restInClusterConfig
			defer func() { restInClusterConfig = old }()

			restInClusterConfig = func() (*rest.Config, error) {
				return nil, fmt.Errorf("InClusterConfig error")
			}
			config := NewConfig()
			config.Namespace = "testNamespace"
			configWatcher, err := New(*config)
			Expect(configWatcher).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("InClusterConfig"))
		})

		It("should return NewForConfig error", func() {
			restInClusterConfig = MockInClusterConfig
			old := kubernetesNewForConfig
			defer func() { kubernetesNewForConfig = old }()

			kubernetesNewForConfig = func(c *rest.Config) (*kubernetes.Clientset, error) {
				return nil, fmt.Errorf("NewForConfig error")
			}
			config := NewConfig()
			config.Namespace = "testNamespace"
			configWatcher, err := New(*config)
			Expect(configWatcher).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("NewForConfig"))
		})

		It("should return no error", func() {
			restInClusterConfig = MockInClusterConfig

			config := NewConfig()
			config.Namespace = "testNamespace"
			configWatcher, err := New(*config)
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

		It("should return configmap labels was an uninitialized map", func() {
			var cmw K8sConfigMapWatcher
			cmw.Namespace = "testNamespace"
			err := validate(&cmw)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("configmap labels was an uninitialized map"))
		})

		It("should return k8s client was not setup properly", func() {
			var cmw K8sConfigMapWatcher
			cmw.Namespace = "testNamespace"
			cmw.ConfigMapLabels = make(map[string]string)
			err := validate(&cmw)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("k8s client was not setup properly"))
		})
	})

	Context("Watch error", func() {
		It("should return unable to create watcher", func() {
			var cmw K8sConfigMapWatcher
			cmw.Config = *NewConfig()
			cmw.Namespace = "testNamespace"
			k8sConfig, _ := MockInClusterConfig()
			clientset, _ := kubernetesNewForConfig(k8sConfig)
			cmw.client = clientset.CoreV1()
			sigChan := make(chan interface{}, 10)
			stopCh := ctrl.SetupSignalHandler()
			err := cmw.Watch(sigChan, stopCh)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("unable to create watcher"))
		})
	})

	Context("get", func() {
		It("should return error when ConfigMaps.List", func() {
			var cmw K8sConfigMapWatcher
			cmw.Config = *NewConfig()
			cmw.Namespace = "testNamespace"
			k8sConfig, _ := MockInClusterConfig()
			clientset, _ := kubernetesNewForConfig(k8sConfig)
			cmw.client = clientset.CoreV1()
			_, err := cmw.Get()
			Expect(err).ToNot(BeNil())
		})
	})

	Context("InjectionConfigsFromConfigMap", func() {
		It("should return nil", func() {
			var cm v1.ConfigMap
			_, err := InjectionConfigsFromConfigMap(cm)
			Expect(err).To(BeNil())
		})

		It("should return error parsing ConfigMap", func() {
			var cm v1.ConfigMap
			cm.Data = make(map[string]string)
			cm.Data["testkey"] = "testvalue"
			_, err := InjectionConfigsFromConfigMap(cm)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(ContainSubstring("error parsing ConfigMap"))
		})

		It("should return nil", func() {
			var cm v1.ConfigMap
			cm.Data = make(map[string]string)
			cm.Data["testkey"] = "name: \"testname\""
			_, err := InjectionConfigsFromConfigMap(cm)
			Expect(err).To(BeNil())  //error parsing ConfigMap
		})
	})
})
