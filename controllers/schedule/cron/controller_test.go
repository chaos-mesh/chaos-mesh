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
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestReconcileDeletedScheduleLogsNoError(t *testing.T) {
	g := NewGomegaWithT(t)

	scheme := runtime.NewScheme()
	g.Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

	var errorLogs []string
	inner := funcr.New(func(_, _ string) {}, funcr.Options{})
	r := &Reconciler{
		Client: fake.NewClientBuilder().WithScheme(scheme).Build(),
		Log:    logr.New(&errorCountingSink{LogSink: inner.GetSink(), count: &errorLogs}),
	}

	result, err := r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "default", Name: "deleted-schedule"},
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))
	g.Expect(errorLogs).To(BeEmpty())
}

// errorCountingSink records the message of every Error call.
type errorCountingSink struct {
	logr.LogSink
	count *[]string
}

func (s *errorCountingSink) Error(err error, msg string, keysAndValues ...interface{}) {
	*s.count = append(*s.count, msg)
	s.LogSink.Error(err, msg, keysAndValues...)
}
