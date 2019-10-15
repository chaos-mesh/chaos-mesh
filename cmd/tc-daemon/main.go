package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/pingcap/chaos-operator/pkg/tcdaemon"
	"github.com/pingcap/chaos-operator/pkg/version"
	"k8s.io/apiserver/pkg/util/logs"
)

var (
	printVersion bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo()

	if printVersion {
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	raw_port := os.Getenv("PORT")
	if raw_port == "" {
		raw_port = "8080"
	}

	port, err := strconv.Atoi(raw_port)
	if err != nil {
		glog.Errorf("Error while parsing PORT environment variable: {}", raw_port)
	}
	tcdaemon.StartServer("0.0.0.0", port)
}
