// Copyright 2026 Chaos Mesh Authors.
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

package steps

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/cucumber/godog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/kubernetes/test/e2e/framework"
)

var _ = Describe("Scenario namespaces", func() {
	DescribeTable("reports cleanup failures", func(verb string) {
		cli := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}})
		cli.PrependReactor("delete", "namespaces", func(clienttesting.Action) (bool, runtime.Object, error) { return true, nil, nil })
		cli.PrependReactor(verb, "namespaces", func(clienttesting.Action) (bool, runtime.Object, error) {
			return true, nil, apierrors.NewForbidden(schema.GroupResource{Resource: "namespaces"}, "test", fmt.Errorf("denied"))
		})
		tc := &TestContext{KubeCli: cli, Namespace: "test"}
		err := tc.cleanupNamespace(context.Background())
		Expect(apierrors.IsForbidden(err)).To(BeTrue())
		Expect(err.Error()).To(ContainSubstring("test"))
	}, Entry("delete fails", "delete"), Entry("get fails", "get"))

	DescribeTable("accepts an absent namespace", func(exists bool) {
		cli := fake.NewSimpleClientset()
		if exists {
			_, err := cli.CoreV1().Namespaces().Create(context.Background(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
		}
		tc := &TestContext{KubeCli: cli, Namespace: "test"}
		Expect(tc.cleanupNamespace(context.Background())).To(Succeed())
	}, Entry("already absent", false), Entry("deleted", true))

	It("fails when namespace deletion never completes", func() {
		cli := fake.NewSimpleClientset(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}})
		cli.PrependReactor("delete", "namespaces", func(clienttesting.Action) (bool, runtime.Object, error) { return true, nil, nil })
		tc := &TestContext{KubeCli: cli, Namespace: "test"}
		ctx, cancel := shortContext()
		defer cancel()
		Expect(tc.cleanupNamespace(ctx)).To(MatchError(context.DeadlineExceeded))
	})

	It("retains a created namespace when service account readiness fails", func() {
		previous := framework.TestContext.VerifyServiceAccount
		framework.TestContext.VerifyServiceAccount = true
		defer func() { framework.TestContext.VerifyServiceAccount = previous }()
		cli := fake.NewSimpleClientset()
		ctx, cancel := shortContext()
		defer cancel()
		tc := &TestContext{KubeCli: cli}
		Expect(tc.aNamespaceIsPrepared(ctx)).To(HaveOccurred())
		Expect(tc.Namespace).NotTo(BeEmpty())
		_, err := cli.CoreV1().Namespaces().Get(context.Background(), tc.Namespace, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(tc.cleanupNamespace(context.Background())).To(Succeed())
	})

	DescribeTable("runs cleanup despite a canceled scenario and reports its own error once", func(deny bool, bodyFails bool) {
		deletes := 0
		tc := &TestContext{Namespace: "test", OriginalImage: "old", InitialPods: []corev1.Pod{{}}, ContainerBaseline: &recordedContainer{}}
		tc.KubeCli = requestClient(func(req *http.Request) (*http.Response, error) {
			Expect(req.Context().Err()).NotTo(HaveOccurred())
			_, bounded := req.Context().Deadline()
			Expect(bounded).To(BeTrue())
			if req.Method == http.MethodDelete {
				deletes++
				if deny {
					response := jsonResponse(&metav1.Status{Status: "Failure", Reason: metav1.StatusReasonForbidden, Code: 403, Message: "cleanup denied"})
					response.StatusCode = 403
					return response, nil
				}
				return jsonResponse(&metav1.Status{Status: "Success"}), nil
			}
			response := jsonResponse(&metav1.Status{Status: "Failure", Reason: metav1.StatusReasonNotFound, Code: 404})
			response.StatusCode = 404
			return response, nil
		})
		var output bytes.Buffer
		suite := godog.TestSuite{Name: "cleanup", ScenarioInitializer: func(sc *godog.ScenarioContext) {
			tc.RegisterSteps(sc)
			sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
				ctx, cancel := context.WithCancel(ctx)
				cancel()
				return ctx, nil
			})
			sc.Step(`^the scenario runs$`, func() error {
				if bodyFails {
					return fmt.Errorf("body failed")
				}
				return nil
			})
		}, Options: &godog.Options{Format: "progress", NoColors: true, Output: &output, FeatureContents: []godog.Feature{{Name: "cleanup.feature", Contents: []byte("Feature: cleanup\nScenario: cleanup\nGiven the scenario runs\n")}}}}
		code := suite.Run()
		if deny || bodyFails {
			Expect(code).NotTo(BeZero())
		} else {
			Expect(code).To(BeZero(), output.String())
		}
		if deny {
			Expect(output.String()).To(ContainSubstring("cleanup denied"))
		}
		if bodyFails {
			Expect(output.String()).To(ContainSubstring("body failed"))
		}
		Expect(deletes).To(Equal(1))
		Expect(tc.Namespace).To(BeEmpty())
		Expect(tc.InitialPods).To(BeNil())
		Expect(tc.ContainerBaseline).To(BeNil())
		Expect(tc.OriginalImage).To(BeEmpty())
	}, Entry("successful cleanup", false, false), Entry("cleanup failure", true, false), Entry("body and cleanup failure", true, true))
})
