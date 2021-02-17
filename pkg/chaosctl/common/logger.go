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

package common

import (
	"flag"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
)

type LoggerFlushFunc func()

func SetupKlog() error {
	// setup klog
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	return flag.Set("logtostderr", "true")
}

func NewStderrLogger() (logr.Logger, LoggerFlushFunc, error) {
	logger := klogr.New()
	return logger, klog.Flush, nil
}

var globalLogger logr.Logger

func SetupGlobalLogger(logger logr.Logger) {
	globalLogger = logger
}

func L() logr.Logger {
	if globalLogger == nil {
		panic("global logger not initialized")
	}
	return globalLogger
}
