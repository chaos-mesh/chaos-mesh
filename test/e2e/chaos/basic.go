package chaos

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restClient "k8s.io/client-go/rest"
	"k8s.io/klog"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/utils/exec"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	chaosmeshv1alpha1 "github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/test/pkg/fixture"
)

const (
	pauseImage = "gcr.io/google-containers/pause:latest"
)

var _ = ginkgo.Describe("[Basic]", func() {
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

	// pod chaos case in [PodChaos] context
	ginkgo.Context("[PodChaos]", func() {

		// podfailure chaos case in [PodFailure] context
		ginkgo.Context("[PodFailure]", func() {

			ginkgo.It("[Duration]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewCommonNginxDeployment("nginx", ns, 1)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create nginx deployment error")
				err = waitDeploymentReady("nginx", ns, kubeCli)
				framework.ExpectNoError(err, "wait nginx deployment ready error")

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
				framework.ExpectNoError(err, "create pod failure chaos error")

				err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
					listOption := metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(map[string]string{
							"app": "nginx",
						}).String(),
					}
					pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
					if err != nil {
						return false, nil
					}
					if len(pods.Items) != 1 {
						return false, nil
					}
					pod := pods.Items[0]
					for _, c := range pod.Spec.Containers {
						if c.Image == pauseImage {
							return true, nil
						}
					}
					return false, nil
				})

				err = cli.Delete(ctx, podFailureChaos)
				framework.ExpectNoError(err, "failed to delete pod failure chaos")

				klog.Infof("success to perform pod failure")
				err = wait.PollImmediate(3*time.Second, 1*time.Minute, func() (done bool, err error) {
					listOption := metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(map[string]string{
							"app": "nginx",
						}).String(),
					}
					pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
					if err != nil {
						return false, nil
					}
					if len(pods.Items) != 1 {
						return false, nil
					}
					pod := pods.Items[0]
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

			ginkgo.It("[Pause]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewCommonNginxDeployment("nginx", ns, 1)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create nginx deployment error")
				err = waitDeploymentReady("nginx", ns, kubeCli)
				framework.ExpectNoError(err, "wait nginx deployment ready error")

				var pods *corev1.PodList
				listOption := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app": "nginx",
					}).String(),
				}

				podFailureChaos := &v1alpha1.PodChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-failure",
						Namespace: ns,
					},
					Spec: v1alpha1.PodChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "nginx"},
						},
						Action:   v1alpha1.PodFailureAction,
						Mode:     v1alpha1.OnePodMode,
						Duration: pointer.StringPtr("5s"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
					},
				}
				err = cli.Create(ctx, podFailureChaos)
				framework.ExpectNoError(err, "create pod failure chaos error")

				chaosKey := types.NamespacedName{
					Namespace: ns,
					Name:      "nginx-failure",
				}

				// check whether the pod failure chaos succeeded or not
				err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
					pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
					if err != nil {
						return false, nil
					}
					pod := pods.Items[0]
					for _, c := range pod.Spec.Containers {
						if c.Image == pauseImage {
							return true, nil
						}
					}
					return false, nil
				})

				// pause experiment
				var mergePatch []byte
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": true,
					},
				})
				err = cli.Patch(ctx, podFailureChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch pause pod chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.PodChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get pod chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait for 1 minutes and no pod failure
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get nginx pods error")
				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					pod := pods.Items[0]
					for _, c := range pod.Spec.Containers {
						if c.Image == pauseImage {
							return true, nil
						}
					}
					return false, nil
				})
				framework.ExpectError(err, "wait no pod failure failed")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				// resume experiment
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": false,
					},
				})
				err = cli.Patch(ctx, podFailureChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch resume pod chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.PodChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get pod chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				// pod failure happens again
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					pod := pods.Items[0]
					for _, c := range pod.Spec.Containers {
						if c.Image == pauseImage {
							return true, nil
						}
					}
					return false, nil
				})
				framework.ExpectNoError(err, "wait pod failure failed")

				cancel()
			})
		})

		// podkill chaos case in [PodKill] context
		ginkgo.Context("[PodKill]", func() {

			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				bpod := fixture.NewCommonNginxPod("nginx", ns)
				_, err := kubeCli.CoreV1().Pods(ns).Create(bpod)
				framework.ExpectNoError(err, "create nginx pod error")
				err = waitPodRunning("nginx", ns, kubeCli)
				framework.ExpectNoError(err, "wait nginx running error")

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
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
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

			ginkgo.It("[Pause]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewCommonNginxDeployment("nginx", ns, 3)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create nginx deployment error")
				err = waitDeploymentReady("nginx", ns, kubeCli)
				framework.ExpectNoError(err, "wait nginx deployment ready error")

				var pods *corev1.PodList
				var newPods *corev1.PodList
				listOption := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app": "nginx",
					}).String(),
				}
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get nginx pods error")

				podKillChaos := &v1alpha1.PodChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-kill",
						Namespace: ns,
					},
					Spec: v1alpha1.PodChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "nginx"},
						},
						Action:   v1alpha1.PodKillAction,
						Mode:     v1alpha1.OnePodMode,
						Duration: pointer.StringPtr("5s"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
					},
				}
				err = cli.Create(ctx, podKillChaos)
				framework.ExpectNoError(err, "create pod chaos error")

				chaosKey := types.NamespacedName{
					Namespace: ns,
					Name:      "nginx-kill",
				}

				// some pod is killed as expected
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					return !fixture.HaveSameUIDs(pods.Items, newPods.Items), nil
				})
				framework.ExpectNoError(err, "wait pod killed failed")

				// pause experiment
				var mergePatch []byte
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": true,
					},
				})
				err = cli.Patch(ctx, podKillChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch pause pod chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.PodChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get pod chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait for 1 minutes and no pod is killed
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get nginx pods error")
				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					return !fixture.HaveSameUIDs(pods.Items, newPods.Items), nil
				})
				framework.ExpectError(err, "wait pod not killed failed")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				// resume experiment
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": false,
					},
				})
				err = cli.Patch(ctx, podKillChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch resume pod chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.PodChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get pod chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				// some pod is killed by resumed experiment
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					return !fixture.HaveSameUIDs(pods.Items, newPods.Items), nil
				})
				framework.ExpectNoError(err, "wait pod killed failed")

				cancel()
			})
		})

		// container kill chaos case in [ContainerKill] context
		ginkgo.Context("[ContainerKill]", func() {

			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewCommonNginxDeployment("nginx", ns, 1)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create nginx deployment error")
				err = waitDeploymentReady("nginx", ns, kubeCli)
				framework.ExpectNoError(err, "wait nginx deployment ready error")

				containerKillChaos := &v1alpha1.PodChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-container-kill",
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
						Action:        v1alpha1.ContainerKillAction,
						Mode:          v1alpha1.OnePodMode,
						ContainerName: "nginx",
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
					},
				}
				err = cli.Create(ctx, containerKillChaos)
				framework.ExpectNoError(err, "create container kill chaos error")

				err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
					listOption := metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(map[string]string{
							"app": "nginx",
						}).String(),
					}
					pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
					if err != nil {
						return false, nil
					}
					if len(pods.Items) != 1 {
						return false, nil
					}
					pod := pods.Items[0]
					for _, cs := range pod.Status.ContainerStatuses {
						if cs.Name == "nginx" && cs.Ready == false && cs.LastTerminationState.Terminated != nil {
							return true, nil
						}
					}
					return false, nil
				})

				err = cli.Delete(ctx, containerKillChaos)
				framework.ExpectNoError(err, "failed to delete container kill chaos")

				klog.Infof("success to perform container kill")
				err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
					listOption := metav1.ListOptions{
						LabelSelector: labels.SelectorFromSet(map[string]string{
							"app": "nginx",
						}).String(),
					}
					pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
					if err != nil {
						return false, nil
					}
					if len(pods.Items) != 1 {
						return false, nil
					}
					pod := pods.Items[0]
					for _, cs := range pod.Status.ContainerStatuses {
						if cs.Name == "nginx" && cs.Ready == true && cs.State.Running != nil {
							return true, nil
						}
					}
					return false, nil
				})
				framework.ExpectNoError(err, "container kill recover failed")

				cancel()
			})

			ginkgo.It("[Pause]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewCommonNginxDeployment("nginx", ns, 3)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create nginx deployment error")
				err = waitDeploymentReady("nginx", ns, kubeCli)
				framework.ExpectNoError(err, "wait nginx deployment ready error")

				var pods *corev1.PodList
				var newPods *corev1.PodList
				listOption := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app": "nginx",
					}).String(),
				}
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get nginx pods error")

				// Get the running nginx container ID
				containerID := pods.Items[0].Status.ContainerStatuses[0].ContainerID

				containerKillChaos := &v1alpha1.PodChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "nginx-container-kill",
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
						Action:        v1alpha1.ContainerKillAction,
						Mode:          v1alpha1.OnePodMode,
						ContainerName: "nginx",
						Duration:      pointer.StringPtr("5s"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
					},
				}
				err = cli.Create(ctx, containerKillChaos)
				framework.ExpectNoError(err, "create container kill chaos error")

				chaosKey := types.NamespacedName{
					Namespace: ns,
					Name:      "nginx-container-kill",
				}

				// nginx container is killed as expected
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					return containerID != newPods.Items[0].Status.ContainerStatuses[0].ContainerID, nil
				})
				framework.ExpectNoError(err, "wait container kill failed")

				// pause experiment
				var mergePatch []byte
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": true,
					},
				})
				err = cli.Patch(ctx, containerKillChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch pause pod chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.PodChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get pod chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait for 1 minutes and check whether nginx container will be killed or not
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get nginx pods error")
				containerID = pods.Items[0].Status.ContainerStatuses[0].ContainerID
				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					return containerID != newPods.Items[0].Status.ContainerStatuses[0].ContainerID, nil
				})
				framework.ExpectError(err, "wait container not killed failed")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				// resume experiment
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": false,
					},
				})
				err = cli.Patch(ctx, containerKillChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch resume pod chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.PodChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get pod chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				// nginx container is killed by resumed experiment
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get nginx pods error")
				containerID = pods.Items[0].Status.ContainerStatuses[0].ContainerID
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					newPods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
					framework.ExpectNoError(err, "get nginx pods error")
					return containerID != newPods.Items[0].Status.ContainerStatuses[0].ContainerID, nil
				})
				framework.ExpectNoError(err, "wait container killed failed")

				cancel()
			})
		})
	})

	// time chaos case in [TimeChaos] context
	ginkgo.Context("[TimeChaos]", func() {

		// time skew chaos case in [TimeSkew] context
		ginkgo.Context("[TimeSkew]", func() {

			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewTimerDeployment("timer", ns)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create timer deployment error")
				err = waitDeploymentReady("timer", ns, kubeCli)
				framework.ExpectNoError(err, "wait timer deployment ready error")

				listOption := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app": "timer",
					}).String(),
				}
				pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "failed to get timer pod")

				initTime, err := getPodTimeNS(pods.Items[0].Name, ns)
				framework.ExpectNoError(err, "failed to get pod time")

				timeChaos := &v1alpha1.TimeChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "timer-time-chaos",
						Namespace: ns,
					},
					Spec: v1alpha1.TimeChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "timer"},
						},
						Mode:       v1alpha1.OnePodMode,
						Duration:   pointer.StringPtr("5s"),
						TimeOffset: "-1h",
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
					},
				}
				err = cli.Create(ctx, timeChaos)
				framework.ExpectNoError(err, "create time chaos error")

				err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(pods.Items[0].Name, ns)
					framework.ExpectNoError(err, "failed to get pod time")
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "time chaos doesn't work as expected")

				err = cli.Delete(ctx, timeChaos)
				framework.ExpectNoError(err, "failed to delete time chaos")
				time.Sleep(10 * time.Second)

				klog.Infof("success to perform time chaos")
				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(pods.Items[0].Name, ns)
					framework.ExpectNoError(err, "failed to get pod time")
					// since there is no timechaos now, current pod time should not be earlier
					// than the init time
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait no timechaos error")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				cancel()
			})

			ginkgo.It("[Pause]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				nd := fixture.NewTimerDeployment("timer", ns)
				_, err := kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create timer deployment error")
				err = waitDeploymentReady("timer", ns, kubeCli)
				framework.ExpectNoError(err, "wait timer deployment ready error")

				listOption := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app": "timer",
					}).String(),
				}
				pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "failed to get timer pod")

				initTime, err := getPodTimeNS(pods.Items[0].Name, ns)
				framework.ExpectNoError(err, "failed to get pod time")

				timeChaos := &v1alpha1.TimeChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "timer-time-chaos",
						Namespace: ns,
					},
					Spec: v1alpha1.TimeChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "timer"},
						},
						Mode:       v1alpha1.OnePodMode,
						Duration:   pointer.StringPtr("5s"),
						TimeOffset: "-1h",
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10s",
						},
					},
				}
				err = cli.Create(ctx, timeChaos)
				framework.ExpectNoError(err, "create time chaos error")

				err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(pods.Items[0].Name, ns)
					framework.ExpectNoError(err, "failed to get pod time")
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "time chaos doesn't work as expected")

				chaosKey := types.NamespacedName{
					Namespace: ns,
					Name:      "timer-time-chaos",
				}

				// pause experiment
				var mergePatch []byte
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": true,
					},
				})
				err = cli.Patch(ctx, timeChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch pause time chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.TimeChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get time chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait for 1 minutes and check timer
				pods, err = kubeCli.CoreV1().Pods(ns).List(listOption)
				framework.ExpectNoError(err, "get timer pod error")
				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(pods.Items[0].Name, ns)
					framework.ExpectNoError(err, "failed to get pod time")
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait time chaos paused error")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				// resume experiment
				mergePatch, _ = json.Marshal(map[string]interface{}{
					"spec": map[string]interface{}{
						"paused": false,
					},
				})
				err = cli.Patch(ctx, timeChaos, client.ConstantPatch(types.MergePatchType, mergePatch))
				framework.ExpectNoError(err, "patch resume time chaos error")
				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.TimeChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get time chaos error")
					if chaos.Status.Experiment.Phase == chaosmeshv1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				// timechaos is running again, we want to check pod
				// whether time is earlier than init time,
				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(pods.Items[0].Name, ns)
					framework.ExpectNoError(err, "failed to get pod time")
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "time chaos failed")

				cancel()
			})
		})
	})

})

func waitPodRunning(name, namespace string, cli kubernetes.Interface) error {
	return wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		pod, err := cli.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		return true, nil
	})
}

func waitDeploymentReady(name, namespace string, cli kubernetes.Interface) error {
	return wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
		d, err := cli.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		if d.Status.AvailableReplicas != *d.Spec.Replicas {
			return false, nil
		}
		if d.Status.UpdatedReplicas != *d.Spec.Replicas {
			return false, nil
		}
		return true, nil
	})

}

// get pod current time in nanosecond
func getPodTimeNS(name, namespace string) (*time.Time, error) {
	args := []string{"logs", "-n", namespace, name, "--tail", "1"}
	out, err := exec.New().Command("kubectl", args...).CombinedOutput()
	if err != nil {
		klog.Fatalf("Failed to run 'kubectl %s'\nCombined output: %q\nError: %v", strings.Join(args, " "), string(out), err)
	}

	result := strings.Trim(string(out), "\n")
	t, err := time.Parse(time.RFC3339Nano, result)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
