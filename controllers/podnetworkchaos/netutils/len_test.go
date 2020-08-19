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

package netutils

import (
	"testing"

	. "github.com/onsi/gomega"
)

func Test_compressName(t *testing.T) {
	g := NewWithT(t)

	t.Run("compress name", func(t *testing.T) {
		name := `Running tool: /usr/bin/go test -timeout 30s github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/netutils -run ^Test_compressName$`

		name = CompressName(name, 20, "test")

		g.Expect(name).Should(Equal("Runni_a5e4631cf_test"))

		name = "test executed panic(nil) or runtime.Goexit: subtest may have called FailNow on a parent test"

		name = CompressName(name, 13, "test")

		g.Expect(name).Should(Equal("test _03_test"))
	})
}
