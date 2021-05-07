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

package crclients

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/mock"
)

var _ = Describe("chaosdaemon util", func() {
	Context("CreateContainerRuntimeInfoClient", func() {
		It("should work", func() {
			_, err := CreateContainerRuntimeInfoClient(ContainerRuntimeDocker)
			Expect(err).To(BeNil())
			defer func() {
				err := mock.With("MockContainerdClient", &test.MockClient{})()
				Expect(err).To(BeNil())
			}()
			_, err = CreateContainerRuntimeInfoClient(ContainerRuntimeContainerd)
			Expect(err).To(BeNil())
		})

		It("should error on newContaineredClient", func() {
			errorStr := "this is a mocked error"

			defer func() {
				err := mock.With("NewContainerdClientError", errors.New(errorStr))()
				Expect(err).To(BeNil())
			}()
			_, err := CreateContainerRuntimeInfoClient(ContainerRuntimeContainerd)
			Expect(err).ToNot(BeNil())
			Expect(fmt.Sprintf("%s", err)).To(Equal(errorStr))
		})
	})
})
