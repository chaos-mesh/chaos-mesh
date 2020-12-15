// Copyright 2019 Chaos Mesh Authors.
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
	"context"
	"errors"
	"flag"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/ethercflow/hookfs/hookfs"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosfs"
	"github.com/chaos-mesh/chaos-mesh/pkg/pidfile"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	addr         string
	pidFile      string
	original     string
	mountpoint   string
	printVersion bool

	pf *pidfile.PIDFile
)

var log = ctrl.Log.WithName("chaos-daemon")

func initFlag() {
	flag.StringVar(&addr, "addr", ":65534", "The address to bind to")
	flag.StringVar(&pidFile, "pidfile", "", "PidFile")
	flag.StringVar(&original, "original", "", "ORIGINAL")
	flag.StringVar(&mountpoint, "mountpoint", "", "MOUNTPOINT")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")

	rand.Seed(time.Now().UnixNano())
	flag.Parse()
}

func main() {
	initFlag()

	if err := checkFlag(); err != nil {
		log.Error(err, "invalid flag")
		os.Exit(1)
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	version.PrintVersionInfo("Chaosfs")
	if printVersion {
		os.Exit(0)
	}

	stopCh := ctrl.SetupSignalHandler()

	go func() {
		sig := <-stopCh
		log.Info("Got signal to exit", "signal", sig)

		if pf != nil {
			if err := pf.Remove(); err != nil {
				log.Error(err, "failed to remove pid file")
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := exec.CommandContext(ctx, "fusermount", "-u", mountpoint).Run()
		if err != nil {
			if err1 := exec.CommandContext(ctx, "umount", "-l", mountpoint).Run(); err1 != nil {
				log.Error(err, "failed to fusermount", "umount failed", err1)
			}
			log.Error(err, "failed to fusermount")
		}
		os.Exit(0)
	}()

	log.Info("Init hookfs")
	fs, err := hookfs.NewHookFs(original, mountpoint, &chaosfs.InjuredHook{Addr: addr})
	if err != nil {
		log.Error(err, "failed to init hookfs")
	}

	pf, err := pidfile.New(pidFile)
	if err != nil {
		log.Error(err, "failed to create pid file")
		os.Exit(1)
	}

	defer func() {
		if err := pf.Remove(); err != nil {
			log.Error(err, "failed to remove pid file")
		}
	}()

	log.Info("Starting chaosfs server...")
	if err = fs.Serve(); err != nil {
		log.Error(err, "failed to start fuse server")
		os.Exit(1)
	}
}

func checkFlag() error {
	if original == "" || mountpoint == "" {
		return errors.New("invalid original or mountpoint")
	}

	if pidFile == "" {
		return errors.New("invalid pid file")
	}

	return nil
}
