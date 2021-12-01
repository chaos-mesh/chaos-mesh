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

package common

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var app *fx.App
var k8sClient client.Client
var lister *utils.ActiveLister
var cfg *rest.Config
var testEnv *envtest.Environment
var setupLog = ctrl.Log.WithName("setup")
var testPod = fmt.Sprintf("test-%d", rand.Int63())

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Common suit",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	t := true
	if os.Getenv("USE_EXISTING_CLUSTER") == "true" {
		testEnv = &envtest.Environment{
			UseExistingCluster: &t,
		}
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
		}
	}

	err := v1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	By("create target pod")
	key := types.NamespacedName{
		Name:      testPod,
		Namespace: "default",
	}
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPod,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}
	Expect(k8sClient.Create(context.TODO(), &pod)).To(Succeed())

	By("poll pod to ready")
	{
		err := wait.Poll(time.Second*1, time.Second*30, func() (ok bool, err error) {
			err = k8sClient.Get(context.TODO(), key, &pod)
			if err != nil {
				return false, err
			}
			return pod.Status.Phase == v1.PodRunning, nil
		})
		Expect(err).ToNot(HaveOccurred())
	}

	By("start application")
	app = fx.New(
		fx.Options(
			test.Module,
			chaosimpl.AllImpl,
			selector.Module,
			fx.Provide(chaosdaemon.New),
			fx.Provide(AllSteps),
			fx.Invoke(Bootstrap),
			fx.Supply(cfg),
		),
		fx.Invoke(Run),
	)
	startCtx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		setupLog.Error(err, "fail to start manager")
	}
	Expect(err).ToNot(HaveOccurred())

}, 60)

var _ = AfterSuite(func() {
	By("delete target pod")
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPod,
			Namespace: "default",
		},
	}
	Expect(k8sClient.Delete(context.TODO(), &pod)).To(Succeed())

	By("tearing down the test environment")
	stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		setupLog.Error(err, "fail to stop manager")
	}
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

type RunParams struct {
	fx.In

	Mgr    ctrl.Manager
	Logger logr.Logger
}

func Run(params RunParams) error {
	lister = utils.NewActiveLister(k8sClient, params.Logger)
	return nil
}
