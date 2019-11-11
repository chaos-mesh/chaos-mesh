package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/pingcap/chaos-operator/pkg/tcdaemon"
	"github.com/pingcap/chaos-operator/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("tcdaemon")

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

	if rawPort == "" {
		rawPort = os.Getenv("PORT")
	}

	if rawPort == "" {
		rawPort = "8080"
	}

	port, err := strconv.Atoi(rawPort)
	if err != nil {
		log.Error(err, "Error while parsing PORT environment variable", "port", rawPort)
	}
	tcdaemon.StartServer("0.0.0.0", port)
}
