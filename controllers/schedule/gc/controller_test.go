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

package gc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func TestScheduleGC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Schedule Garbage Collection")
}

var _ = Describe("Schedule history cleanup", func() {
	ctx := context.Background()
	const scheduleName = "daily"
	const namespace = "team-a"
	var (
		scheme   *runtime.Scheme
		schedule *v1alpha1.Schedule
		events   *recorder.DebugRecorder
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())
		schedule = &v1alpha1.Schedule{
			ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: scheduleName},
			Spec:       v1alpha1.ScheduleSpec{Type: "PodChaos", HistoryLimit: 1},
		}
		events = recorder.NewDebugRecorder()
	})

	makePodChaos := func(ns, name string, createdAt int64, phase v1alpha1.Phase) *v1alpha1.PodChaos {
		job := &v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:         ns,
				Name:              name,
				CreationTimestamp: metav1.NewTime(time.Unix(createdAt, 0)),
				Labels:            map[string]string{v1alpha1.LabelManagedBy: scheduleName},
			},
			Spec: v1alpha1.PodChaosSpec{Action: v1alpha1.PodKillAction},
		}
		job.Status.Experiment.Records = []*v1alpha1.Record{{Id: "example", Phase: phase}}
		return job
	}

	It("deletes only old completed jobs of the schedule and requeues for unfinished jobs", func() {
		finished := makePodChaos(namespace, "old-finished", 1, v1alpha1.Injected)
		running := makePodChaos(namespace, "old-running", 2, v1alpha1.NotInjected)
		recent := makePodChaos(namespace, "latest", 3, v1alpha1.Injected)
		otherNamespace := makePodChaos("team-b", "old-finished", 0, v1alpha1.Injected)
		otherSchedule := makePodChaos(namespace, "other-schedule", 0, v1alpha1.Injected)
		otherSchedule.Labels[v1alpha1.LabelManagedBy] = "weekly"
		kube := fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(schedule, finished, running, recent, otherNamespace, otherSchedule).Build()
		log := logr.Discard()
		r := Reconciler{Client: kube, Log: log, Recorder: events, ActiveLister: utils.NewActiveLister(kube, log)}

		result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(schedule)})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.RequeueAfter).To(Equal(time.Second))
		Expect(apierrors.IsNotFound(kube.Get(ctx, client.ObjectKeyFromObject(finished), &v1alpha1.PodChaos{}))).To(BeTrue())
		for _, job := range []*v1alpha1.PodChaos{running, recent, otherNamespace, otherSchedule} {
			Expect(kube.Get(ctx, client.ObjectKeyFromObject(job), &v1alpha1.PodChaos{})).To(Succeed())
		}
	})

	It("records a failed list and leaves history intact", func() {
		job := makePodChaos(namespace, "existing", 1, v1alpha1.Injected)
		kube := fake.NewClientBuilder().WithScheme(scheme).WithObjects(schedule, job).Build()
		listFailure := errors.New("cannot list jobs")
		listClient := interceptor.NewClient(kube, interceptor.Funcs{
			List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
				return listFailure
			},
		})
		log := logr.Discard()
		r := Reconciler{Client: kube, Log: log, Recorder: events, ActiveLister: utils.NewActiveLister(listClient, log)}
		key := client.ObjectKeyFromObject(schedule)

		Expect(r.Reconcile(ctx, ctrl.Request{NamespacedName: key})).To(Equal(ctrl.Result{}))
		Expect(events.Events[key]).To(HaveLen(1))
		Expect(events.Events[key][0]).To(Equal(recorder.Failed{Activity: "list active jobs", Err: listFailure.Error()}))
		Expect(kube.Get(ctx, client.ObjectKeyFromObject(job), &v1alpha1.PodChaos{})).To(Succeed())
	})
})
