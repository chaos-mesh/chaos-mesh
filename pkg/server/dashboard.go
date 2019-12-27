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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"
)

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	ss := strings.Split(r.URL.Path, "/")
	if len(ss) < 3 {
		w.WriteHeader(400)
		return
	}

	u := "http://" + ss[2] + "-chaos-grafana:3000"

	grafanaUrl, err := url.Parse(u)
	if err != nil {
		w.WriteHeader(400)
		s.log.Error(err, "error in parsing url", "url", u)

		_, err = fmt.Fprintln(w, err)
		s.log.Error(err, "error while sending error")
		return
	}

	r.URL.Host = grafanaUrl.Host
	r.URL.Path = path.Join(ss[3:]...)
	r.Host = grafanaUrl.Host

	proxy := httputil.NewSingleHostReverseProxy(grafanaUrl)
	proxy.ModifyResponse = func(r *http.Response) error {
		r.Header.Set("X-Frame-Options", "sameorigin")

		return nil
	}

	proxy.ServeHTTP(w, r)
}
