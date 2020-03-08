package main

import (
	"flag"
	"math/rand"
	"time"

	"github.com/pingcap/chaos-mesh/pkg/chaoscm"

	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	addr         string
	printVersion bool
)

var log = ctrl.Log.WithName("chaoscm")

func initFlags() {
	flag.StringVar(&addr, "addr", ":65533", "The address to bind to")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
}

func main() {
	initFlags()
	log.Info("Starting chaoscm server ...")
	chaoscm.StartServer(addr)
}
