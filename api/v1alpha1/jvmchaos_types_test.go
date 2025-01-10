// Copyright 2021 Chaos Mesh Authors.
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

package v1alpha1

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// These tests are written in BDD-style using Ginkgo framework. Refer to
// http://onsi.github.io/ginkgo to learn more.

var _ = Describe("JVMChaos", func() {
	var (
		key              types.NamespacedName
		created, fetched *JVMChaos
	)

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Create API", func() {
		It("should create an object successfully", func() {
			key = types.NamespacedName{
				Name:      "foo",
				Namespace: "default",
			}

			created = &JVMChaos{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Spec: JVMChaosSpec{
					Action: JVMLatencyAction,
					JVMParameter: JVMParameter{
						JVMClassMethodSpec: JVMClassMethodSpec{
							Class:  "Main",
							Method: "print",
						},
						LatencyDuration: 1000,
					},
					ContainerSelector: ContainerSelector{
						PodSelector: PodSelector{
							Mode: OneMode,
						},
					},
				},
			}

			By("creating an API obj")
			Expect(k8sClient.Create(context.TODO(), created)).To(Succeed())

			fetched = &JVMChaos{}
			Expect(k8sClient.Get(context.TODO(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the created object")
			Expect(k8sClient.Delete(context.TODO(), created)).To(Succeed())
			Expect(k8sClient.Get(context.TODO(), key, created)).ToNot(Succeed())
		})
	})

	Describe("JSON Marshal and Unmarshal", func() {
		object := &JVMChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			Spec: JVMChaosSpec{
				ContainerSelector: ContainerSelector{
					PodSelector: PodSelector{
						Mode:  FixedMode,
						Value: "1337",
					},
				},
				JVMParameter: JVMParameter{
					Name:        "param-name",
					ReturnValue: "param-return-value",
				},
			},
		}
		It("should marshal an object successfully", func() {
			By("marshalling the object")
			data, _ := json.MarshalIndent(object, "", "  ")
			Expect(data).To(MatchJSON(`{
  "metadata": {
    "name": "foo",
    "namespace": "default",
    "creationTimestamp": null
  },
  "spec": {
    "selector": {},
    "mode": "fixed",
    "value": "1337",
    "action": "",
    "name": "param-name",
    "returnValue": "param-return-value",
    "exception": "",
    "latency": 0,
    "ruleData": ""
  },
  "status": {
    "experiment": {}
  }
}`))
		})
		It("should unmarshal the object successfully", func() {
			marshalledJson := []byte(`{
  "metadata": {
    "name": "foo",
    "namespace": "default"
  },
  "spec": {
    "name": "param-name",
    "returnValue": "param-return-value",
    "mode": "fixed",
    "value": "1337"
  }
}`)
			By("unmarshalling the object")
			var unmarshalledJVMChaos JVMChaos
			Expect(json.Unmarshal(marshalledJson, &unmarshalledJVMChaos)).To(Succeed())
			Expect(&unmarshalledJVMChaos).To(HaveField("ObjectMeta.Name", "foo"))
			Expect(&unmarshalledJVMChaos).To(HaveField("Spec.ContainerSelector.PodSelector.Mode", FixedMode))
			Expect(&unmarshalledJVMChaos).To(HaveField("Spec.ContainerSelector.PodSelector.Value", "1337"))
			Expect(&unmarshalledJVMChaos).To(HaveField("Spec.JVMParameter.Name", "param-name"))
			Expect(&unmarshalledJVMChaos).To(HaveField("Spec.JVMParameter.ReturnValue", "param-return-value"))
		})
	})
})
