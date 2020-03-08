package timechaos_test

import (
	"context"
	"errors"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	v1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	. "github.com/pingcap/chaos-mesh/controllers/timechaos"
	"github.com/pingcap/chaos-mesh/pkg/mock"

	. "github.com/pingcap/chaos-mesh/controllers/test"
)

func TestTimechaos(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"TimeChaos Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ = Describe("TimeChaos", func() {
	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additional CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("TimeChaos", func() {
		podObjects, pods := GenerateNPods("p", 1, v1.PodRunning, metav1.NamespaceDefault, nil, map[string]string{"l1": "l1"})

		mock.With("MockSelectAndGeneratePods", func() []v1.Pod {
			return pods
		})
		mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})

		duration := "invalid_duration"

		timechaos := v1alpha1.TimeChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "TimeChaos",
				APIVersion: "v1",
			},
			Spec: v1alpha1.TimeChaosSpec{
				Mode:       "FixedPodMode",
				Value:      "0",
				Selector:   v1alpha1.SelectorSpec{Namespaces: []string{metav1.NamespaceDefault}},
				TimeOffset: v1alpha1.TimeOffset{},
				Duration:   &duration,
				Scheduler:  nil,
			},
		}

		It("TimeChaos Action", func() {
			scheme := runtime.NewScheme()
			Expect(v1.AddToScheme(scheme)).To(Succeed())

			r := Reconciler{
				Client:        fake.NewFakeClientWithScheme(scheme, podObjects...),
				EventRecorder: &record.FakeRecorder{},
				Log:           ctrl.Log.WithName("controllers").WithName("TimeChaos"),
			}

			var err error

			err = r.Apply(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).ToNot(HaveOccurred())

			mock.With("MockSetTimeOffsetError", errors.New("SetTimeOffsetError"))

			err = r.Apply(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("SetTimeOffsetError"))

			err = r.Recover(context.TODO(), ctrl.Request{}, &timechaos)
			Expect(err).ToNot(HaveOccurred())

			mock.With("MockSetTimeOffsetError", nil)
			err = r.Apply(context.TODO(), ctrl.Request{}, &timechaos)
			mock.With("MockRecoverTimeOffsetError", errors.New("RecoverTimeOffsetError"))

			err = r.Recover(context.TODO(), ctrl.Request{}, &timechaos)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("RecoverTimeOffsetError"))
		})
	})
})
