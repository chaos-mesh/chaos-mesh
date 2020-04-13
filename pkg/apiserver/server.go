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

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gin-gonic/gin"

	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/swaggerserver"
	"github.com/pingcap/chaos-mesh/pkg/uiserver"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	Module = fx.Options(
		fx.Provide(
			newAPIHandlerEngine,
			NewServer,
		),
		handlerModule,
		fx.Invoke(serverRegister))

	log = ctrl.Log.WithName("apiserver")
)

type Server struct {
	ctx context.Context

	conf *config.ChaosServerConfig

	uiAssetFS        *assetfs.AssetFS
	apiHandlerEngine *gin.Engine
}

func NewServer(
	lx fx.Lifecycle,
	conf *config.ChaosServerConfig,
	uiAssetFS *assetfs.AssetFS,
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

func serverRegister(lx fx.Lifecycle, s *Server, conf *config.ChaosServerConfig) {
	listenAddr := fmt.Sprintf("%s:%d", conf.ListenHost, conf.ListenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Error(err, "chaos-server listen failed", "host", conf.ListenHost, "port", conf.ListenPort)
		os.Exit(1)
	}

	mux := http.DefaultServeMux

	mux.Handle("/", http.StripPrefix("/", uiserver.Handler()))
	mux.Handle("/api/", Handler(s))
	mux.Handle("/api/swagger/", swaggerserver.Handler())

	srv := &http.Server{Handler: mux}

	lx.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				if err := srv.Serve(listener); err != http.ErrServerClosed {
					log.Error(err, "chaos-server aborted with an error")
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
	apiHandlerEngine := gin.New()
	endpoint := apiHandlerEngine.Group("/api")

	return apiHandlerEngine, endpoint
}

func Handler(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handler(w, r)
	})
}
