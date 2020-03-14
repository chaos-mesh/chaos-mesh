package twophase_test

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/reconciler"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
)

func TestTwoPhase(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"TwoPhase Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

	Expect(addFakeToScheme(scheme.Scheme)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ twophase.InnerSchedulerObject = (*fakeTwoPhaseChaos)(nil)

type fakeTwoPhaseChaos struct {
	fakeChaos
}

var _ reconciler.InnerReconciler = (*fakeReconciler)(nil)

type fakeReconciler struct{}

func (r fakeReconciler) Apply(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	panic("implement me")
}

func (r fakeReconciler) Recover(ctx context.Context, req ctrl.Request, chaos reconciler.InnerObject) error {
	panic("implement me")
}

func (in *fakeChaos) GetStatus() *v1alpha1.ChaosStatus {
	return &in.Status
}

func (in *fakeChaos) IsDeleted() bool {
	return false
}

func (r fakeReconciler) Object() reconciler.InnerObject {
	return &fakeChaos{}
}

var _ twophase.InnerSchedulerObject = (*fakeChaos)(nil)

type fakeChaos struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            v1alpha1.ChaosStatus `json:"status,omitempty"`

	// Selector is used to select pods that are used to inject chaos action.
	Selector v1alpha1.SelectorSpec `json:"selector"`

	// Duration represents the duration of the chaos action
	Duration *string `json:"duration,omitempty"`

	// Scheduler defines some schedule rules to control the running time of the chaos experiment about time.
	Scheduler *v1alpha1.SchedulerSpec `json:"scheduler,omitempty"`

	// Next time when this action will be applied again
	// +optional
	NextStart *metav1.Time `json:"nextStart,omitempty"`

	// Next time when this action will be recovered
	// +optional
	NextRecover *metav1.Time `json:"nextRecover,omitempty"`
}

func (in *fakeChaos) GetDuration() (*time.Duration, error) {
	if in.Duration == nil {
		return nil, nil
	}
	duration, err := time.ParseDuration(*in.Duration)
	if err != nil {
		return nil, err
	}
	return &duration, nil
}

func (in *fakeChaos) GetNextStart() time.Time {
	if in.NextStart == nil {
		return time.Time{}
	}
	return in.NextStart.Time
}

func (in *fakeChaos) SetNextStart(t time.Time) {
	if t.IsZero() {
		in.NextStart = nil
		return
	}

	if in.NextStart == nil {
		in.NextStart = &metav1.Time{}
	}
	in.NextStart.Time = t
}

func (in *fakeChaos) GetNextRecover() time.Time {
	if in.NextRecover == nil {
		return time.Time{}
	}
	return in.NextRecover.Time
}

func (in *fakeChaos) SetNextRecover(t time.Time) {
	if t.IsZero() {
		in.NextRecover = nil
		return
	}

	if in.NextRecover == nil {
		in.NextRecover = &metav1.Time{}
	}
	in.NextRecover.Time = t
}

func (in *fakeChaos) GetScheduler() *v1alpha1.SchedulerSpec {
	return in.Scheduler
}

func (in *fakeChaos) DeepCopyInto(out *fakeChaos) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
}

func (in *fakeChaos) DeepCopy() *fakeChaos {
	if in == nil {
		return nil
	}
	out := new(fakeChaos)
	in.DeepCopyInto(out)
	return out
}

func (in *fakeChaos) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

var (
	schemeBuilder   = runtime.NewSchemeBuilder(addKnownTypes)
	addFakeToScheme = schemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schema.GroupVersion{Group: "", Version: "v1"},
		&fakeTwoPhaseChaos{},
		&fakeChaos{},
	)
	return nil
}

var _ = Describe("TwoPhase", func() {
	Context("TwoPhase", func() {
		past := time.Now().Add(-1)
		var _ = past

		It("TwoPhase Action", func() {
			req := ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name:      "fakechaos-name",
					Namespace: metav1.NamespaceDefault,
				},
			}

			chaos := fakeChaos{
				TypeMeta: metav1.TypeMeta{
					Kind:       "PodChaos",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: metav1.NamespaceDefault,
					Name:      "fakechaos-name",
				},
			}
			twoPhaseChaos := fakeTwoPhaseChaos{fakeChaos: chaos}

			c := fake.NewFakeClientWithScheme(scheme.Scheme, &chaos, &twoPhaseChaos)

			r := twophase.Reconciler{
				InnerReconciler: fakeReconciler{},
				Client:          c,
				Log:             ctrl.Log.WithName("controllers").WithName("TwoPhase"),
			}

			_, err := r.Reconcile(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("misdefined scheduler"))
		})
	})
})
