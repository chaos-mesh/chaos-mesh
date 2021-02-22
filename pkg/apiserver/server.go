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

package apiserver

import (
	"fmt"
	"net/http"

	"go.uber.org/fx"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	apiutils "github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/apivalidator"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/swaggerserver"
	"github.com/chaos-mesh/chaos-mesh/pkg/uiserver"
)

var (
	// Module includes the providers (gin engine and api router) and the registers.
	Module = fx.Options(
		fx.Provide(
			newEngine,
			newAPIRouter,
		),
		handlerModule,
		fx.Invoke(serverRegister),
	)
)

func serverRegister(r *gin.Engine, conf *config.ChaosDashboardConfig) {
	listenAddr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)

	go r.Run(listenAddr)
}

func newEngine() *gin.Engine {
	r := gin.Default()

	// default is "/debug/pprof/"
	pprof.Register(r)

	r.Use(apiutils.MWHandleErrors())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("NameValid", apivalidator.NameValid)
		v.RegisterValidation("NamespaceSelectorsValid", apivalidator.NamespaceSelectorsValid)
		v.RegisterValidation("MapSelectorsValid", apivalidator.MapSelectorsValid)
		v.RegisterValidation("RequirementSelectorsValid", apivalidator.RequirementSelectorsValid)
		v.RegisterValidation("PhaseSelectorsValid", apivalidator.PhaseSelectorsValid)
		v.RegisterValidation("CronValid", apivalidator.CronValid)
		v.RegisterValidation("DurationValid", apivalidator.DurationValid)
		v.RegisterValidation("ValueValid", apivalidator.ValueValid)
		v.RegisterValidation("PodsValid", apivalidator.PodsValid)
		v.RegisterValidation("RequiredFieldEqual", apivalidator.RequiredFieldEqualValid, true)
	}

	moveToUIRoot := func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/dashboard")
	}

	r.GET("/", moveToUIRoot)
	ui := uiserver.AssetsFS()
	if ui != nil {
		newDashboardRouter(r, ui)
	} else {
		r.GET("/dashboard", func(c *gin.Context) {
			c.String(http.StatusOK, "Dashboard UI is not built. Please run `UI=1 make`.")
		})
	}
	r.NoRoute(moveToUIRoot)

	return r
}

func newAPIRouter(r *gin.Engine) *gin.RouterGroup {
	api := r.Group("/api")
	{
		api.GET("/swagger/*any", swaggerserver.Handler())
	}

	return api
}

func newDashboardRouter(r *gin.Engine, ui http.FileSystem) {
	renderRequest := func(c *gin.Context) {
		c.FileFromFS(c.Request.URL.Path, ui)
	}

	r.GET("/dashboard/*any", func(c *gin.Context) {
		c.FileFromFS("/", ui)
	})
	r.GET("/static/*any", renderRequest)
	r.GET("/favicon.ico", renderRequest)
}
