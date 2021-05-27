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

package controllers

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/controllers/desiredphase"
	"github.com/chaos-mesh/chaos-mesh/controllers/finalizers"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/test"
	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[Controllers]", func() {
	ginkgo.Context("[DesiredPhase]", func() {
		var app *fx.App
		ginkgo.JustBeforeEach(func() {
			app = startNewAppWithGivenOptions(fx.Provide(fx.Annotated{
				Group:  "controller",
				Target: desiredphase.NewController,
			}))
		})

		ginkgo.JustAfterEach(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()

			if err := app.Stop(stopCtx); err != nil {
				setupLog.Error(err, "fail to stop manager")
			}
		})
		Context("Setting phase", func() {
			It("should set phase", func() {
				desiredphase.TestDesiredPhaseBasic(k8sClient)
			})
			It("should set phase due to pause", func() {
				desiredphase.TestDesiredPhasePause(k8sClient)
			})
		})
	})
	ginkgo.Context("[Finalizer]", func() {
		var app *fx.App
		ginkgo.JustBeforeEach(func() {
			app = startNewAppWithGivenOptions(fx.Provide(fx.Annotated{
				Group:  "controller",
				Target: finalizers.NewController,
			}))
		})

		ginkgo.JustAfterEach(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()

			if err := app.Stop(stopCtx); err != nil {
				setupLog.Error(err, "fail to stop manager")
			}
		})
		Context("Add finalizer", func() {
			It("should add default finalizer", func() {
				finalizers.TestAddDefaultFinalizer(k8sClient)
			})
		})
	})
	ginkgo.Context("[Schedule]", func() {
		var app *fx.App
		ginkgo.JustBeforeEach(func() {
			app = startNewAppWithGivenOptions(schedule.Module)
		})

		ginkgo.JustAfterEach(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()

			if err := app.Stop(stopCtx); err != nil {
				setupLog.Error(err, "fail to stop manager")
			}
		})
		Context("Basic", func() {
			It(("Should be created and deleted successfully"), func() {
				schedule.TestScheduleBasic(k8sClient)
			})
		})
		Context("Common chaos", func() {
			It("should create non-concurrent chaos", func() {
				schedule.TestScheduleChaos(k8sClient)

			})
			It("should create concurrent chaos", func() {
				schedule.TestScheduleConcurrentChaos(k8sClient)

			})
			It("should collect garbage", func() {
				schedule.TestScheduleGC(k8sClient)
			})
		})
		Context("workflow", func() {
			It("should create non-concurrent workflow", func() {
				schedule.TestScheduleWorkflow(k8sClient)
			})
			It("should collect garbage", func() {
				schedule.TestScheduleWorkflowGC(k8sClient)
			})
		})
	})
})

func startNewAppWithGivenOptions(options fx.Option) *fx.App {
	app := fx.New(
		fx.Options(
			fx.Provide(
				NewTestOption,
				provider.NewClient,
				provider.NewReader,
				provider.NewLogger,
				provider.NewAuthCli,
				provider.NewScheme,
				test.NewTestManager,
			),
			options,
			fx.Supply(config),
			types.ChaosObjects,
		),
		fx.Invoke(Run),
	)
	startCtx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()
	var err error
	if err = app.Start(startCtx); err != nil {
		setupLog.Error(err, "fail to start manager")
	}
	Expect(err).ToNot(HaveOccurred())
	return app
}

func NewTestOption(logger logr.Logger) *ctrl.Options {
	setupLog := logger.WithName("setup")
	scheme := runtime.NewScheme()

	clientgoscheme.AddToScheme(scheme)

	v1alpha1.AddToScheme(scheme)
	fmt.Println("Bind to port:", 9443+ginkgoConfig.GinkgoConfig.ParallelNode)
	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: ":" + fmt.Sprint(10080+ginkgoConfig.GinkgoConfig.ParallelNode),
		LeaderElection:     ccfg.ControllerCfg.EnableLeaderElection,
		Port:               9443 + ginkgoConfig.GinkgoConfig.ParallelNode,
	}

	if ccfg.ControllerCfg.ClusterScoped {
		setupLog.Info("Chaos controller manager is running in cluster scoped mode.")
		// will not specific a certain namespace
	} else {
		setupLog.Info("Chaos controller manager is running in namespace scoped mode.", "targetNamespace", ccfg.ControllerCfg.TargetNamespace)
		options.Namespace = ccfg.ControllerCfg.TargetNamespace
	}

	return &options
}
