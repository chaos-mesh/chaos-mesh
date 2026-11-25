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

package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

func TestScheduleJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schedule Job Listing")
}

var _ = Describe("Listing schedule jobs", func() {
	ctx := context.Background()
	schedule := &v1alpha1.Schedule{
		ObjectMeta: metav1.ObjectMeta{Namespace: "team", Name: "daily"},
		Spec:       v1alpha1.ScheduleSpec{Type: v1alpha1.ScheduleTypeWorkflow},
	}

	It("returns workflow jobs from the schedule's namespace and label", func() {
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		child := &v1alpha1.Workflow{ObjectMeta: metav1.ObjectMeta{
			Namespace: "team", Name: "child", Labels: map[string]string{v1alpha1.LabelManagedBy: "daily"},
		}}
		differentSchedule := child.DeepCopy()
		differentSchedule.Name = "other"
		differentSchedule.Labels[v1alpha1.LabelManagedBy] = "weekly"
		differentNamespace := child.DeepCopy()
		differentNamespace.Namespace = "other"
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(child, differentSchedule, differentNamespace).Build()

		result, err := NewActiveLister(kube, logr.Discard()).ListActiveJobs(ctx, schedule)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(BeAssignableToTypeOf(&v1alpha1.WorkflowList{}))
		Expect(result.GetItems()).To(ConsistOf(HaveField("Name", "child")))
	})

	It("returns errors for an unknown kind and for a failed list", func() {
		kube := fake.NewClientBuilder().Build()
		listFailure := errors.New("list unavailable")
		lister := NewActiveLister(interceptor.NewClient(kube, interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return listFailure
			},
		}), logr.Discard())

		unknown := schedule.DeepCopy()
		unknown.Spec.Type = "unknown"
		result, err := lister.ListActiveJobs(ctx, unknown)
		Expect(err).To(MatchError(ContainSubstring("unknown")))
		Expect(result).To(BeNil())
		result, err = lister.ListActiveJobs(ctx, schedule)
		Expect(err).To(MatchError(listFailure))
		Expect(result).To(BeNil())
	})
})
