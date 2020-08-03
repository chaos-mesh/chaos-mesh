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

package chaosdaemon

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("netem server", func() {
	Context("newDaemonServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &MockClient{})()
			_, err := newDaemonServer(containerRuntimeContainerd)
			Expect(err).To(BeNil())
		})

		It("should fail on CreateContainerRuntimeInfoClient", func() {
			_, err := newDaemonServer("invalid-runtime")
			Expect(err).ToNot(BeNil())
		})
	})

	Context("newGRPCServer", func() {
		It("should work", func() {
			defer mock.With("MockContainerdClient", &MockClient{})()
			_, err := newGRPCServer(containerRuntimeContainerd, &MockRegisterer{})
			Expect(err).To(BeNil())
		})

		It("should panic", func() {
			Ω(func() {
				defer mock.With("MockContainerdClient", &MockClient{})()
				defer mock.With("PanicOnMustRegister", "mock panic")()
				newGRPCServer(containerRuntimeContainerd, &MockRegisterer{})
			}).Should(Panic())
		})
	})
})
