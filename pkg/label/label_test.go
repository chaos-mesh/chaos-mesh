// Copyright 2019 Chaos Mesh Authors.
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
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLabelString(t *testing.T) {
	g := NewGomegaWithT(t)

	la := Label(make(map[string]string))
	la["test-label-1"] = "t1"
	la["test-label-2"] = "t2"

	g.Expect(strings.Contains(la.String(), "test-label-1=t1")).To(Equal(true))
	g.Expect(strings.Contains(la.String(), "test-label-2=t2")).To(Equal(true))
	g.Expect(strings.Contains(la.String(), ",")).To(Equal(true))

	g.Expect(len(la.String())).To(Equal(len("test-label-1=t1,test-label-2=t2")))

	la[""] = "t3"
	g.Expect(len(la.String())).To(Equal(len("test-label-1=t1,test-label-2=t2")))
	g.Expect(strings.Contains(la.String(), "t3")).To(Equal(false))
}
