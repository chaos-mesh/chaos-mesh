package tcdaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/unrolled/render"
)

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

func (s *Server) CreateRouter() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/{containerId}/netem", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		containerId := vars["containerId"]
		delay := new(Netem)

		if err := readJSON(r.Body, delay); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("read json body error %v", err))
			return
		}

		if err := delay.Verify(); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("verify delay %+v error %v", delay, err))
			return
		}

		pid, err := s.crClient.GetPidFromContainerId(ctx, containerId)
		if err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("get pid from containerId error: %v", err))
		}

		if err := delay.Apply(pid); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("delay apply error: %v", err))
			return
		}

		s.rdr.JSON(w, http.StatusOK, successResponse("ok"))
	}).Methods("PUT")

	router.HandleFunc("/{containerId}/netem", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		vars := mux.Vars(r)

		containerId := vars["containerId"]
		delay := new(Netem)

		if err := readJSON(r.Body, delay); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("read json body error %v", err))
			return
		}

		if err := delay.Verify(); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("verify delay %+v error %v", delay, err))
			return
		}

		pid, err := s.crClient.GetPidFromContainerId(ctx, containerId)
		if err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("get pid from containerId error: %v", err))
		}

		if err := delay.Cancel(pid); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("delay apply error: %v", err))
			return
		}

		s.rdr.JSON(w, http.StatusOK, successResponse("ok"))
	}).Methods("DELETE")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		glog.Infof("Request : %s - %s - %s", r.RemoteAddr, r.Method, r.URL)
		start := time.Now()

		router.ServeHTTP(w, r)
		glog.Infof("Response: %s - %s - %s (%.3f sec)", r.RemoteAddr, r.Method, r.URL, time.Since(start).Seconds())
	})
}

func StartServer(host string, port int) {
	server, err := newServer()

	if err != nil {
		glog.Errorf("Error while starting server: %v", err)
	}

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), server.CreateRouter())
}
