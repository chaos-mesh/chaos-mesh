package tcdaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/unrolled/render"
)

// Server represents an HTTP server for tc daemon
type Server struct {
	rdr      *render.Render
	crClient ContainerRuntimeInfoClient
}

func readJSON(r io.ReadCloser, data interface{}) error {
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err == nil {
		err = json.Unmarshal(b, data)
	}

	return errors.Trace(err)
}

func newServer() (*Server, error) {
	crClient, err := CreateContainerRuntimeInfoClient()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Server{
		rdr:      render.New(),
		crClient: crClient,
	}, nil
}

func (s *Server) parseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		containerID := vars["containerID"]
		netem := new(Netem)

		if err := readJSON(r.Body, netem); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("read json body error %v", err))
			return
		}
		context.Set(r, "netem", netem)

		pid, err := s.crClient.GetPidFromContainerID(ctx, containerID)
		if err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("get pid from containerID error: %v", err))
		}
		context.Set(r, "pid", pid)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("Request : %s - %s - %s", r.RemoteAddr, r.Method, r.URL)
		start := time.Now()

		next.ServeHTTP(w, r)
		glog.Infof("Response: %s - %s - %s (%.3f sec)", r.RemoteAddr, r.Method, r.URL, time.Since(start).Seconds())
	})
}

// CreateRouter will create a Handler for related Server
func (s *Server) CreateRouter() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/{containerID}/netem", func(w http.ResponseWriter, r *http.Request) {
		netem := context.Get(r, "netem").(*Netem)
		pid := context.Get(r, "pid").(int)

		if err := netem.Apply(pid); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("delay apply error: %v", err))
			return
		}

		s.rdr.JSON(w, http.StatusOK, successResponse("ok"))
	}).Methods("PUT")

	router.HandleFunc("/{containerID}/netem", func(w http.ResponseWriter, r *http.Request) {
		netem := context.Get(r, "netem").(*Netem)
		pid := context.Get(r, "pid").(int)

		if err := netem.Cancel(pid); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("delay apply error: %v", err))
			return
		}

		s.rdr.JSON(w, http.StatusOK, successResponse("ok"))
	}).Methods("DELETE")

	router.Use(s.parseMiddleware)

	router.Use(s.logMiddleware)

	return router
}

// StartServer will start listening on newly created server
func StartServer(host string, port int) {
	server, err := newServer()

	if err != nil {
		glog.Errorf("Error while starting server: %v", err)
	}

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), server.CreateRouter())
}
