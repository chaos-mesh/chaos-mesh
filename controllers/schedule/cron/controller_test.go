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

package cron

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestCron(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schedule cron suite")
}

var _ = Describe("Reconcile", func() {
	It("logs no error when the Schedule has been deleted", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		var errorLogs []string
		inner := funcr.New(func(_, _ string) {}, funcr.Options{})
		r := &Reconciler{
			Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
			Log:    logr.New(&errorCountingSink{LogSink: inner.GetSink(), count: &errorLogs}),
		}

		result, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "deleted-schedule"},
		})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(ctrl.Result{}))
		Expect(errorLogs).To(BeEmpty())
	})

	It("returns other Schedule read errors for retry", func() {
		readErr := k8sError.NewServiceUnavailable("failed to read Schedule")
		r := &Reconciler{
			Client: fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
				Get: func(context.Context, client.WithWatch, client.ObjectKey, client.Object, ...client.GetOption) error {
					return readErr
				},
			}).Build(),
		}

		result, err := r.Reconcile(context.Background(), ctrl.Request{
			NamespacedName: types.NamespacedName{Namespace: "default", Name: "schedule"},
		})

		Expect(err).To(MatchError(readErr))
		Expect(result).To(Equal(ctrl.Result{}))
	})
})

// errorCountingSink records the message of every Error call.
type errorCountingSink struct {
	logr.LogSink
	count *[]string
}

func (s *errorCountingSink) Error(err error, msg string, keysAndValues ...interface{}) {
	*s.count = append(*s.count, msg)
	s.LogSink.Error(err, msg, keysAndValues...)
}
