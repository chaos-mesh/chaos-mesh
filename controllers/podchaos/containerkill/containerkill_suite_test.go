package containerkill_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	controllerkill "github.com/pingcap/chaos-mesh/controllers/podchaos/containerkill"
	. "github.com/pingcap/chaos-mesh/controllers/test"
	"github.com/pingcap/chaos-mesh/pkg/mock"
)

func TestContainerKill(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"ContainerKill Suite",
		[]Reporter{envtest.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))
	close(done)
}, 60)

var _ = AfterSuite(func() {
})

var _ = Describe("PodChaos", func() {
	Context("ContainerKill", func() {
		mock.With("MockChaosDaemonClient", &MockChaosDaemonClient{})

		req := ctrl.Request{NamespacedName: types.NamespacedName{
			Namespace: metav1.NamespaceDefault,
			Name:      "podchaos-name",
		}}
		var objs []runtime.Object

		objs, _ = GenerateNPods("p", 1, v1.PodRunning, metav1.NamespaceDefault, nil, nil, v1.ContainerStatus{
			ContainerID: "fake-container-id",
			Name:        "container-name",
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
				Selector:      v1alpha1.SelectorSpec{},
				Mode:          v1alpha1.OnePodMode,
				ContainerName: "container-name",
				Scheduler:     &v1alpha1.SchedulerSpec{Cron: "@hourly"},
			},
		}
		objs = append(objs, &podChaos)

		It("ContainerKill Action", func() {
			scheme := runtime.NewScheme()
			Expect(v1.AddToScheme(scheme)).To(Succeed())
			Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

			r := controllerkill.Reconciler{
				Client:        fake.NewFakeClientWithScheme(scheme, objs...),
				EventRecorder: &record.FakeRecorder{},
				Log:           ctrl.Log.WithName("controllers").WithName("PodChaos"),
			}

			var err error

			err = r.Apply(context.TODO(), req, r.Object())
			Expect(err).ToNot(HaveOccurred())

			mock.With("MockContainerKillError", errors.New("ContainerKillError"))

			err = r.Apply(context.TODO(), req, r.Object())

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ContainerKillError"))

			err = r.Recover(context.TODO(), req, r.Object())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
