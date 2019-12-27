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
