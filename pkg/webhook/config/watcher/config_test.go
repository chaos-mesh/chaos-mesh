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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("webhook config watcher", func() {
	Context("Test webhook config", func() {
		It("should return NewConfig", func() {
			config := NewConfig()
			Expect(config.Namespace).To(Equal(""))
			Expect(config.ConfigMapLabels).To(Equal(map[string]string{}))
		})

	})
})
