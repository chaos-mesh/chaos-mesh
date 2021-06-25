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

package provider

import (
	"context"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("UpdatedClient", func() {

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context(("UpdatedClient"), func() {
		It(("Should create and delete successfully"), func() {
			obj := &corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
					Name:      "test-configmap-create-delete",
				},
				Data: map[string]string{
					"test": "1",
				},
			}
			err := k8sClient.Create(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())

			err = k8sClient.Delete(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Data should always be updated", func() {
			obj := &corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
					Name:      "test-configmap-update",
				},
				Data: map[string]string{
					"test": "0",
				},
			}
			err := k8sClient.Create(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())

			for i := 0; i <= 200; i++ {
				data := strconv.Itoa(i)

				obj.Data["test"] = data
				err = k8sClient.Update(context.TODO(), obj)
				Expect(err).ToNot(HaveOccurred())

				err = k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "test-configmap-update",
				}, obj)
				Expect(err).ToNot(HaveOccurred())

				Expect(obj.Data["test"]).To(Equal(data))
			}

			err = k8sClient.Delete(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Newer data should be returned", func() {
			obj := &corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "default",
					Name:      "test-configmap-another-update",
				},
				Data: map[string]string{
					"test": "0",
				},
			}
			err := k8sClient.Create(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())

			obj.Data["test"] = "1"
			err = k8sClient.Update(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())

			newObj := &corev1.ConfigMap{}
			err = k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "test-configmap-another-update",
			}, newObj)
			Expect(err).ToNot(HaveOccurred())

			Expect(newObj.Data["test"]).To(Equal("1"))
			newObj.Data["test"] = "2"
			anotherCleanClient := mgr.GetClient()
			err = anotherCleanClient.Update(context.TODO(), obj)
			Expect(err).ToNot(HaveOccurred())

			newObj = &corev1.ConfigMap{}
			err = k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "test-configmap-another-update",
			}, newObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(newObj.Data["test"]).To(Equal("2"))
		})
	})
})
