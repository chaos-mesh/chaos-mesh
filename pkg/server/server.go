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

package server

import (
	"net/http"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gorilla/mux"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Server struct {
	router *mux.Router
	log    logr.Logger
	client client.Client
}

func SetupServer(client client.Client) Server {
	r := mux.NewRouter()

	server := Server{
		router: r,
		log:    ctrl.Log.WithName("server"),
		client: client,
	}

	r.PathPrefix("/dashboard/").HandlerFunc(server.dashboard)
	r.PathPrefix("/services").HandlerFunc(server.services)
	r.PathPrefix("/").Handler(server.web("/", "/web"))

	return server
}

func (s *Server) Run() {
	s.log.Info("Starting Server on 0.0.0.0:80")
	err := http.ListenAndServe("0.0.0.0:80", s.router)
	if err != nil {
		s.log.Error(err, "Error while listening 0.0.0.0:80")
	}
}
