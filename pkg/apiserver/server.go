// Copyright 2020 PingCAP, Inc.
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
	"context"
	"fmt"
	"net"
	"net/http"
	"os"

	"go.uber.org/fx"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	apiutils "github.com/chaos-mesh/chaos-mesh/pkg/apiserver/utils"
	"github.com/chaos-mesh/chaos-mesh/pkg/apivalidator"
	"github.com/chaos-mesh/chaos-mesh/pkg/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/swaggerserver"
	"github.com/chaos-mesh/chaos-mesh/pkg/uiserver"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// Module includes the providers and registers provided by apiserver and handlers.
	Module = fx.Options(
		fx.Provide(
			uiserver.AssetFS,
			newAPIHandlerEngine,
			NewServer,
		),
		fx.Invoke(serverRegister),
		handlerModule)

	log = ctrl.Log.WithName("apiserver")
)

// Server is a server to run API service.
type Server struct {
	ctx context.Context

	conf *config.ChaosDashboardConfig

	uiAssetFS        http.FileSystem
	apiHandlerEngine *gin.Engine
}

// NewServer returns a Server instance.
func NewServer(
	conf *config.ChaosDashboardConfig,
	uiAssetFS http.FileSystem,
	apiHandlerEngine *gin.Engine,
) *Server {
	return &Server{
		conf:             conf,
		uiAssetFS:        uiAssetFS,
		apiHandlerEngine: apiHandlerEngine,
	}
}

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	s.apiHandlerEngine.ServeHTTP(w, r)
}

func serverRegister(lx fx.Lifecycle, s *Server, conf *config.ChaosDashboardConfig, assetFs http.FileSystem) {
	listenAddr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Error(err, "chaos-dashboard listen failed", "host", conf.ListenHost, "port", conf.ListenPort)
		os.Exit(1)
	}

	mux := http.DefaultServeMux

	mux.Handle("/", http.StripPrefix("/", uiserver.Handler(assetFs)))
	mux.Handle("/api/", Handler(s))
	mux.Handle("/api/swagger/", swaggerserver.Handler())
	mux.HandleFunc("/ping", pingHandler)

	srv := &http.Server{Handler: mux}

	lx.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.Serve(listener); err != http.ErrServerClosed {
					log.Error(err, "chaos-dashboard aborted with an error")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Close()
		},
	})
}

func newAPIHandlerEngine() (*gin.Engine, *gin.RouterGroup) {
	apiHandlerEngine := gin.Default()
	apiHandlerEngine.Use(apiutils.MWHandleErrors())

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("NameValid", apivalidator.NameValid)
		v.RegisterValidation("NamespaceSelectorsValid", apivalidator.NamespaceSelectorsValid)
		v.RegisterValidation("MapSelectorsValid", apivalidator.MapSelectorsValid)
		v.RegisterValidation("PhaseSelectorsValid", apivalidator.PhaseSelectorsValid)
		v.RegisterValidation("CronValid", apivalidator.CronValid)
		v.RegisterValidation("DurationValid", apivalidator.DurationValid)
		v.RegisterValidation("ValueValid", apivalidator.ValueValid)
		v.RegisterValidation("PodsValid", apivalidator.PodsValid)
	}

	endpoint := apiHandlerEngine.Group("/api")

	return apiHandlerEngine, endpoint
}

// Handler returns a `http.Handler`
func Handler(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handler(w, r)
	})
}

func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "pong\n")
}
