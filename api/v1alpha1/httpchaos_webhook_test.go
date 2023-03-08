// Copyright 2021 Chaos Mesh Authors.
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
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("HTTPChaos Webhook", func() {
	Context("webhook.Validator of httpchaos", func() {
		It("Validate", func() {
			type TestCase struct {
				name    string
				chaos   HTTPChaos
				execute func(chaos *HTTPChaos) error
				expect  string
			}
			errorDuration := "400S"
			errorMethod := "gET"
			validMethod := http.MethodGet
			errorDelay := "1"
			valideDelay := "1s"

			tcs := []TestCase{
				{
					name: "simple ValidateCreate",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo1",
						},
						Spec: HTTPChaosSpec{
							Target: PodHttpRequest,
							Port:   80,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "",
				},
				{
					name: "simple ValidateUpdate",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo2",
						},
						Spec: HTTPChaosSpec{
							Target: PodHttpRequest,
							Port:   80,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateUpdate(chaos)
					},
					expect: "",
				},
				{
					name: "simple ValidateDelete",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo3",
						},
						Spec: HTTPChaosSpec{
							Target: PodHttpRequest,
							Port:   80,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateDelete()
					},
					expect: "",
				},
				{
					name: "parse the duration error",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo6",
						},
						Spec: HTTPChaosSpec{
							Duration: &errorDuration,
							Target:   PodHttpRequest,
							Port:     80,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo7",
						},
						Spec: HTTPChaosSpec{
							PodSelector: PodSelector{
								Value: "0",
								Mode:  FixedMode,
							},
							Port:   80,
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode, parse value error",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo8",
						},
						Spec: HTTPChaosSpec{
							PodSelector: PodSelector{
								Value: "num",
								Mode:  FixedMode,
							},
							Port:   80,
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentMode",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo9",
						},
						Spec: HTTPChaosSpec{
							PodSelector: PodSelector{
								Value: "0",
								Mode:  RandomMaxPercentMode,
							},
							Port:   80,
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with RandomMaxPercentMode ,parse value error",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo10",
						},
						Spec: HTTPChaosSpec{
							PodSelector: PodSelector{
								Value: "num",
								Mode:  RandomMaxPercentMode,
							},
							Port:   80,
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate value with FixedPercentMode",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo11",
						},
						Spec: HTTPChaosSpec{
							PodSelector: PodSelector{
								Value: "101",
								Mode:  FixedPercentMode,
							},
							Port:   80,
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate port 1",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo12",
						},
						Spec: HTTPChaosSpec{
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate port 2",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo13",
						},
						Spec: HTTPChaosSpec{
							Port:   -1,
							Target: PodHttpRequest,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate target 1",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo14",
						},
						Spec: HTTPChaosSpec{
							Port: 80,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "validate target 2",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo15",
						},
						Spec: HTTPChaosSpec{
							Port:   80,
							Target: "request",
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "valid method 1",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo16",
						},
						Spec: HTTPChaosSpec{
							Port:   80,
							Target: PodHttpRequest,
							Method: &validMethod,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "ok",
				},
				{
					name: "valid method 2",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo17",
						},
						Spec: HTTPChaosSpec{
							Port:   80,
							Target: PodHttpResponse,
							Method: &errorMethod,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "ok",
				},
				{
					name: "invalid method",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo18",
						},
						Spec: HTTPChaosSpec{
							Port:   80,
							Target: PodHttpRequest,
							Method: &errorMethod,
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
				{
					name: "valid delay",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo19",
						},
						Spec: HTTPChaosSpec{
							Port:   80,
							Target: PodHttpRequest,
							PodHttpChaosActions: PodHttpChaosActions{
								Delay: &valideDelay,
							},
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "ok",
				},
				{
					name: "invalid delay",
					chaos: HTTPChaos{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: metav1.NamespaceDefault,
							Name:      "foo20",
						},
						Spec: HTTPChaosSpec{
							Port:   80,
							Target: PodHttpRequest,
							PodHttpChaosActions: PodHttpChaosActions{
								Delay: &errorDelay,
							},
						},
					},
					execute: func(chaos *HTTPChaos) error {
						return chaos.ValidateCreate()
					},
					expect: "error",
				},
			}

			for _, tc := range tcs {
				err := tc.execute(&tc.chaos)
				if tc.expect == "error" {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			}
		})
	})
})
