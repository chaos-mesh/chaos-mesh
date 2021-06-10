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

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/cmd/chaos-controller-manager/provider"
	"github.com/chaos-mesh/chaos-mesh/controllers/desiredphase"
	"github.com/chaos-mesh/chaos-mesh/controllers/finalizers"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule"
	"github.com/chaos-mesh/chaos-mesh/controllers/schedule/utils"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/test"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/controllers"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[Controllers]", func() {
	ginkgo.Context("[DesiredPhase]", func() {
		var app *fx.App
		ginkgo.JustBeforeEach(func() {
			app = startNewAppWithGivenOptions(fx.Provide(fx.Annotated{
				Group:  "controller",
				Target: desiredphase.NewController,
			}), Run)
		})

		ginkgo.JustAfterEach(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()

			if err := app.Stop(stopCtx); err != nil {
				setupLog.Error(err, "fail to stop manager")
			}
		})
		ginkgo.Context("Setting phase", func() {
			ginkgo.It("should set phase", func() {
				desiredphase.TestDesiredPhaseBasic(k8sClient)
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
			}), Run)
		})

		ginkgo.JustAfterEach(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()

			if err := app.Stop(stopCtx); err != nil {
				setupLog.Error(err, "fail to stop manager")
			}
		})
		ginkgo.Context("Add finalizer", func() {
			ginkgo.It("should add default finalizer", func() {
				finalizers.TestAddDefaultFinalizer(k8sClient)
			})
		})
	})
	ginkgo.Context("[Schedule]", func() {
		var app *fx.App
		ginkgo.JustBeforeEach(func() {
			app = startNewAppWithGivenOptions(schedule.Module, RunWithWorkflow)

		})

		ginkgo.JustAfterEach(func() {
			stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
			defer cancel()

			if err := app.Stop(stopCtx); err != nil {
				setupLog.Error(err, "fail to stop manager")
			}
		})
		ginkgo.It(("Should work fine"), func() {
			schedule.TestScheduleBasic(k8sClient)
			schedule.TestScheduleChaos(k8sClient)
			schedule.TestScheduleConcurrentChaos(k8sClient)
			schedule.TestScheduleGC(k8sClient)
			schedule.TestScheduleWorkflow(k8sClient)
			schedule.TestScheduleWorkflowGC(k8sClient)
		})
	})
})

func startNewAppWithGivenOptions(options fx.Option, Run func(RunParams) error) *fx.App {
	app := fx.New(
		fx.Options(
			fx.Provide(
				test.NewTestOption,
				provider.NewClient,
				provider.NewGlobalCacheReader,
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

type RunParams struct {
	fx.In

	Mgr    ctrl.Manager
	Logger logr.Logger

	Controllers []types.Controller `group:"controller"`
	Objs        []types.Object     `group:"objs"`
}

func Run(params RunParams) error {
	lister = utils.NewActiveLister(k8sClient, params.Logger)
	return nil
}

func RunWithWorkflow(params RunParams) error {
	lister = utils.NewActiveLister(k8sClient, params.Logger)
	err := controllers.BootstrapWorkflowControllers(params.Mgr, params.Logger)
	return err
}
