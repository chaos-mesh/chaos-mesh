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

package manager

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestRunnerEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	r1 := &Runner{
		Name: "t1",
		Rule: "* * * 1 *",
		Job:  fakeJob2{},
	}
	r2 := &Runner{
		Name: "t1",
		Rule: "* * * 1 *",
		Job:  fakeJob2{},
	}

	g.Expect(r1.Equal(r2)).To(Equal(true))

	r2.Name = "t2"
	g.Expect(r1.Equal(r2)).To(Equal(false))

	r2.Name = "t1"
	r2.Rule = "* * 1 * *"
	g.Expect(r1.Equal(r2)).To(Equal(false))

	r2.Name = "t1"
	r2.Rule = "* * * 1 *"
	g.Expect(r1.Equal(r2)).To(Equal(true))

	r1.Job = fakeJob{}
	g.Expect(r1.Equal(r2)).To(Equal(false))
}

func TestRunnerValidate(t *testing.T) {
	g := NewGomegaWithT(t)

	r1 := &Runner{
		Rule: "* * * 1 *",
		Job:  fakeJob2{},
	}
	g.Expect(r1.Validate()).Should(HaveOccurred())

	r1.Name = "t1"
	g.Expect(r1.Validate()).ShouldNot(HaveOccurred())

	r1.Rule = ""
	g.Expect(r1.Validate()).Should(HaveOccurred())

	r1.Rule = "* * * * * *"
	g.Expect(r1.Validate()).Should(HaveOccurred())

	r1.Rule = "* * 3 * *"
	g.Expect(r1.Validate()).ShouldNot(HaveOccurred())

	r1.Job = nil
	g.Expect(r1.Validate()).Should(HaveOccurred())
}
