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

package pause

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestSchedulePause(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schedule Pause")
}

var _ = Describe("Schedule pause propagation", func() {
	ctx := context.Background()
	var (
		schedule *v1alpha1.Schedule
		scheme   *runtime.Scheme
		events   *recorder.DebugRecorder
		log      = logr.Discard()
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		schedule = &v1alpha1.Schedule{
			ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "daily",
				Annotations: map[string]string{v1alpha1.PauseAnnotationKey: "true"}},
			Spec: v1alpha1.ScheduleSpec{Type: "PodChaos"},
		}
		events = recorder.NewDebugRecorder()
	})

	It("refetches after a conflict, preserves the concurrent update, and resumes a paused child", func() {
		job := &v1alpha1.PodChaos{ObjectMeta: metav1.ObjectMeta{
			Namespace: "test", Name: "child",
			Labels: map[string]string{v1alpha1.LabelManagedBy: schedule.Name},
		}}
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule, job).Build()
		attempts := 0
		runtimeClient := interceptor.NewClient(kube, interceptor.Funcs{
			Update: func(ctx context.Context, c client.WithWatch, obj client.Object, options ...client.UpdateOption) error {
				attempts++
				if attempts != 1 {
					return c.Update(ctx, obj, options...)
				}
				current := &v1alpha1.PodChaos{}
				Expect(c.Get(ctx, client.ObjectKeyFromObject(obj), current)).To(Succeed())
				current.Annotations = map[string]string{"owner": "team"}
				Expect(c.Update(ctx, current)).To(Succeed())
				return apierrors.NewConflict(schema.GroupResource{Group: "chaos-mesh.org", Resource: "podchaos"},
					obj.GetName(), errors.New("concurrent update"))
			},
		})
		r := Reconciler{Client: runtimeClient, Log: log, Recorder: events, ActiveLister: utils.NewActiveLister(kube, log)}
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}

		Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		Expect(attempts).To(Equal(2))
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
		Expect(job.Annotations).To(Equal(map[string]string{"owner": "team", v1alpha1.PauseAnnotationKey: "true"}))
		Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		Expect(attempts).To(Equal(2))

		Expect(kube.Get(ctx, req.NamespacedName, schedule)).To(Succeed())
		schedule.Annotations[v1alpha1.PauseAnnotationKey] = "false"
		Expect(kube.Update(ctx, schedule)).To(Succeed())
		Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		Expect(attempts).To(Equal(3))
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(job), job)).To(Succeed())
		Expect(job.Annotations).To(Equal(map[string]string{"owner": "team", v1alpha1.PauseAnnotationKey: "false"}))
	})

	It("records that workflow scheduling cannot be paused without listing children", func() {
		schedule.Spec.Type = v1alpha1.ScheduleTypeWorkflow
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule).Build()
		r := Reconciler{Client: kube, Log: log, Recorder: events}
		req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)}

		Expect(r.Reconcile(ctx, req)).To(Equal(ctrl.Result{}))
		Expect(events.Events[req.NamespacedName]).To(Equal([]recorder.ChaosEvent{
			recorder.NotSupported{Activity: "pausing a workflow schedule"},
		}))
	})
})
