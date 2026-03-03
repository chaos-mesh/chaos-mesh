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
	"flag"
	stdlog "log"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/crclients"
	"github.com/chaos-mesh/chaos-mesh/pkg/fusedev"
	"github.com/chaos-mesh/chaos-mesh/pkg/log"
	"github.com/chaos-mesh/chaos-mesh/pkg/sysfs"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"
)

var (
	conf = &chaosdaemon.Config{Host: "0.0.0.0", CrClientConfig: &crclients.CrClientConfig{}}

	printVersion bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.IntVar(&conf.GRPCPort, "grpc-port", 31767, "the port which grpc server listens on")
	flag.IntVar(&conf.HTTPPort, "http-port", 31766, "the port which http server listens on")
	flag.StringVar(&conf.CrClientConfig.Runtime, "runtime", "docker", "current container runtime")
	flag.StringVar(&conf.CrClientConfig.SocketPath, "runtime-socket-path", "", "current container runtime socket path")
	flag.StringVar(&conf.CrClientConfig.ContainerdNS, "containerd-ns", "k8s.io", "namespace used for containerd")
	flag.StringVar(&conf.CaCert, "ca", "", "ca certificate of grpc server")
	flag.StringVar(&conf.Cert, "cert", "", "certificate of grpc server")
	flag.StringVar(&conf.Key, "key", "", "key of grpc server")
	flag.BoolVar(&conf.Profiling, "pprof", false, "enable pprof")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo("Chaos-daemon")

	if printVersion {
		os.Exit(0)
	}

	rootLogger, err := log.NewDefaultZapLogger()
	if err != nil {
		stdlog.Fatal("failed to create root logger", err)
	}
	rootLogger = rootLogger.WithName("chaos-daemon.daemon-server")
	log.ReplaceGlobals(rootLogger)
	ctrl.SetLogger(rootLogger)

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		// Use collectors as prometheus functions deprecated
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	rootLogger.Info("grant access to /dev/fuse")
	err = fusedev.GrantAccess()
	if err != nil {
		rootLogger.Error(err, "grant access to /dev/fuse")
	}

	rootLogger.Info("remount /sys with read-write permission")
	err = sysfs.RemountWithOption()
	if err != nil {
		rootLogger.Error(err, "remount /sys with read-write permission")
	}

	server, err := chaosdaemon.BuildServer(conf, reg, rootLogger)
	if err != nil {
		rootLogger.Error(err, "build chaos-daemon server")
		os.Exit(1)
	}

	errs := make(chan error)
	go func() {
		errs <- server.Start()
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGINT,
		syscall.SIGTERM)

	select {
	case sig := <-sigc:
		rootLogger.Info("received signal", "signal", sig)
	case err = <-errs:
		if err != nil {
			rootLogger.Error(err, "chaos-daemon runtime")
		}
	}
	if err = server.Shutdown(); err != nil {
		rootLogger.Error(err, "chaos-daemon shutdown")
	}
}
