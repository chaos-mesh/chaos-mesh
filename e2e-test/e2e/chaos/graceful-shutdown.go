// Copyright 2021 Chaos Mesh Authors.
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
//

package chaos

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	restClient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/pod-security-admission/api"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	httpchaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/httpchaos"
	iochaostestcases "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/chaos/iochaos"
	e2econfig "github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/e2e/util"
	"github.com/chaos-mesh/chaos-mesh/e2e-test/pkg/fixture"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/portforward" // testcases
)

var _ = ginkgo.Describe("[Graceful-Shutdown]", func() {
	f := framework.NewDefaultFramework("chaos-mesh")
	f.NamespacePodSecurityEnforceLevel = api.LevelPrivileged
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
		logger, err := log.NewDefaultZapLogger()
		framework.ExpectNoError(err, "failed to create logger")
		fw, err = portforward.NewPortForwarder(ctx, e2econfig.NewSimpleRESTClientGetter(clientRawConfig), true, logger)
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

	// io chaos case in [IOChaos] context
	ginkgo.Context("[IOChaos]", func() {
		var (
			err      error
			port     uint16
			pfCancel context.CancelFunc
		)

		ginkgo.JustBeforeEach(func() {
			svc := fixture.NewE2EService("io", ns)
			_, err = kubeCli.CoreV1().Services(ns).Create(context.TODO(), svc, metav1.CreateOptions{})
			framework.ExpectNoError(err, "create service error")
			nd := fixture.NewIOTestDeployment("io-test", ns)
			_, err = kubeCli.AppsV1().Deployments(ns).Create(context.TODO(), nd, metav1.CreateOptions{})
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

		// io chaos case in [Shutdown] context
		ginkgo.It("[Shutdown]", func() {
			iochaostestcases.TestcaseIOErrorGracefulShutdown(ns, cli, c, port)
		})
	})

	//http chaos case in [HTTPChaos] context
	ginkgo.Context("[HTTPChaos]", func() {
		var (
			err      error
			port     uint16
			pfCancel context.CancelFunc
			client   httpchaostestcases.HTTPE2EClient
		)

		ginkgo.JustBeforeEach(func() {
			svc := fixture.NewE2EService("http", ns)
			svc, err = kubeCli.CoreV1().Services(ns).Create(context.TODO(), svc, metav1.CreateOptions{})
			framework.ExpectNoError(err, "create service error")
			for _, servicePort := range svc.Spec.Ports {
				if servicePort.Name == "http" {
					port = uint16(servicePort.NodePort)
					break
				}
			}
			nd := fixture.NewHTTPTestDeployment("http-test", ns)
			_, err = kubeCli.AppsV1().Deployments(ns).Create(context.TODO(), nd, metav1.CreateOptions{})
			framework.ExpectNoError(err, "create http-test deployment error")
			err = util.WaitDeploymentReady("http-test", ns, kubeCli)
			framework.ExpectNoError(err, "wait http-test deployment ready error")
			podlist, err := kubeCli.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
			framework.ExpectNoError(err, "find pod list error")
			for _, item := range podlist.Items {
				if strings.Contains(item.Name, "http-test") {
					framework.Logf("get http-test-pod %v", item)
					client.IP = item.Status.HostIP
					break
				}
			}
			client.C = &c
		})

		ginkgo.JustAfterEach(func() {
			if pfCancel != nil {
				pfCancel()
			}
		})

		// http chaos case in [Shutdown] context
		ginkgo.It("[Shutdown]", func() {
			httpchaostestcases.TestcaseHttpGracefulAbortShutdown(ns, cli, client, port)
		})
	})
})
