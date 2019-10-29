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
	rawPort      string
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	flag.StringVar(&rawPort, "port", "", "the port which server listens on")

	flag.Parse()
}

func main() {
	version.PrintVersionInfo()

	if printVersion {
		os.Exit(0)
	}

	logs.InitLogs()
	defer logs.FlushLogs()

	if rawPort == "" {
		rawPort := os.Getenv("PORT")
		if rawPort == "" {
			rawPort = "8080"
		}
	}

	port, err := strconv.Atoi(rawPort)
	if err != nil {
		glog.Fatalf("Error while parsing PORT environment variable: {}", rawPort)
	}
	tcdaemon.StartServer("0.0.0.0", port)
}
