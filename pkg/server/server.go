package server

import (
	"github.com/go-logr/logr"
	"net/http"

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
