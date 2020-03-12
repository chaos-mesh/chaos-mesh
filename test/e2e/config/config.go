package config

import (
	"flag"
	"github.com/pingcap/chaos-mesh/test"
)

var TestConfig *test.Config = test.NewDefaultConfig()

func RegisterChaosMeshConfig(flags *flag.FlagSet) {
	flags.StringVar(&TestConfig.ChartDir, "chart-dir", "/charts", "chart dir")
}
