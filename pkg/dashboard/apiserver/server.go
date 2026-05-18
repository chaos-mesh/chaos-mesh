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

package apiserver

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	controllermetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	config "github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/apivalidator"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/swaggerserver"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/uiserver"
	"github.com/chaos-mesh/chaos-mesh/pkg/metrics"
)

var (
	// Module includes the providers (gin engine and api router) and the registers.
	Module = fx.Options(
		fx.Provide(
			newEngine,
			newAPIRouter,
		),
		handlerModule,
		fx.Provide(func() prometheus.Registerer {
			return controllermetrics.Registry
		}),
		fx.Invoke(metrics.NewChaosDashboardMetricsCollector),
		fx.Invoke(register),
	)
)

func register(r *gin.Engine, conf *config.ChaosDashboardConfig) {
	listenAddr := net.JoinHostPort(conf.ListenHost, fmt.Sprintf("%d", conf.ListenPort))

	go r.Run(listenAddr)
}

func newEngine(config *config.ChaosDashboardConfig) *gin.Engine {
	r := gin.Default()

	if prefix := normalizeBasePath(config.UIBasePath); prefix != "" {
		r.Use(stripBasePath(prefix))
	}

	if config.EnableProfiling {
		// default is "/debug/pprof"
		pprof.Register(r)
	}

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
		v.RegisterValidation("PhysicalMachineValid", apivalidator.PhysicalMachineValid)
	}

	ui := uiserver.AssetsFS()
	if ui != nil {
		r.NoRoute(func(c *gin.Context) {
			c.FileFromFS(c.Request.URL.Path, ui)
		})
	} else {
		r.GET("/", func(c *gin.Context) {
			c.String(http.StatusOK, "Dashboard UI is not built.")
		})
	}

	return r
}

func newAPIRouter(r *gin.Engine) *gin.RouterGroup {
	api := r.Group("/api")

	api.GET("/swagger/*any", swaggerserver.Handler)

	return api
}

// normalizeBasePath returns the prefix with a single leading "/" and no trailing
// "/", or an empty string if the input represents the root ("" or "/"). The
// returned prefix is the form used for path matching by stripBasePath.
func normalizeBasePath(p string) string {
	p = "/" + strings.Trim(p, "/")
	if p == "/" {
		return ""
	}
	return p
}

// stripBasePath returns gin middleware that removes prefix from incoming
// request paths so the rest of the router (UI assets and /api routes) can match
// unchanged. The prefix must be a normalized base path produced by
// normalizeBasePath. Paths that do not start with prefix (followed by "/" or
// end-of-path) are passed through untouched, so the same binary continues to
// work when the ingress already strips the prefix or when the dashboard is
// served at the root.
func stripBasePath(prefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := c.Request.URL.Path
		switch {
		case p == prefix:
			c.Request.URL.Path = "/"
		case strings.HasPrefix(p, prefix+"/"):
			c.Request.URL.Path = strings.TrimPrefix(p, prefix)
		}
		c.Next()
	}
}
