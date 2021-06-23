// Copyright 2021 Chaos Mesh Authors.
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

package schedule

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/controllers"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var app *fx.App
var k8sClient client.Client
var lister *utils.ActiveLister
var config *rest.Config
var testEnv *envtest.Environment
var setupLog = ctrl.Log.WithName("setup")

func TestSchedule(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Schedule suit",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	By("bootstrapping test environment")
	t := true
	if os.Getenv("TEST_USE_EXISTING_CLUSTER") == "true" {
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

	config, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(config).ToNot(BeNil())

	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	app = fx.New(
		fx.Options(
			test.Module,
			fx.Supply(config),
			Module,
			types.ChaosObjects,
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

	Mgr             ctrl.Manager
	Logger          logr.Logger
	RecorderBuilder *recorder.RecorderBuilder

	Controllers []types.Controller `group:"controller"`
	Objs        []types.Object     `group:"objs"`
}

func Run(params RunParams) error {
	lister = utils.NewActiveLister(k8sClient, params.Logger)
	err := controllers.BootstrapWorkflowControllers(params.Mgr, params.Logger, params.RecorderBuilder)
	if err != nil {
		return err
	}
	return nil
}
