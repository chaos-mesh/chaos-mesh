package chaosdaemon

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func newHTTPServer(addr string, reg prometheus.Gatherer) *http.Server {
	mux := http.NewServeMux()
	registerMetrics(mux, reg)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

func registerMetrics(mux *http.ServeMux, g prometheus.Gatherer) {
	if g != nil {
		mux.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
	}
}
