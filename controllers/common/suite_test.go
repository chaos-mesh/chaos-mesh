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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/condition"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/desiredphase"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/finalizers"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/pipeline"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
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

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Common suit")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	logf.SetLogger(log.NewZapLoggerWithWriter(GinkgoWriter))
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

	rootLogger, err := log.NewDefaultZapLogger()
	Expect(err).ToNot(HaveOccurred())
	By("start application")
	app = fx.New(
		fx.Options(
			fx.Supply(rootLogger),
			test.Module,
			chaosimpl.AllImpl,
			selector.Module,
			fx.Provide(chaosdaemon.New),
			fx.Provide(func() []pipeline.PipelineStep {
				return []pipeline.PipelineStep{finalizers.InitStep, desiredphase.Step, condition.Step, finalizers.CleanStep}
			}),
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

}, NodeTimeout(60*time.Second))

var _ = AfterSuite(func() {
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
