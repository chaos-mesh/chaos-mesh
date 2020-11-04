// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package chaos

import (
	"bytes"
	"context"
	"fmt"

	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/pingcap/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
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

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	e2econfig "github.com/chaos-mesh/chaos-mesh/test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/e2econst"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util/portforward"
	"github.com/chaos-mesh/chaos-mesh/test/pkg/fixture"

	// testcases
	podchaostestcases "github.com/chaos-mesh/chaos-mesh/test/e2e/chaos/podchaos"
)

var _ = ginkgo.Describe("[Basic]", func() {
	f := framework.NewDefaultFramework("chaos-mesh")
	var ns string
	var fwCancel context.CancelFunc
	var fw portforward.PortForward
	var kubeCli kubernetes.Interface
	var config *restClient.Config
	var cli client.Client
	c := http.Client{
		Timeout: 10 * time.Second,
	}

	ginkgo.BeforeEach(func() {
		ns = f.Namespace.Name
		ctx, cancel := context.WithCancel(context.Background())
		clientRawConfig, err := e2econfig.LoadClientRawConfig()
		framework.ExpectNoError(err, "failed to load raw config")
		fw, err = portforward.NewPortForwarder(ctx, e2econfig.NewSimpleRESTClientGetter(clientRawConfig))
		framework.ExpectNoError(err, "failed to create port forwarder")
		fwCancel = cancel
		kubeCli = f.ClientSet
		config, err = framework.LoadConfig()
		framework.ExpectNoError(err, "config error")
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		_ = v1alpha1.AddToScheme(scheme)
		cli, err = client.New(config, client.Options{Scheme: scheme})
		framework.ExpectNoError(err, "create client error")
	})

	ginkgo.AfterEach(func() {
		if fwCancel != nil {
			fwCancel()
		}
	})

	ginkgo.Context("[PodChaos]", func() {
		ginkgo.Context("[PodFailure]", func() {
			ginkgo.It("[Schedule]", func() {
				podchaostestcases.TestcasePodFailureOnceThenDelete(ns, kubeCli, cli)
			})
			ginkgo.It("[Pause]", func() {
				podchaostestcases.TestcasePodFailurePauseThenUnPause(ns, kubeCli, cli)
			})
		})
		ginkgo.Context("[PodKill]", func() {
			ginkgo.It("[Schedule]", func() {
				podchaostestcases.TestcasePodKillOnceThenDelete(ns, kubeCli, cli)
			})
			ginkgo.It("[Pause]", func() {
				podchaostestcases.TestcasePodKillPauseThenUnPause(ns, kubeCli, cli)
			})
		})
		ginkgo.Context("[ContainerKill]", func() {
			ginkgo.It("[Schedule]", func() {
				podchaostestcases.TestcaseContainerKillOnceThenDelete(ns, kubeCli, cli)
			})
			ginkgo.It("[Pause]", func() {
				podchaostestcases.TestcaseContainerKillPauseThenUnPause(ns, kubeCli, cli)
			})
		})
	})

	// time chaos case in [TimeChaos] context
	ginkgo.Context("[TimeChaos]", func() {

		var err error
		var port uint16
		var pfCancel context.CancelFunc

		ginkgo.JustBeforeEach(func() {
			svc := fixture.NewE2EService("timer", ns)
			_, err = kubeCli.CoreV1().Services(ns).Create(svc)
			framework.ExpectNoError(err, "create service error")
			nd := fixture.NewTimerDeployment("timer", ns)
			_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
			framework.ExpectNoError(err, "create timer deployment error")
			err = util.WaitDeploymentReady("timer", ns, kubeCli)
			framework.ExpectNoError(err, "wait timer deployment ready error")
			_, port, pfCancel, err = portforward.ForwardOnePort(fw, ns, "svc/timer", 8080)
			framework.ExpectNoError(err, "create helper port-forward failed")
		})

		ginkgo.JustAfterEach(func() {
			if pfCancel != nil {
				pfCancel()
			}
		})

		// time skew chaos case in [TimeSkew] context
		ginkgo.Context("[TimeSkew]", func() {

			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				err = waitE2EHelperReady(c, port)
				framework.ExpectNoError(err, "wait e2e helper ready error")

				initTime, err := getPodTimeNS(c, port)
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
						Duration:   pointer.StringPtr("9m"),
						TimeOffset: "-1h",
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, timeChaos)
				framework.ExpectNoError(err, "create time chaos error")

				err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(c, port)
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
					podTime, err := getPodTimeNS(c, port)
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
				err = waitE2EHelperReady(c, port)
				framework.ExpectNoError(err, "wait e2e helper ready error")

				initTime, err := getPodTimeNS(c, port)
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
						Duration:   pointer.StringPtr("9m"),
						TimeOffset: "-1h",
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, timeChaos)
				framework.ExpectNoError(err, "create time chaos error")

				err = wait.PollImmediate(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(c, port)
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
				err = util.PauseChaos(ctx, cli, timeChaos)
				framework.ExpectNoError(err, "pause chaos error")

				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.TimeChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get time chaos error")
					if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait for 1 minutes and check timer
				framework.ExpectNoError(err, "get timer pod error")
				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(c, port)
					framework.ExpectNoError(err, "failed to get pod time")
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait time chaos paused error")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				// resume experiment
				err = util.UnPauseChaos(ctx, cli, timeChaos)
				framework.ExpectNoError(err, "resume chaos error")

				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.TimeChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get time chaos error")
					if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				// timechaos is running again, we want to check pod
				// whether time is earlier than init time,
				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					podTime, err := getPodTimeNS(c, port)
					framework.ExpectNoError(err, "failed to get pod time")
					if podTime.Before(*initTime) {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "time chaos failed")

				cli.Delete(ctx, timeChaos)
				cancel()
			})
		})
	})

	// io chaos case in [IOChaos] context
	ginkgo.Context("[IOChaos]", func() {

		var (
			err      error
			port     uint16
			pfCancel context.CancelFunc
		)

		ginkgo.JustBeforeEach(func() {
			svc := fixture.NewE2EService("io", ns)
			_, err = kubeCli.CoreV1().Services(ns).Create(svc)
			framework.ExpectNoError(err, "create service error")
			nd := fixture.NewIOTestDeployment("io-test", ns)
			_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
			framework.ExpectNoError(err, "create io-test deployment error")
			err = util.WaitDeploymentReady("io-test", ns, kubeCli)
			framework.ExpectNoError(err, "wait io-test deployment ready error")
			_, port, pfCancel, err = portforward.ForwardOnePort(fw, ns, "svc/io", 8080)
			framework.ExpectNoError(err, "create helper io port port-forward failed")
		})

		ginkgo.JustAfterEach(func() {
			if pfCancel != nil {
				pfCancel()
			}
		})

		// io chaos case in [IODelay] context
		ginkgo.Context("[IODelay]", func() {

			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				err = waitE2EHelperReady(c, port)
				framework.ExpectNoError(err, "wait e2e helper ready error")

				ioChaos := &v1alpha1.IoChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "io-chaos",
						Namespace: ns,
					},
					Spec: v1alpha1.IoChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "io"},
						},
						Action:     v1alpha1.IoLatency,
						Mode:       v1alpha1.OnePodMode,
						VolumePath: "/var/run/data",
						Path:       "/var/run/data/*",
						Delay:      "1s",
						Percent:    100,
						Duration:   pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, ioChaos)
				framework.ExpectNoError(err, "create io chaos error")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					dur, _ := getPodIODelay(c, port)

					second := dur.Seconds()
					klog.Infof("get io delay %fs", second)
					// IO Delay >= 1s
					if second >= 1 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "io chaos doesn't work as expected")
				klog.Infof("apply io chaos successfully")

				err = cli.Delete(ctx, ioChaos)
				framework.ExpectNoError(err, "failed to delete io chaos")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					dur, _ := getPodIODelay(c, port)

					second := dur.Seconds()
					klog.Infof("get io delay %fs", second)
					// IO Delay shouldn't longer than 1s
					if second >= 1 {
						return false, nil
					}
					return true, nil
				})
				framework.ExpectNoError(err, "fail to recover io chaos")
				cancel()
			})

			ginkgo.It("[Pause]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				err = waitE2EHelperReady(c, port)
				framework.ExpectNoError(err, "wait e2e helper ready error")

				ioChaos := &v1alpha1.IoChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "io-chaos",
						Namespace: ns,
					},
					Spec: v1alpha1.IoChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "io"},
						},
						Action:     v1alpha1.IoLatency,
						Mode:       v1alpha1.OnePodMode,
						VolumePath: "/var/run/data",
						Path:       "/var/run/data/*",
						Delay:      "1s",
						Percent:    100,
						Duration:   pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, ioChaos)
				framework.ExpectNoError(err, "error occurs while applying io chaos")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					dur, _ := getPodIODelay(c, port)

					second := dur.Seconds()
					klog.Infof("get io delay %fs", second)
					// IO Delay >= 1s
					if second >= 1 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "io chaos doesn't work as expected")

				chaosKey := types.NamespacedName{
					Namespace: ns,
					Name:      "io-chaos",
				}

				// pause experiment
				err = util.PauseChaos(ctx, cli, ioChaos)
				framework.ExpectNoError(err, "pause chaos error")

				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.IoChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get io chaos error")
					if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait 1 min to check whether io delay still exists
				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					dur, _ := getPodIODelay(c, port)

					second := dur.Seconds()
					klog.Infof("get io delay %fs", second)
					// IO Delay shouldn't longer than 1s
					if second >= 1 {
						return false, nil
					}
					return true, nil
				})
				framework.ExpectNoError(err, "fail to recover io chaos")

				// resume experiment
				err = util.UnPauseChaos(ctx, cli, ioChaos)
				framework.ExpectNoError(err, "resume chaos error")

				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.IoChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get io chaos error")
					if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					dur, _ := getPodIODelay(c, port)

					second := dur.Seconds()
					klog.Infof("get io delay %fs", second)
					// IO Delay >= 1s
					if second >= 1 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "io chaos doesn't work as expected")

				// cleanup
				cli.Delete(ctx, ioChaos)
				cancel()
			})
		})

		// io chaos case in [IOError] context
		ginkgo.Context("[IOErrno]", func() {

			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				err = waitE2EHelperReady(c, port)
				framework.ExpectNoError(err, "wait e2e helper ready error")

				ioChaos := &v1alpha1.IoChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "io-chaos",
						Namespace: ns,
					},
					Spec: v1alpha1.IoChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "io"},
						},
						Action:     v1alpha1.IoFaults,
						Mode:       v1alpha1.OnePodMode,
						VolumePath: "/var/run/data",
						Path:       "/var/run/data/*",
						Percent:    100,
						// errno 5 is EIO -> I/O error
						Errno: 5,
						// only inject write method
						Methods:  []v1alpha1.IoMethod{v1alpha1.Write},
						Duration: pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, ioChaos)
				framework.ExpectNoError(err, "create io chaos")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					_, err = getPodIODelay(c, port)
					// input/output error is errno 5
					if err != nil && strings.Contains(err.Error(), "input/output error") {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "io chaos doesn't work as expected")

				err = cli.Delete(ctx, ioChaos)
				framework.ExpectNoError(err, "failed to delete io chaos")

				klog.Infof("success to perform io chaos")
				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					_, err = getPodIODelay(c, port)

					if err == nil {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "fail to recover io chaos")

				cancel()
			})

			ginkgo.It("[Pause]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				err = waitE2EHelperReady(c, port)
				framework.ExpectNoError(err, "wait e2e helper ready error")

				ioChaos := &v1alpha1.IoChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "io-chaos",
						Namespace: ns,
					},
					Spec: v1alpha1.IoChaosSpec{
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "io"},
						},
						Action:     v1alpha1.IoFaults,
						Mode:       v1alpha1.OnePodMode,
						VolumePath: "/var/run/data",
						Path:       "/var/run/data/*",
						Percent:    100,
						// errno 5 is EIO -> I/O error
						Errno: 5,
						// only inject write method
						Methods:  []v1alpha1.IoMethod{v1alpha1.Write},
						Duration: pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, ioChaos)
				framework.ExpectNoError(err, "create io chaos error")

				klog.Info("create iochaos successfully")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					_, err = getPodIODelay(c, port)
					// input/output error is errno 5
					if err != nil && strings.Contains(err.Error(), "input/output error") {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "io chaos doesn't work as expected")

				chaosKey := types.NamespacedName{
					Namespace: ns,
					Name:      "io-chaos",
				}

				// pause experiment
				err = util.PauseChaos(ctx, cli, ioChaos)
				framework.ExpectNoError(err, "pause chaos error")

				klog.Info("pause iochaos")

				err = wait.Poll(5*time.Second, 5*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.IoChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get io chaos error")
					if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhasePaused {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check paused chaos failed")

				// wait 1 min to check whether io delay still exists
				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					_, err = getPodIODelay(c, port)

					if err == nil {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "fail to recover io chaos")

				// resume experiment
				err = util.UnPauseChaos(ctx, cli, ioChaos)
				framework.ExpectNoError(err, "resume chaos error")

				err = wait.Poll(5*time.Second, 1*time.Minute, func() (done bool, err error) {
					chaos := &v1alpha1.IoChaos{}
					err = cli.Get(ctx, chaosKey, chaos)
					framework.ExpectNoError(err, "get io chaos error")
					if chaos.Status.Experiment.Phase == v1alpha1.ExperimentPhaseRunning {
						return true, nil
					}
					return false, err
				})
				framework.ExpectNoError(err, "check resumed chaos failed")

				err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
					_, err = getPodIODelay(c, port)
					// input/output error is errno 5
					if err != nil && strings.Contains(err.Error(), "input/output error") {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectNoError(err, "io chaos doesn't work as expected")

				// cleanup
				cli.Delete(ctx, ioChaos)
				cancel()
			})
		})
	})

	ginkgo.Context("[Sidecar Config]", func() {
		var (
			cmName      string
			cmNamespace string
		)

		// delete the created config map in each test case
		ginkgo.JustAfterEach(func() {
			kubeCli.CoreV1().ConfigMaps(cmNamespace).Delete(cmName, &metav1.DeleteOptions{})
		})

		ginkgo.Context("[Template Config]", func() {

			ginkgo.It("[InValid ConfigMap key]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cmName = "incorrect-key-name"
				cmNamespace = e2econst.ChaosMeshNamespace
				err := createTemplateConfig(ctx, cli, cmName,
					map[string]string{
						"chaos-pd.yaml": `name: chaosfs-pd
selector:
  labelSelectors:
    "app.kubernetes.io/component": "pd"`})
				framework.ExpectNoError(err, "failed to create template config")

				listOptions := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app.kubernetes.io/component": "controller-manager",
					}).String(),
				}
				pods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
				framework.ExpectNoError(err, "get chaos mesh controller pods error")

				err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
					newPods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
					framework.ExpectNoError(err, "get chaos mesh controller pods error")
					if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
						return true, nil
					}
					if newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait chaos mesh not dies")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				cancel()
			})

			ginkgo.It("[InValid Configuration]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cmName = "incorrect-configuration"
				cmNamespace = e2econst.ChaosMeshNamespace
				err := createTemplateConfig(ctx, cli, cmName,
					map[string]string{
						"data": `name: chaosfs-pd
selector:
  labelSelectors:
    "app.kubernetes.io/component": "pd"`})
				framework.ExpectNoError(err, "failed to create template config")

				listOptions := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app.kubernetes.io/component": "controller-manager",
					}).String(),
				}
				pods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
				framework.ExpectNoError(err, "get chaos mesh controller pods error")

				err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
					newPods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
					framework.ExpectNoError(err, "get chaos mesh controller pods error")
					if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
						return true, nil
					}
					if newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait chaos mesh not dies")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				cancel()
			})
		})

		ginkgo.Context("[Injection Config]", func() {
			ginkgo.It("[No Template]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cmName = "no-template-name"
				cmNamespace = ns
				err := createInjectionConfig(ctx, cli, ns, cmName,
					map[string]string{
						"chaosfs-io": `name: chaosfs-io
selector:
  labelSelectors:
    app: io`})
				framework.ExpectNoError(err, "failed to create injection config")

				listOptions := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app.kubernetes.io/component": "controller-manager",
					}).String(),
				}
				pods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
				framework.ExpectNoError(err, "get chaos mesh controller pods error")

				err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
					newPods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
					framework.ExpectNoError(err, "get chaos mesh controller pods error")
					if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
						return true, nil
					}
					if newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait chaos mesh not dies")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				err = enableWebhook(ns)
				framework.ExpectNoError(err, "enable webhook on ns error")
				nd := fixture.NewIOTestDeployment("io-test", ns)
				_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create io-test deployment error")
				err = util.WaitDeploymentReady("io-test", ns, kubeCli)
				framework.ExpectNoError(err, "wait io-test deployment ready error")

				cancel()
			})

			ginkgo.It("[No Template Args]", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cmName = "no-template-args"
				cmNamespace = ns
				err := createInjectionConfig(ctx, cli, ns, cmName,
					map[string]string{
						"chaosfs-io": `name: chaosfs-io
template: chaosfs-sidecar
selector:
  labelSelectors:
    app: io`})
				framework.ExpectNoError(err, "failed to create injection config")

				listOptions := metav1.ListOptions{
					LabelSelector: labels.SelectorFromSet(map[string]string{
						"app.kubernetes.io/component": "controller-manager",
					}).String(),
				}
				pods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
				framework.ExpectNoError(err, "get chaos mesh controller pods error")

				err = wait.Poll(time.Second, 10*time.Second, func() (done bool, err error) {
					newPods, err := kubeCli.CoreV1().Pods(e2econst.ChaosMeshNamespace).List(listOptions)
					framework.ExpectNoError(err, "get chaos mesh controller pods error")
					if !fixture.HaveSameUIDs(pods.Items, newPods.Items) {
						return true, nil
					}
					if newPods.Items[0].Status.ContainerStatuses[0].RestartCount > 0 {
						return true, nil
					}
					return false, nil
				})
				framework.ExpectError(err, "wait chaos mesh not dies")
				framework.ExpectEqual(err.Error(), wait.ErrWaitTimeout.Error())

				err = enableWebhook(ns)
				framework.ExpectNoError(err, "enable webhook on ns error")
				nd := fixture.NewIOTestDeployment("io-test", ns)
				_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create io-test deployment error")
				err = util.WaitDeploymentReady("io-test", ns, kubeCli)
				framework.ExpectNoError(err, "wait io-test deployment ready error")

				cancel()
			})
		})
	})

	ginkgo.Context("[NetworkChaos]", func() {
		var err error

		var networkPeers []*v1.Pod
		var ports []uint16
		var pfCancels []context.CancelFunc

		ginkgo.JustBeforeEach(func() {
			ports = []uint16{}
			networkPeers = []*v1.Pod{}
			for index := 0; index < 4; index++ {
				name := fmt.Sprintf("network-peer-%d", index)

				svc := fixture.NewE2EService(name, ns)
				_, err = kubeCli.CoreV1().Services(ns).Create(svc)
				framework.ExpectNoError(err, "create service error")
				nd := fixture.NewNetworkTestDeployment(name, ns, map[string]string{"partition": strconv.Itoa(index % 2)})
				_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create network-peer deployment error")
				err = util.WaitDeploymentReady(name, ns, kubeCli)
				framework.ExpectNoError(err, "wait network-peer deployment ready error")

				pod, err := getPod(kubeCli, ns, name)
				framework.ExpectNoError(err, "select network-peer pod error")
				networkPeers = append(networkPeers, pod)

				_, port, pfCancel, err := portforward.ForwardOnePort(fw, ns, "svc/"+svc.Name, 8080)
				ports = append(ports, port)
				pfCancels = append(pfCancels, pfCancel)
				framework.ExpectNoError(err, "create helper io port port-forward failed")
			}
		})

		ginkgo.Context("[ForbidHostNetwork]", func() {
			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())

				name := "network-peer-4"
				nd := fixture.NewNetworkTestDeployment(name, ns, map[string]string{"partition": "0"})
				nd.Spec.Template.Spec.HostNetwork = true
				_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create network-peer deployment error")
				err = util.WaitDeploymentReady(name, ns, kubeCli)
				framework.ExpectNoError(err, "wait network-peer deployment ready error")

				networkPartition := &v1alpha1.NetworkChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "network-chaos-1",
						Namespace: ns,
					},
					Spec: v1alpha1.NetworkChaosSpec{
						Action: v1alpha1.PartitionAction,
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "network-peer-4"},
						},
						Mode:      v1alpha1.OnePodMode,
						Direction: v1alpha1.To,
						Target: &v1alpha1.Target{
							TargetSelector: v1alpha1.SelectorSpec{
								Namespaces:     []string{ns},
								LabelSelectors: map[string]string{"app": "network-peer-1"},
							},
							TargetMode: v1alpha1.OnePodMode,
						},
						Duration: pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}

				err = cli.Create(ctx, networkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)

				cli.Get(ctx, types.NamespacedName{
					Namespace: ns,
					Name:      "network-chaos-1",
				}, networkPartition)
				framework.ExpectEqual(networkPartition.Status.ChaosStatus.Experiment.Phase, v1alpha1.ExperimentPhaseFailed)
				framework.ExpectEqual(strings.Contains(networkPartition.Status.ChaosStatus.FailedMessage, "it's dangerous to inject network chaos on a pod"), true)

				cancel()
			})
		})

		ginkgo.Context("[NetworkPartition]", func() {
			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())

				for index := range networkPeers {
					err = waitE2EHelperReady(c, ports[index])

					framework.ExpectNoError(err, "wait e2e helper ready error")
				}
				connect := func(source, target int) bool {
					err := sendUDPPacket(c, ports[source], networkPeers[target].Status.PodIP)
					if err != nil {
						klog.Infof("Error: %v", err)
						return false
					}

					data, err := recvUDPPacket(c, ports[target])
					if err != nil || data != "ping\n" {
						klog.Infof("Error: %v, Data: %s", err, data)
						return false
					}

					return true
				}
				allBlockedConnection := func() [][]int {
					var result [][]int
					for source := range networkPeers {
						for target := range networkPeers {
							if source == target {
								continue
							}

							if !connect(source, target) {
								result = append(result, []int{source, target})
							}
						}
					}

					return result
				}
				framework.ExpectEqual(len(allBlockedConnection()), 0)

				baseNetworkPartition := &v1alpha1.NetworkChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "network-chaos-1",
						Namespace: ns,
					},
					Spec: v1alpha1.NetworkChaosSpec{
						Action: v1alpha1.PartitionAction,
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "network-peer-0"},
						},
						Mode:      v1alpha1.OnePodMode,
						Direction: v1alpha1.To,
						Target: &v1alpha1.Target{
							TargetSelector: v1alpha1.SelectorSpec{
								Namespaces:     []string{ns},
								LabelSelectors: map[string]string{"app": "network-peer-1"},
							},
							TargetMode: v1alpha1.OnePodMode,
						},
						Duration: pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}})

				err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allBlockedConnection()), 0)

				baseNetworkPartition.Spec.Direction = v1alpha1.Both
				err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}, {1, 0}})

				err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allBlockedConnection()), 0)

				baseNetworkPartition.Spec.Direction = v1alpha1.From
				err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allBlockedConnection(), [][]int{{1, 0}})

				err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allBlockedConnection()), 0)

				baseNetworkPartition.Spec.Direction = v1alpha1.Both
				baseNetworkPartition.Spec.Target.TargetSelector.LabelSelectors = map[string]string{"partition": "1"}
				baseNetworkPartition.Spec.Target.TargetMode = v1alpha1.AllPodMode
				err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}, {0, 3}, {1, 0}, {3, 0}})

				err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allBlockedConnection()), 0)

				// Multiple network partition chaos on peer-0
				anotherNetworkPartition := baseNetworkPartition.DeepCopy()
				anotherNetworkPartition.Name = "network-chaos-2"
				anotherNetworkPartition.Spec.Direction = v1alpha1.To
				anotherNetworkPartition.Spec.Target.TargetSelector.LabelSelectors = map[string]string{"partition": "0"}
				anotherNetworkPartition.Spec.Target.TargetMode = v1alpha1.AllPodMode
				err = cli.Create(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				err = cli.Create(ctx, anotherNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allBlockedConnection(), [][]int{{0, 1}, {0, 2}, {0, 3}, {1, 0}, {3, 0}})

				err = cli.Delete(ctx, baseNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				err = cli.Delete(ctx, anotherNetworkPartition.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allBlockedConnection()), 0)

				cancel()
			})
		})

		ginkgo.Context("[Netem]", func() {
			ginkgo.It("[Schedule]", func() {
				ctx, cancel := context.WithCancel(context.Background())

				for index := range networkPeers {
					err = waitE2EHelperReady(c, ports[index])

					framework.ExpectNoError(err, "wait e2e helper ready error")
				}

				testDelay := func(from int, to int) int64 {
					delay, err := testNetworkDelay(c, ports[from], networkPeers[to].Status.PodIP)
					framework.ExpectNoError(err, "send request to test delay failed")

					return delay
				}
				allSlowConnection := func() [][]int {
					var result [][]int
					for source := 0; source < len(networkPeers); source++ {
						for target := source + 1; target < len(networkPeers); target++ {
							delay := testDelay(source, target)
							klog.Infof("delay from %d to %d: %d", source, target, delay)
							if delay > 100*1e6 {
								result = append(result, []int{source, target})
							}
						}
					}

					return result
				}

				framework.ExpectEqual(len(allSlowConnection()), 0)

				// normal delay chaos
				networkDelay := &v1alpha1.NetworkChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "network-chaos-1",
						Namespace: ns,
					},
					Spec: v1alpha1.NetworkChaosSpec{
						Action: v1alpha1.DelayAction,
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "network-peer-0"},
						},
						Mode: v1alpha1.OnePodMode,
						TcParameter: v1alpha1.TcParameter{
							Delay: &v1alpha1.DelaySpec{
								Latency:     "200ms",
								Correlation: "25",
								Jitter:      "0ms",
							},
						},
						Duration: pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				klog.Infof("Injecting delay for 0")
				err = cli.Create(ctx, networkDelay.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}, {0, 2}, {0, 3}})

				err = cli.Delete(ctx, networkDelay.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allSlowConnection()), 0)

				networkDelay.Spec.Target = &v1alpha1.Target{
					TargetSelector: v1alpha1.SelectorSpec{
						Namespaces:     []string{ns},
						LabelSelectors: map[string]string{"app": "network-peer-1"},
					},
					TargetMode: v1alpha1.OnePodMode,
				}
				klog.Infof("Injecting delay for 0 -> 1")
				err = cli.Create(ctx, networkDelay.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}})

				err = cli.Delete(ctx, networkDelay.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allSlowConnection()), 0)

				evenNetworkDelay := networkDelay.DeepCopy()
				evenNetworkDelay.Name = "network-chaos-2"
				evenNetworkDelay.Spec.Target.TargetSelector.LabelSelectors = map[string]string{"partition": "0"}
				evenNetworkDelay.Spec.Target.TargetMode = v1alpha1.AllPodMode
				klog.Infof("Injecting delay for 0 -> even partition")
				err = cli.Create(ctx, evenNetworkDelay.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allSlowConnection(), [][]int{{0, 2}})

				klog.Infof("Injecting delay for 0 -> 1")
				err = cli.Create(ctx, networkDelay.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}, {0, 2}})

				err = cli.Delete(ctx, networkDelay.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allSlowConnection(), [][]int{{0, 2}})
				err = cli.Delete(ctx, evenNetworkDelay.DeepCopy())
				framework.ExpectNoError(err, "delete network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(len(allSlowConnection()), 0)

				complicateNetem := &v1alpha1.NetworkChaos{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "network-chaos-3",
						Namespace: ns,
					},
					Spec: v1alpha1.NetworkChaosSpec{
						Action: v1alpha1.DelayAction,
						Selector: v1alpha1.SelectorSpec{
							Namespaces:     []string{ns},
							LabelSelectors: map[string]string{"app": "network-peer-0"},
						},
						Mode: v1alpha1.OnePodMode,
						TcParameter: v1alpha1.TcParameter{
							Delay: &v1alpha1.DelaySpec{
								Latency:     "200ms",
								Correlation: "25",
								Jitter:      "0ms",
							},
							Loss: &v1alpha1.LossSpec{
								Loss:        "25",
								Correlation: "25",
							},
							Duplicate: &v1alpha1.DuplicateSpec{
								Duplicate:   "25",
								Correlation: "25",
							},
							Corrupt: &v1alpha1.CorruptSpec{
								Corrupt:     "25",
								Correlation: "25",
							},
						},
						Duration: pointer.StringPtr("9m"),
						Scheduler: &v1alpha1.SchedulerSpec{
							Cron: "@every 10m",
						},
					},
				}
				klog.Infof("Injecting delay for 0")
				err = cli.Create(ctx, complicateNetem.DeepCopy())
				framework.ExpectNoError(err, "create network chaos error")
				time.Sleep(5 * time.Second)
				framework.ExpectEqual(allSlowConnection(), [][]int{{0, 1}, {0, 2}, {0, 3}})

				cancel()
			})
		})

		ginkgo.JustAfterEach(func() {
			for _, cancel := range pfCancels {
				cancel()
			}
		})
	})

})

func waitE2EHelperReady(c http.Client, port uint16) error {
	return wait.Poll(10*time.Second, 5*time.Minute, func() (done bool, err error) {
		if _, err = c.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err != nil {
			return false, nil
		}
		return true, nil
	})
}

// get pod current time in nanosecond
func getPodTimeNS(c http.Client, port uint16) (*time.Time, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/time", port))
	if err != nil {
		return nil, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339Nano, string(out))
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// get pod io delay
func getPodIODelay(c http.Client, port uint16) (time.Duration, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/io", port))
	if err != nil {
		return 0, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	result := string(out)
	if strings.Contains(result, "failed to write file") {
		return 0, errors.New(result)
	}
	dur, err := time.ParseDuration(result)
	if err != nil {
		return 0, err
	}

	return dur, nil
}

func testNetworkDelay(c http.Client, port uint16, targetIP string) (int64, error) {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	klog.Infof("sending request to localhost:%d with body: %s", port, string(body))

	resp, err := c.Post(fmt.Sprintf("http://localhost:%d/network/ping", port), "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	result := string(out)
	parts := strings.Split(result, " ")
	if len(parts) != 2 {
		return 0, fmt.Errorf("the length of parts is not 2 %v", parts)
	}

	if parts[0] != "OK" {
		return 0, fmt.Errorf("the first part of response is not OK")
	}

	return strconv.ParseInt(parts[1], 10, 64)
}

func recvUDPPacket(c http.Client, port uint16) (string, error) {
	klog.Infof("sending request to http://localhost:%d/network/recv", port)
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/network/recv", port))
	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	result := string(out)
	return result, nil
}

func sendUDPPacket(c http.Client, port uint16, targetIP string) error {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	klog.Infof("sending request to http://localhost:%d/network/send with body: %s", port, string(body))

	resp, err := c.Post(fmt.Sprintf("http://localhost:%d/network/send", port), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	result := string(out)
	if result != "send successfully\n" {
		return fmt.Errorf("doesn't send successfully")
	}

	klog.Info("send request successfully")
	return nil
}

func getPod(kubeCli kubernetes.Interface, ns string, appLabel string) (*v1.Pod, error) {
	listOption := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app": appLabel,
		}).String(),
	}

	pods, err := kubeCli.CoreV1().Pods(ns).List(listOption)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) > 1 {
		return nil, fmt.Errorf("select more than one pod")
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("cannot select any pod")
	}

	return &pods.Items[0], nil
}

// enableWebhook enables webhook on the specific namespace
func enableWebhook(ns string) error {
	args := []string{"label", "ns", ns, "admission-webhook=enabled"}
	out, err := exec.New().Command("kubectl", args...).CombinedOutput()
	if err != nil {
		klog.Fatalf("Failed to run 'kubectl %s'\nCombined output: %q\nError: %v", strings.Join(args, " "), string(out), err)
	}
	return nil
}

func createTemplateConfig(
	ctx context.Context,
	cli client.Client,
	name string,
	data map[string]string,
) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: e2econst.ChaosMeshNamespace,
			Name:      name,
			Labels: map[string]string{
				"app.kubernetes.io/component": "template",
			},
		},
		Data: data,
	}
	return cli.Create(ctx, cm)
}

func createInjectionConfig(
	ctx context.Context,
	cli client.Client,
	ns, name string,
	data map[string]string,
) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
			Labels: map[string]string{
				"app.kubernetes.io/component": "webhook",
			},
		},
		Data: data,
	}
	return cli.Create(ctx, cm)
}
