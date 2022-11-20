// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/containerd/cgroups"
)

func main() {
	port := flag.Int("port", 8080, "listen port")
	dataDir := flag.String("data-dir", "/var/run/data", "data dir is the dir to write temp file, only used in io test")

	flag.Parse()

	s := newServer(*dataDir)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go s.childProcessTimeServer.Start(ctx)
	err := s.setupUDPServer()
	if err != nil {
		fmt.Println("failed to serve udp server", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	if err := http.ListenAndServe(addr, s.mux); err != nil {
		fmt.Println("failed to serve http server", err)
		os.Exit(1)
	}
}

type server struct {
	mux                    *http.ServeMux
	dataDir                string
	childProcessTimeServer childProcessTimeServer

	// ONLY FOR TEST: a buf without lock
	recvBuf []byte
}

func newServer(dataDir string) *server {
	s := &server{
		mux:                    http.NewServeMux(),
		dataDir:                dataDir,
		recvBuf:                make([]byte, 5),
		childProcessTimeServer: &defaultChildProcessTimeServer{},
	}
	s.mux.HandleFunc("/ping", pong)
	s.mux.HandleFunc("/time", s.time)
	s.mux.HandleFunc("/child-process-time", s.childProcessTime)
	s.mux.HandleFunc("/io", s.ioTest)
	s.mux.HandleFunc("/mistake", s.mistakeTest)
	s.mux.HandleFunc("/network/send", s.networkSendTest)
	s.mux.HandleFunc("/network/recv", s.networkRecvTest)
	s.mux.HandleFunc("/network/ping", s.networkPingTest)
	s.mux.HandleFunc("/dns", s.dnsTest)
	s.mux.HandleFunc("/stress", s.stressCondition)
	s.mux.HandleFunc("/http", s.httpEcho)
	s.mux.HandleFunc("/setup_https", s.SetupHTTPSServer)
	return s
}

func pong(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("pong"))
}

func (s *server) setupUDPServer() error {
	pc, err := net.ListenPacket("udp", "0.0.0.0:1070")
	if err != nil {
		return err
	}

	go func() {
		for {
			_, _, err := pc.ReadFrom(s.recvBuf)
			fmt.Println("receive buf " + string(s.recvBuf))
			if err != nil {
				return
			}
		}
	}()

	return nil
}

// time write the current time as response body
func (s *server) time(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte(time.Now().Format(time.RFC3339Nano)))
}

// childProcessTime write the child process current time as response body
func (s *server) childProcessTime(w http.ResponseWriter, _ *http.Request) {
	now, err := s.childProcessTimeServer.Time()
	if err != nil {
		panic(err)
	}
	w.Write([]byte(now.Format(time.RFC3339Nano)))
}

// a handler to test io chaos
func (s *server) mistakeTest(w http.ResponseWriter, _ *http.Request) {
	path := filepath.Join(s.dataDir, "e2e-test")
	origData := []byte("hello world!!!!!!!!!!!!")

	err := os.WriteFile(path, origData, 0644)
	if err != nil {
		w.Write([]byte(fmt.Sprintf("failed to write file %v", err)))
		return
	}
	gotData, err := os.ReadFile(path)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	result := bytes.Equal(origData, gotData)
	if result {
		w.Write([]byte("false"))
		return
	}
	for i := 0; i < 10; i++ {
		tmp, err := os.ReadFile(path)
		if err != nil {
			w.Write([]byte(err.Error()))
		}
		if !bytes.Equal(tmp, gotData) {
			w.Write([]byte("true"))
			return
		}
	}
	w.Write([]byte("err"))
}

// a handler to test io chaos
func (s *server) ioTest(w http.ResponseWriter, _ *http.Request) {
	t1 := time.Now()
	f, err := os.CreateTemp(s.dataDir, "e2e-test")
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

// a handler to test dns chaos
func (s *server) dnsTest(w http.ResponseWriter, r *http.Request) {

	url, ok := r.URL.Query()["url"]

	if !ok || len(url[0]) < 1 {
		http.Error(w, "failed", http.StatusBadRequest)
		return
	}

	ips, err := net.LookupIP(url[0])
	if err != nil {
		http.Error(w, "failed", http.StatusBadRequest)
		return
	}

	if len(ips) == 0 {
		http.Error(w, "failed", http.StatusBadRequest)
		return
	}

	w.Write([]byte(ips[0].String()))
}

type networkSendTestBody struct {
	TargetIP string `json:"targetIP"`
}

// a handler to test network chaos
func (s *server) networkPingTest(w http.ResponseWriter, r *http.Request) {
	var body networkSendTestBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := http.Client{
		Timeout: 2 * time.Second,
	}
	startTime := time.Now()
	resp, err := c.Get(fmt.Sprintf("http://%s:8080/ping", body.TargetIP))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	endTime := time.Now()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if string(out) != "pong" {
		http.Error(w, "response is not pong", http.StatusBadRequest)
		return
	}

	w.Write([]byte(fmt.Sprintf("OK %d", endTime.UnixNano()-startTime.UnixNano())))
}

// a handler to test network chaos
func (s *server) networkSendTest(w http.ResponseWriter, r *http.Request) {
	var body networkSendTestBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(body.TargetIP),
		Port: 1070,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	n, err := io.WriteString(conn, "ping\n")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if n != 5 {
		http.Error(w, "udp send less than 5 bytes", http.StatusBadRequest)
		return
	}
	w.Write([]byte("send successfully\n"))
}

// a handler to test network chaos
func (s *server) networkRecvTest(w http.ResponseWriter, r *http.Request) {
	w.Write(s.recvBuf)

	s.recvBuf = []byte{}
}

func (s *server) stressCondition(w http.ResponseWriter, r *http.Request) {
	control, err := cgroups.Load(cgroups.V1, cgroups.PidPath(1))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats, err := control.Stat(cgroups.IgnoreNotExist)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(map[string]uint64{
		"cpuTime":     stats.CPU.Usage.Total,
		"memoryUsage": stats.Memory.Usage.Usage - stats.Memory.Kernel.Usage - stats.Memory.Cache,
	})
	if err != nil {
		http.Error(w, "fail to marshal response", http.StatusInternalServerError)
		return
	}

	w.Write(response)
}

func (s *server) httpEcho(w http.ResponseWriter, r *http.Request) {
	secrets := r.Header["Secret"]
	if len(secrets) == 0 {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	for _, secret := range secrets {
		w.Header().Add("Secret", secret)
	}
	defer r.Body.Close()
	_, err := io.Copy(w, r.Body)
	if err != nil {
		http.Error(w, "fail to copy body between request and response", http.StatusInternalServerError)
		return
	}
}

type TLSServerKeys struct {
	Cert []byte `json:"cert"`
	Key  []byte `json:"key"`
}

func (s *server) SetupHTTPSServer(w http.ResponseWriter, r *http.Request) {
	var body TLSServerKeys
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = os.WriteFile("/tmp/server.crt", body.Cert, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = os.WriteFile("/tmp/server.key", body.Key, 0644)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	srv := &server{
		mux: http.NewServeMux(),
	}
	srv.mux.HandleFunc("/ping", pong)

	go func() {
		panic(http.ListenAndServeTLS("0.0.0.0:8081", "/tmp/server.crt", "/tmp/server.key", srv.mux))
	}()
	if err != nil {
		return
	}
}
