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

package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/signals"
	"github.com/pingcap/chaos-operator/pkg/version"
	"github.com/pingcap/chaos-operator/pkg/webhook"

	"golang.org/x/sync/errgroup"

	"k8s.io/apiserver/pkg/util/logs"
)

var (
	printVersion bool
	parameters   = webhook.Parameters{}
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.StringVar(&parameters.Addr, "addr", ":9443",
		"address to serve on")
	flag.StringVar(&parameters.CertFile, "tls-cert-file", "/etc/webhook/certs/cert.pem",
		"file containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.KeyFile, "tls-key-file", "/etc/webhook/certs/cert.key",
		"file containing the x509 private key to --tls-cert-file.")
	flag.StringVar(&parameters.ConfigDirectory, "config-directory", "/etc/webhook/conf/",
		"config directory (will load all .yaml files in this directory)")
	flag.StringVar(&parameters.AnnotationNamespace, "annotation-namespace", "admission-webhook.pingcap.com",
		"Override the AnnotationNamespace")

	flag.Parse()
}

func main() {
	if printVersion {
		version.PrintVersionInfo()
		os.Exit(0)
	}

	version.LogVersionInfo()

	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := signals.SetupSignalHandler()

	ws := webhook.NewWebHookServer(parameters)

	g := errgroup.Group{}

	g.Go(func() error {
		return ws.Run(stopCh)
	})

	if err := g.Wait(); err != nil {
		glog.Fatal(err)
	}
}
