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

	type TestCase struct {
		name          string
		orignRunner   *Runner
		newRunner     *Runner
		expectedValue bool
	}

	tcs := []TestCase{
		{
			name: "same runner",
			orignRunner: &Runner{
				Name: "t1",
				Rule: "* * * 1 *",
				Job:  fakeJob2{},
			},
			newRunner: &Runner{
				Name: "t1",
				Rule: "* * * 1 *",
				Job:  fakeJob2{},
			},
			expectedValue: true,
		},
		{
			name: "different rule",
			orignRunner: &Runner{
				Name: "t1",
				Rule: "* * * 1 *",
				Job:  fakeJob2{},
			},
			newRunner: &Runner{
				Name: "t1",
				Rule: "@every 2m",
				Job:  fakeJob2{},
			},
			expectedValue: false,
		},
		{
			name: "different name",
			orignRunner: &Runner{
				Name: "t1",
				Rule: "* * * 1 *",
				Job:  fakeJob2{},
			},
			newRunner: &Runner{
				Name: "t2",
				Rule: "* * * 1 *",
				Job:  fakeJob2{},
			},
			expectedValue: false,
		},
		{
			name: "different job",
			orignRunner: &Runner{
				Name: "t1",
				Rule: "* * * 1 *",
				Job:  fakeJob{},
			},
			newRunner: &Runner{
				Name: "t1",
				Rule: "* * * 1 *",
				Job:  fakeJob2{},
			},
			expectedValue: false,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.orignRunner.Equal(tc.newRunner)).To(Equal(tc.expectedValue), tc.name)
	}
}

func TestRunnerValidate(t *testing.T) {
	g := NewGomegaWithT(t)

	type TestCase struct {
		name           string
		runner         *Runner
		expectedResult resultF
	}

	tcs := []TestCase{
		{
			name: "valid",
			runner: &Runner{
				Name: "t1",
				Rule: "* 1 * * *",
				Job:  fakeJob{},
			},
			expectedResult: Succeed,
		},
		{
			name: "no name",
			runner: &Runner{
				Name: "",
				Rule: "* 1 * * *",
				Job:  fakeJob{},
			},
			expectedResult: HaveOccurred,
		},
		{
			name: "no rule",
			runner: &Runner{
				Name: "t1",
				Rule: "",
				Job:  fakeJob{},
			},
			expectedResult: HaveOccurred,
		},
		{
			name: "no job",
			runner: &Runner{
				Name: "t1",
				Rule: "",
				Job:  nil,
			},
			expectedResult: HaveOccurred,
		},
		{
			name: "invalid rule",
			runner: &Runner{
				Name: "t1",
				Rule: "* 1 * * * *",
				Job:  nil,
			},
			expectedResult: HaveOccurred,
		},
	}

	for _, tc := range tcs {
		g.Expect(tc.runner.Validate()).Should(tc.expectedResult(), tc.name)
	}
}
