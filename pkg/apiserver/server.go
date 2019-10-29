package apiserver

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/chaos-operator/pkg/apiserver/filter"
	"github.com/pingcap/chaos-operator/pkg/apiserver/storage"
	"github.com/pingcap/chaos-operator/pkg/apiserver/types"
	"github.com/unrolled/render"
)

func readJSON(r io.ReadCloser, data interface{}) error {
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err == nil {
		err = json.Unmarshal(b, data)
	}

	return errors.Trace(err)
}

// Server represents API Server
type Server struct {
	rdr     *render.Render
	storage *storage.SqlClient
}

// NewServer will create a Server
func NewServer() (*Server, error) {
	rdr := render.New()
	storage, err := storage.NewSqlClient()
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Server{
		rdr,
		storage,
	}, nil
}

// CreateRouter will create router for Server
func (s *Server) CreateRouter() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/task", func(w http.ResponseWriter, r *http.Request) {
		job := new(types.Task)

		if err := readJSON(r.Body, job); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("read json body error %v", err))
			return
		}

		if err := job.Verify(); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("verify job %+v error %v", job, err))
			return
		}

		if err := s.storage.CreateTask(job); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("create job %+v error %v", job, err))
			return
		}

		s.rdr.JSON(w, http.StatusOK, successResponse("ok"))
	}).Methods("PUT")

	router.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		filters, ok := r.URL.Query()["filters"]

		var filter filter.Filter
		if ok && len(filters) > 0 {
			json.Unmarshal([]byte(filters[0]), &filter)
		}

		jobs, err := s.storage.GetTasks(&filter)
		if err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("select jobs %+v error %v", filter, err))
		}

		s.rdr.JSON(w, http.StatusOK, successResponse(jobs))
	}).Methods("GET")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Request : %s - %s - %s", r.RemoteAddr, r.Method, r.URL)
		start := time.Now()

		router.ServeHTTP(w, r)
		log.Infof("Response: %s - %s - %s (%.3f sec)", r.RemoteAddr, r.Method, r.URL, time.Since(start).Seconds())
	})
}
