package chaos

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/test/pkg/fixture"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"

	"github.com/onsi/ginkgo"
	"k8s.io/kubernetes/test/e2e/framework"

	// load pprof
	_ "net/http/pprof"
)

const (
	pauseImage = "gcr.io/google-containers/pause:latest"
)

var _ = ginkgo.Describe("[chaos-mesh] Basic", func() {
	f := framework.NewDefaultFramework("chaos-mesh")
	var ns string
	var fwCancel context.CancelFunc
	var kubeCli kubernetes.Interface
	var config *restClient.Config
	var cli client.Client

	ginkgo.BeforeEach(func() {
		ns = f.Namespace.Name
		_, cancel := context.WithCancel(context.Background())
		fwCancel = cancel
		kubeCli = f.ClientSet
		var err error
		config, err = framework.LoadConfig()
		framework.ExpectNoError(err, "config error")
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = chaosmeshv1alpha1.AddToScheme(scheme)
		cli, err = client.New(config, client.Options{Scheme: scheme})

	})

	ginkgo.AfterEach(func() {
		if fwCancel != nil {
			fwCancel()
		}
	})

	ginkgo.It("[PodFailure]", func() {
		ctx, cancel := context.WithCancel(context.Background())
		bpod := fixture.NewCommonNginxPod("nginx", ns)
		_, err := kubeCli.CoreV1().Pods(ns).Create(bpod)
		framework.ExpectNoError(err, "create nginx pod error")

		podFailureChaos := &v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-failure",
				Namespace: ns,
			},
			Spec: v1alpha1.PodChaosSpec{
				Selector: v1alpha1.SelectorSpec{
					Namespaces: []string{
						ns,
					},
					LabelSelectors: map[string]string{
						"app": "nginx",
					},
				},
				Action: v1alpha1.PodFailureAction,
				Mode:   v1alpha1.OnePodMode,
			},
		}
		err = cli.Create(ctx, podFailureChaos)
		framework.ExpectNoError(err, "create pod chaos error")

		err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
			pod, err := kubeCli.CoreV1().Pods(ns).Get("nginx", metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, c := range pod.Spec.Containers {
				if c.Image == pauseImage {
					return true, nil
				}
			}
			return false, nil
		})

		err = cli.Delete(ctx, podFailureChaos)
		framework.ExpectNoError(err, "failed to delete pod failure chaos")

		err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
			pod, err := kubeCli.CoreV1().Pods(ns).Get("nginx", metav1.GetOptions{})
			if err != nil {
				return false, err
			}
			for _, c := range pod.Spec.Containers {
				if c.Image == "nginx:latest" {
					return true, nil
				}
			}
			return false, nil
		})
		framework.ExpectNoError(err, "pod failure recover failed")

		cancel()
	})

	ginkgo.It("[PodKill]", func() {
		ctx, cancel := context.WithCancel(context.Background())
		bpod := fixture.NewCommonNginxPod("nginx", ns)
		_, err := kubeCli.CoreV1().Pods(ns).Create(bpod)
		framework.ExpectNoError(err, "create nginx pod error")

		podKillChaos := &v1alpha1.PodChaos{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-kill",
				Namespace: ns,
			},
			Spec: v1alpha1.PodChaosSpec{
				Selector: v1alpha1.SelectorSpec{
					Namespaces: []string{
						ns,
					},
					LabelSelectors: map[string]string{
						"app": "nginx",
					},
				},
				Action: v1alpha1.PodKillAction,
				Mode:   v1alpha1.OnePodMode,
			},
		}
		err = cli.Create(ctx, podKillChaos)
		framework.ExpectNoError(err, "create pod chaos error")

		err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
			_, err = kubeCli.CoreV1().Pods(ns).Get("nginx", metav1.GetOptions{})
			if err != nil && errors.IsNotFound(err) {
				return true, nil
			}
			return false, nil
		})
		framework.ExpectNoError(err, "Pod kill chaos perform failed")
		cancel()
	})
})
