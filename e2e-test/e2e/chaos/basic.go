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
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/onsi/ginkgo"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restClient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/test/e2e/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	e2econfig "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/e2econst"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
	"github.com/chaos-mesh/chaos-mesh/pkg/portforward"

	// testcases
	dnschaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/dnschaos"
	httpchaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/httpchaos"
	iochaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/iochaos"
	networkchaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/networkchaos"
	podchaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/podchaos"
	sidecartestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/sidecar"
	stresstestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/stresschaos"
	timechaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/timechaos"
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
		fw, err = portforward.NewPortForwarder(ctx, e2econfig.NewSimpleRESTClientGetter(clientRawConfig), true)
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
				timechaostestcases.TestcaseTimeSkewOnceThenRecover(ns, cli, c, port)
			})

			ginkgo.It("[Pause]", func() {
				timechaostestcases.TestcaseTimeSkewPauseThenUnpause(ns, cli, c, port)
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
				iochaostestcases.TestcaseIODelayDurationForATimeThenRecover(ns, cli, c, port)
			})

			ginkgo.It("[Pause]", func() {
				iochaostestcases.TestcaseIODelayDurationForATimePauseAndUnPause(ns, cli, c, port)
			})
			ginkgo.It("[SpecifyContainer]", func() {
				iochaostestcases.TestcaseIODelayWithSpecifiedContainer(ns, cli, c, port)
			})
			ginkgo.It("[WrongSpec]", func() {
				iochaostestcases.TestcaseIODelayWithWrongSpec(ns, cli, c, port)
			})
		})

		// io chaos case in [IOError] context
		ginkgo.Context("[IOErrno]", func() {

			ginkgo.It("[Schedule]", func() {
				iochaostestcases.TestcaseIOErrorDurationForATimeThenRecover(ns, cli, c, port)
			})
			ginkgo.It("[Pause]", func() {
				iochaostestcases.TestcaseIOErrorDurationForATimePauseAndUnPause(ns, cli, c, port)
			})
			ginkgo.It("[SpecifyContainer]", func() {
				iochaostestcases.TestcaseIOErrorWithSpecifiedContainer(ns, cli, c, port)
			})
		})

		// io mistake case in [IOMistake] context
		ginkgo.Context("[IOMistake]", func() {

			ginkgo.It("[Schedule]", func() {
				iochaostestcases.TestcaseIOMistakeDurationForATimeThenRecover(ns, cli, c, port)
			})
			ginkgo.It("[Pause]", func() {
				iochaostestcases.TestcaseIOMistakeDurationForATimePauseAndUnPause(ns, cli, c, port)
			})
			ginkgo.It("[SpecifyContainer]", func() {
				iochaostestcases.TestcaseIOMistakeWithSpecifiedContainer(ns, cli, c, port)
			})
		})
	})

	// http chaos case in [HTTPChaos] context
	ginkgo.Context("[HTTPChaos]", func() {

		var (
			err      error
			port     uint16
			pfCancel context.CancelFunc
		)

		ginkgo.JustBeforeEach(func() {
			svc := fixture.NewE2EService("http", ns)
			_, err = kubeCli.CoreV1().Services(ns).Create(svc)
			framework.ExpectNoError(err, "create service error")
			nd := fixture.NewHTTPTestDeployment("http-test", ns)
			_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
			framework.ExpectNoError(err, "create http-test deployment error")
			err = util.WaitDeploymentReady("http-test", ns, kubeCli)
			framework.ExpectNoError(err, "wait http-test deployment ready error")
			_, port, pfCancel, err = portforward.ForwardOnePort(fw, ns, "svc/http", 8080)
			framework.ExpectNoError(err, "create helper io port port-forward failed")
		})

		ginkgo.JustAfterEach(func() {
			if pfCancel != nil {
				pfCancel()
			}
		})

		// http chaos case in [HTTPDelay] context
		ginkgo.Context("[HTTPDelay]", func() {
			ginkgo.It("[Schedule]", func() {
				httpchaostestcases.TestcaseHttpDelayDurationForATimeThenRecover(ns, cli, c, port)
			})
			ginkgo.It("[Pause]", func() {
				httpchaostestcases.TestcaseHttpDelayDurationForATimePauseAndUnPause(ns, cli, c, port)
			})
		})

		// http chaos case in [HTTPAbort] context
		ginkgo.Context("[HTTPAbort]", func() {
			ginkgo.It("[Schedule]", func() {
				httpchaostestcases.TestcaseHttpAbortThenRecover(ns, cli, c, port)
			})
			ginkgo.It("[Pause]", func() {
				httpchaostestcases.TestcaseHttpAbortPauseAndUnPause(ns, cli, c, port)
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
				cmName = "incorrect-key-name"
				cmNamespace = e2econst.ChaosMeshNamespace
				sidecartestcases.TestcaseInvalidConfigMapKey(ns, cmNamespace, cmName, kubeCli, cli)
			})

			ginkgo.It("[InValid Configuration]", func() {
				cmName = "incorrect-configuration"
				cmNamespace = e2econst.ChaosMeshNamespace
				sidecartestcases.TestcaseInvalidConfiguration(ns, cmNamespace, cmName, kubeCli, cli)
			})
		})

		ginkgo.Context("[Injection Config]", func() {
			ginkgo.It("[No Template]", func() {
				cmName = "no-template-name"
				cmNamespace = e2econst.ChaosMeshNamespace
				sidecartestcases.TestcaseNoTemplate(ns, cmNamespace, cmName, kubeCli, cli)
			})

			ginkgo.It("[No Template Args]", func() {
				cmName = "no-template-args"
				cmNamespace = e2econst.ChaosMeshNamespace
				sidecartestcases.TestcaseNoTemplateArgs(ns, cmNamespace, cmName, kubeCli, cli)
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
				networkchaostestcases.TestcaseForbidHostNetwork(ns, kubeCli, cli)
			})
		})

		ginkgo.Context("[NetworkPartition]", func() {
			ginkgo.It("[Schedule]", func() {
				networkchaostestcases.TestcaseNetworkPartition(ns, cli, networkPeers, ports, c)
			})
		})

		ginkgo.Context("[Netem]", func() {
			ginkgo.It("[Schedule]", func() {
				networkchaostestcases.TestcaseNetworkDelay(ns, cli, networkPeers, ports, c)
			})
			ginkgo.It("[PeersCrossoverWithDirectionBoth]", func() {
				networkchaostestcases.TestcasePeersCrossover(ns, cli, networkPeers, ports, c)
			})
		})

		ginkgo.JustAfterEach(func() {
			for _, cancel := range pfCancels {
				cancel()
			}
		})
	})
	// DNS chaos case in [DNSChaos] context
	ginkgo.Context("[DNSChaos]", func() {
		var err error
		var port uint16

		ginkgo.JustBeforeEach(func() {
			name := "network-peer"

			svc := fixture.NewE2EService(name, ns)
			_, err = kubeCli.CoreV1().Services(ns).Create(svc)
			framework.ExpectNoError(err, "create service error")
			nd := fixture.NewNetworkTestDeployment(name, ns, map[string]string{"partition": "0"})
			_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
			framework.ExpectNoError(err, "create network-peer deployment error")
			err = util.WaitDeploymentReady(name, ns, kubeCli)
			framework.ExpectNoError(err, "wait network-peer deployment ready error")

			_, err = getPod(kubeCli, ns, name)
			framework.ExpectNoError(err, "select network-peer pod error")

			_, port, _, err = portforward.ForwardOnePort(fw, ns, "svc/"+svc.Name, 8080)
			framework.ExpectNoError(err, "create helper io port port-forward failed")
		})
		ginkgo.It("[RANDOM]", func() {
			dnschaostestcases.TestcaseDNSRandom(ns, cli, port, c)
		})

		ginkgo.It("[ERROR]", func() {
			dnschaostestcases.TestcaseDNSError(ns, cli, port, c)
		})
	})
	// DNS chaos case in [StressChaos] context
	ginkgo.Context("[StressChaos]", func() {
		var err error

		var ports []uint16
		var stressPeers []*v1.Pod
		var pfCancels []context.CancelFunc

		ginkgo.JustBeforeEach(func() {
			ports = []uint16{}
			stressPeers = []*v1.Pod{}
			for index := 0; index < 2; index++ {
				name := fmt.Sprintf("stress-peer-%d", index)

				svc := fixture.NewE2EService(name, ns)
				_, err = kubeCli.CoreV1().Services(ns).Create(svc)
				framework.ExpectNoError(err, "create service error")
				nd := fixture.NewStressTestDeployment(name, ns, map[string]string{"partition": strconv.Itoa(index % 2)})
				_, err = kubeCli.AppsV1().Deployments(ns).Create(nd)
				framework.ExpectNoError(err, "create network-peer deployment error")
				err = util.WaitDeploymentReady(name, ns, kubeCli)
				framework.ExpectNoError(err, "wait network-peer deployment ready error")

				pod, err := getPod(kubeCli, ns, name)
				framework.ExpectNoError(err, "select network-peer pod error")
				stressPeers = append(stressPeers, pod)

				_, port, pfCancel, err := portforward.ForwardOnePort(fw, ns, "svc/"+svc.Name, 8080)
				ports = append(ports, port)
				pfCancels = append(pfCancels, pfCancel)
				framework.ExpectNoError(err, "create helper io port port-forward failed")
			}
		})

		ginkgo.It("[CPU]", func() {
			stresstestcases.TestcaseCPUStressInjectionOnceThenRecover(ns, cli, stressPeers, ports, c)
		})

		// TODO: unstable test
		ginkgo.It("[Memory]", func() {
			stresstestcases.TestcaseMemoryStressInjectionOnceThenRecover(ns, cli, stressPeers, ports, c)
		})

		ginkgo.JustAfterEach(func() {
			for _, cancel := range pfCancels {
				cancel()
			}
		})
	})
})

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
