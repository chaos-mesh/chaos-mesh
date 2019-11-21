package main

import (
	"flag"
	"os"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/pingcap/chaos-operator/pkg/chaosdaemon"
	"github.com/pingcap/chaos-operator/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
)

var log = ctrl.Log.WithName("chaos-daemon")

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

	ctrl.SetLogger(zap.Logger(true))

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
	log.Info("starting server")
	chaosdaemon.StartServer("0.0.0.0", port)
}
