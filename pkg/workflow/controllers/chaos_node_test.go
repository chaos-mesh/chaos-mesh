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

package controllers

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// integration tests
var _ = Describe("Workflow", func() {
	var ns string
	BeforeEach(func() {
		ctx := context.TODO()
		newNs := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "chaos-mesh-",
			},
			Spec: corev1.NamespaceSpec{},
		}
		Expect(kubeClient.Create(ctx, &newNs)).To(Succeed())
		ns = newNs.Name
		By(fmt.Sprintf("create new namespace %s", ns))
	})

	AfterEach(func() {
		ctx := context.TODO()
		nsToDelete := corev1.Namespace{}
		Expect(kubeClient.Get(ctx, types.NamespacedName{Name: ns}, &nsToDelete)).To(Succeed())
		Expect(kubeClient.Delete(ctx, &nsToDelete)).To(Succeed())
		By(fmt.Sprintf("cleanupz namespace %s", ns))
	})

	Context("one chaos node", func() {
		It("could spawn one chaos", func() {
			// TODO: unfinished test case
		})

		It("could spawn one schedule", func() {
			// TODO: unfinished test case
		})
	})
})
