// Copyright Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v1alpha1

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("statuscheck_webhook", func() {
	Context("webhook.Defaultor of statuscheck", func() {
		It("Default", func() {
			statusCheck := &StatusCheck{}
			statusCheck.Default()
			Expect(statusCheck.Spec.Mode).To(Equal(StatusCheckSynchronous))
		})
	})
	Context("webhook.Validator of statuscheck", func() {
		It("Validate", func() {
			type TestCase struct {
				name        string
				statusCheck StatusCheck
				expect      string
			}
			tcs := []TestCase{
				{
					name: "simple Validate",
					statusCheck: StatusCheck{
						Spec: StatusCheckSpec{
							Type: TypeHTTP,
							EmbedStatusCheck: &EmbedStatusCheck{
								HTTPStatusCheck: &HTTPStatusCheck{
									RequestUrl: "http://1.1.1.1",
									Criteria: HTTPCriteria{
										StatusCode: "200",
									},
								},
							},
						},
					},
					expect: "",
				},
				{
					name: "simple Validate with status code range",
					statusCheck: StatusCheck{
						Spec: StatusCheckSpec{
							Type: TypeHTTP,
							EmbedStatusCheck: &EmbedStatusCheck{
								HTTPStatusCheck: &HTTPStatusCheck{
									RequestUrl: "http://1.1.1.1",
									Criteria: HTTPCriteria{
										StatusCode: "200-400",
									},
								},
							},
						},
					},
					expect: "",
				},
				{
					name: "unknown type",
					statusCheck: StatusCheck{
						Spec: StatusCheckSpec{
							Type: "CMD",
						},
					},
					expect: "unrecognized type",
				},
				{
					name: "invalid request url",
					statusCheck: StatusCheck{
						Spec: StatusCheckSpec{
							Type: TypeHTTP,
							EmbedStatusCheck: &EmbedStatusCheck{
								HTTPStatusCheck: &HTTPStatusCheck{
									RequestUrl: "1.1.1.1",
									Criteria: HTTPCriteria{
										StatusCode: "-1",
									},
								},
							},
						},
					},
					expect: "invalid http request url",
				},
				{
					name: "invalid status code",
					statusCheck: StatusCheck{
						Spec: StatusCheckSpec{
							Type: TypeHTTP,
							EmbedStatusCheck: &EmbedStatusCheck{
								HTTPStatusCheck: &HTTPStatusCheck{
									RequestUrl: "http://1.1.1.1",
									Criteria: HTTPCriteria{
										StatusCode: "-1",
									},
								},
							},
						},
					},
					expect: "invalid status code",
				},
				{
					name: "invalid status code range",
					statusCheck: StatusCheck{
						Spec: StatusCheckSpec{
							Type: TypeHTTP,
							EmbedStatusCheck: &EmbedStatusCheck{
								HTTPStatusCheck: &HTTPStatusCheck{
									RequestUrl: "http://1.1.1.1",
									Criteria: HTTPCriteria{
										StatusCode: "200-x",
									},
								},
							},
						},
					},
					expect: "incorrect status code format",
				},
			}

			for _, tc := range tcs {
				err := tc.statusCheck.ValidateCreate()
				if len(tc.expect) != 0 {
					Expect(err).To(HaveOccurred())
					Expect(strings.Contains(err.Error(), tc.expect)).To(BeTrue(), "expected error: %s, got: %s", tc.expect, err.Error())
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			}
		})
	})
})
