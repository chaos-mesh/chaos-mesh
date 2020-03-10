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

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("networkchaos_webhook", func() {
	Context("Defaulter", func() {
		It("set default namespace selector", func() {
			networkchaos := &NetworkChaos{
				ObjectMeta: metav1.ObjectMeta{Namespace: metav1.NamespaceDefault},
			}
			networkchaos.Default()
			Expect(networkchaos.Spec.Selector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
			Expect(networkchaos.Spec.Target.TargetSelector.Namespaces[0]).To(Equal(metav1.NamespaceDefault))
		})
	})
})
