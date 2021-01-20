package common

import (
	"flag"
	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
)

type LoggerFlushFunc func()

func NewStderrLogger() (logr.Logger, LoggerFlushFunc, error) {
	klog.InitFlags(flag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	err := flag.Set("logtostderr", "true")
	if err != nil {
		return nil, nil, err
	}

	logger := klogr.New()
	return logger, klog.Flush, nil
}
