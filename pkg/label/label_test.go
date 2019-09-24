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

package label

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestLabelString(t *testing.T) {
	g := NewGomegaWithT(t)

	la := Label(make(map[string]string))
	la["test-label-1"] = "t1"
	la["test-label-2"] = "t2"

	g.Expect(la.String()).To(Equal("test-label-1=t1,test-label-2=t2"))

	la[""] = "t3"
	g.Expect(la.String()).To(Equal("test-label-1=t1,test-label-2=t2"))
}
