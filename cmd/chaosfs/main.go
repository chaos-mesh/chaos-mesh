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
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosfs"
	"github.com/chaos-mesh/chaos-mesh/pkg/pidfile"
	"github.com/chaos-mesh/chaos-mesh/pkg/version"

	"github.com/ethercflow/hookfs/hookfs"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var log = ctrl.Log.WithName("chaos-daemon")

var (
	addr         string
	pidFile      string
	original     string
	mountpoint   string
	printVersion bool

	once sync.Once
	pf   *pidfile.PIDFile
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.StringVar(&addr, "addr", ":65534", "The address to bind to")
	flag.StringVar(&pidFile, "pidfile", "", "PidFile")
	flag.StringVar(&original, "original", "", "ORIGINAL")
	flag.StringVar(&mountpoint, "mountpoint", "", "MOUNTPOINT")
	flag.BoolVar(&printVersion, "version", false, "print version information and exit")
}

func main() {
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	version.PrintVersionInfo("Chaosfs")
	if printVersion {
		os.Exit(0)
	}

	if err := checkFlags(); err != nil {
		log.Error(err, "Failed to check flags")
		os.Exit(1)
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-stopCh
		log.Info("Got signal to exit", "signal", sig)

		removePidFileOnce()

		umountTimeout := time.Second * 5
		err := RunUmountCmd(umountTimeout, mountpoint)
		if err != nil {
			log.Error(err, "Failed to umount file", "mountpoint", mountpoint)
		}
		os.Exit(0)
	}()

	log.Info("Init hookfs", "original", original, "mountpoint", mountpoint, "address", addr)
	fs, err := hookfs.NewHookFs(original, mountpoint, &chaosfs.InjuredHook{Addr: addr})
	if err != nil {
		log.Error(err, "Failed to init hookfs")
		os.Exit(1)
	}

	log.Info("Create pidFile", "pidFile", pidFile)
	pf, err = pidfile.New(pidFile)
	if err != nil {
		log.Error(err, "Failed to create pid file")
		os.Exit(1)
	}
	defer removePidFileOnce()

	log.Info("Starting chaosfs server...")
	if err = fs.Serve(); err != nil {
		log.Error(err, "Failed to start fuse server")
		os.Exit(1)
	}
}

func checkFlags() error {
	if original == "" {
		return errors.New("original is empty")
	}
	if mountpoint == "" {
		return errors.New("mountpoint is empty")
	}
	if pidFile == "" {
		return errors.New("pidFile is empty")
	}
	return nil
}

func removePidFileOnce() {
	once.Do(func() {
		if pf != nil {
			if err := pf.Remove(); err != nil {
				log.Error(err, "Failed to remove pid file", "pidFile", pidFile)
			}
		}
	})
}

func RunUmountCmd(timeout time.Duration, mountpoint string) error {
	err := runFuserMountCmd(timeout, mountpoint)
	if err != nil {
		err = runUmountCmd(timeout, mountpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func runFuserMountCmd(timeout time.Duration, mountpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return exec.CommandContext(ctx, "fusermount", "-u", mountpoint).Run()
}

func runUmountCmd(timeout time.Duration, mountpoint string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return exec.CommandContext(ctx, "umount", "-l", mountpoint).Run()
}
