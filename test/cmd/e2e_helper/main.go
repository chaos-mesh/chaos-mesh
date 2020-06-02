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

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "listen port")
	dataDir := flag.String("data-dir", "/var/run/data/test", "data dir is the dir to write temp file, only used in io test")

	flag.Parse()

	s := newServer(*dataDir)
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		fmt.Println("failed to serve http server", err)
		os.Exit(1)
	}
}

type server struct {
	mux     *http.ServeMux
	dataDir string
}

func newServer(dataDir string) *server {
	s := &server{
		mux:     http.NewServeMux(),
		dataDir: dataDir,
	}
	s.mux.HandleFunc("/ping", pong)
	s.mux.HandleFunc("/time", s.timer)
	s.mux.HandleFunc("/io", s.ioTest)
	return s
}

func pong(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("pong"))
}

// a handler to print out the current time
func (s *server) timer(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(time.Now().Format(time.RFC3339Nano)))
}

// a handler to test io chaos
func (s *server) ioTest(w http.ResponseWriter, _ *http.Request) {
	t1 := time.Now()
	f, err := ioutil.TempFile(s.dataDir, "e2e-test")
	if err != nil {
		w.Write([]byte(fmt.Sprintf("failed to create temp file %v", err)))
		return
	}
	if _, err := f.Write([]byte("hello world")); err != nil {
		w.Write([]byte(fmt.Sprintf("failed to write file %v", err)))
		return
	}
	t2 := time.Now()
	w.Write([]byte(t2.Sub(t1).String()))
}
