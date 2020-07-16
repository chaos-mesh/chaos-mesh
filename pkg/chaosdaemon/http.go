// Copyright 2020 Chaos Mesh Authors.
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

package chaosdaemon

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type httpServerBuilder struct {
	mux       *http.ServeMux
	addr      string
	profiling bool
	reg       prometheus.Gatherer
}

func newHTTPServerBuilder() *httpServerBuilder {
	return &httpServerBuilder{
		mux: http.NewServeMux(),
	}
}

// Addr sets the addr of http server
func (b *httpServerBuilder) Addr(addr string) *httpServerBuilder {
	b.addr = addr

	return b
}

// Metrics sets the prometheus gatherer for http server
func (b *httpServerBuilder) Metrics(reg prometheus.Gatherer) *httpServerBuilder {
	b.reg = reg

	return b
}

// Profiling turns on or off profiling server of http server
func (b *httpServerBuilder) Profiling(profiling bool) *httpServerBuilder {
	b.profiling = profiling

	return b
}

// Build builds an http server
func (b *httpServerBuilder) Build() *http.Server {
	registerMetrics(b.mux, b.reg)

	if b.profiling {
		registerProfiler(b.mux)
	}

	return &http.Server{
		Addr:    b.addr,
		Handler: b.mux,
	}
}

func registerMetrics(mux *http.ServeMux, g prometheus.Gatherer) {
	if g != nil {
		mux.Handle("/metrics", promhttp.HandlerFor(g, promhttp.HandlerOpts{}))
	}
}

func registerProfiler(mux *http.ServeMux) {
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
