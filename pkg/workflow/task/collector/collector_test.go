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

package collector

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestTaskCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workflow Task Collector")
}

var _ = Describe("Workflow task results", func() {
	ctx := context.Background()
	var (
		scheme *runtime.Scheme
		pod    *corev1.Pod
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		pod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "workflow", Name: "task"},
			Status: corev1.PodStatus{ContainerStatuses: []corev1.ContainerStatus{
				{Name: "sidecar", State: corev1.ContainerState{Running: &corev1.ContainerStateRunning{}}},
				{Name: "main", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 23}}},
			}},
		}
	})

	It("collects the task container's exit code and trimmed logs", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer GinkgoRecover()
			Expect(r.Method).To(Equal(http.MethodGet))
			Expect(r.URL.Path).To(Equal("/api/v1/namespaces/workflow/pods/task/log"))
			Expect(r.URL.Query().Get("container")).To(Equal("main"))
			_, _ = io.WriteString(w, "  task output\n")
		}))
		DeferCleanup(server.Close)
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
		collector := DefaultCollector(kube, &rest.Config{Host: server.URL}, pod.Namespace, pod.Name, "main")
		Expect(collector.CollectContext(ctx)).To(Equal(map[string]any{ExitCode: int32(23), Stdout: "task output"}))
	})

	It("does not return a partial result when log collection fails", func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "logs unavailable", http.StatusServiceUnavailable)
		}))
		DeferCleanup(server.Close)
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
		collector := DefaultCollector(kube, &rest.Config{Host: server.URL}, pod.Namespace, pod.Name, "main")
		env, err := collector.CollectContext(ctx)
		Expect(err).To(HaveOccurred())
		Expect(env).To(BeNil())
	})

	It("distinguishes missing pods from missing or unfinished containers", func() {
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pod).Build()
		env, err := NewExitCodeCollector(kube, pod.Namespace, "missing", "main").CollectContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(env).To(BeNil())

		for _, name := range []string{"missing", "sidecar"} {
			env, err := NewExitCodeCollector(kube, pod.Namespace, pod.Name, name).CollectContext(ctx)
			Expect(err).To(MatchError(ContainSubstring(name)))
			Expect(env).To(BeNil())
		}
	})

	It("restores stored task context in order and rejects invalid context", func() {
		status := v1alpha1.ConditionalBranchesStatus{
			Context: []string{`{"stdout":"first","exitCode":23}`, `null`, `{"stdout":"last"}`},
		}
		Expect(NewValuedCollector(status).CollectContext()).To(Equal(map[string]any{Stdout: "last", ExitCode: float64(23)}))
		status.Context = append(status.Context, `{"stdout":`)
		env, err := NewValuedCollector(status).CollectContext()
		Expect(err).To(HaveOccurred())
		Expect(env).To(BeNil())
		env, err = NewValuedCollector(v1alpha1.ConditionalBranchesStatus{}).CollectContext()
		Expect(err).NotTo(HaveOccurred())
		Expect(env).To(BeNil())
	})
})
