// Copyright 2019 PingCAP, Inc.
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

package server

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"

	"github.com/pingcap/chaos-mesh/pkg/config"
	"github.com/pingcap/chaos-mesh/pkg/swaggerserver"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Server struct {
	log logr.Logger
}

func SetupServer(conf config.ChaosServerConfig) Server {
	r := mux.NewRouter()

	server := Server{
		log: ctrl.Log.WithName("server"),
	}

	// r.PathPrefix("/dashboard/").HandlerFunc(server.dashboard)
	// r.PathPrefix("/").Handler(server.web("/", "/web"))

	mux := http.DefaultServeMux
	// mux.Handle("/dashboard/", http.StripPrefix("/dashboard", uiserver.Handler()))
	// mux.Handle("/dashboard/api/", apiserver.Handler(s))
	// mux.Handle("/dashboard/api/swagger/", swaggerserver.Handler())

	// server.log.Info(fmt.Sprintf("Dashboard server is listening at %s", listenAddr))
	// server.log.Info(fmt.Sprintf("UI:      http://127.0.0.1:%d/dashboard/", cliConfig.ListenPort))
	// server.log.Info(fmt.Sprintf("API:     http://127.0.0.1:%d/dashboard/api/", cliConfig.ListenPort))
	// server.log.Info(fmt.Sprintf("Swagger: http://127.0.0.1:%d/dashboard/api/swagger/", cliConfig.ListenPort))

	return server
}

func (s *Server) Run() {
	s.log.Info("Starting Server on 0.0.0.0:80")
	err := http.ListenAndServe("0.0.0.0:80", s.router)
	if err != nil {
		s.log.Error(err, "Error while listening 0.0.0.0:80")
	}
}
