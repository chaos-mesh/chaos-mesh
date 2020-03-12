package podfailure_test

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/podchaos/podfailure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/pingcap/chaos-mesh/controllers/test"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func TestPodFailure(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"PodFailure Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ = Describe("PodChaos", func() {
	Context("PodFailure", func() {
		mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})

		req := ctrl.Request{NamespacedName: types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "podchaos-name",
		}}

		objs, pods := GenerateNPods("p", 1, v1.PodRunning, metav1.NamespaceDefault, nil, nil, v1.ContainerStatus{
			ContainerID: "fake-container-id",
			Name:        "container-name",
		})

		mock.With("MockSelectAndGeneratePods", func() []v1.Pod {
			return pods
		})

		podChaos := v1alpha1.PodChaos{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodChaos",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: metav1.NamespaceDefault,
				Name:      "podchaos-name",
			},
			Spec: v1alpha1.PodChaosSpec{
				Selector:      v1alpha1.SelectorSpec{Namespaces: []string{metav1.NamespaceDefault}},
				Mode:          v1alpha1.OnePodMode,
				ContainerName: "container-name",
				Scheduler:     &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			},
		}

		It("PodFailure Action", func() {
			scheme := runtime.NewScheme()
			Expect(v1.AddToScheme(scheme)).To(Succeed())
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

			r := podfailure.Reconciler{
				Client:        fake.NewFakeClientWithScheme(scheme, objs...),
				EventRecorder: &record.FakeRecorder{},
				Log:           ctrl.Log.WithName("controllers").WithName("PodChaos"),
			}

			var err error

			err = r.Apply(context.TODO(), req, &podChaos)
			Expect(err).ToNot(HaveOccurred())

			err = r.Recover(context.TODO(), req, &podChaos)
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
