package config

import (
	"flag"
	"github.com/pingcap/chaos-mesh/test"
)

var TestConfig *test.Config = test.NewDefaultConfig()

const (
	imagePullPolicyIfNotPresent = "IfNotPresent"
)

func RegisterChaosMeshConfig(flags *flag.FlagSet) {
	flags.StringVar(&TestConfig.ChartDir, "chart-dir", "/charts", "chart dir")
}

func NewDefaultOperatorConfig() test.OperatorConfig {
	return test.OperatorConfig{
		Namespace:   "chaos-testing",
		ReleaseName: "chaos-mesh",
		Tag:         "e2e",
		Manager: test.ManagerConfig{
			Image:           "localhost:5000:pingcap/chaos-mesh",
			Tag:             "latest",
			ImagePullPolicy: imagePullPolicyIfNotPresent,
		},
		Daemon: test.DaemonConfig{
			Image:           "localhost:5000:pingcap/chaos-daemon",
			Tag:             "latest",
			ImagePullPolicy: imagePullPolicyIfNotPresent,
			Runtime:         "containerd",
			SocketPath:      "/run/containerd/containerd.sock",
		},
	}
}
