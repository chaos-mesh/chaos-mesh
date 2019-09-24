package api_server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"github.com/ngaut/log"
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

type Server struct {
	rdr     *render.Render
	storage *mysqlClient
}

func NewServer(dataSource string) (*Server, error) {
	rdr := render.New()
	storage, err := NewMysqlClient(dataSource)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Server{
		rdr,
		storage,
	}, nil
}

func (s *Server) CreateRouter() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/job", func(w http.ResponseWriter, r *http.Request) {
		job := new(Job)

		if err := readJSON(r.Body, job); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("read json body error %v", err))
			return
		}

		if err := job.Verify(); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("verify job %+v error %v", job, err))
			return
		}

		if err := s.storage.createJob(job); err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("create job %+v error %v", job, err))
			return
		}

		s.rdr.JSON(w, http.StatusOK, successResponse("ok"))
	}).Methods("PUT")

	router.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		filters, ok := vars["filters"]

		var fs Filters
		if ok {
			json.Unmarshal([]byte(filters), &fs)
		}

		jobs, err := s.storage.getJobs(&fs)
		if err != nil {
			s.rdr.JSON(w, http.StatusOK, errResponsef("select jobs %+v error %v", fs, err))
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
