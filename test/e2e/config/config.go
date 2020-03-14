package config

import (
	"github.com/pingcap/chaos-mesh/test"
)

const (
	imagePullPolicyIfNotPresent = "IfNotPresent"
)

func NewDefaultOperatorConfig() test.OperatorConfig {
	return test.OperatorConfig{
		Namespace:   "chaos-testing",
		ReleaseName: "chaos-mesh",
		Tag:         "e2e",
		Manager: test.ManagerConfig{
			Image:           "localhost:5000/pingcap/chaos-mesh",
			Tag:             "latest",
			ImagePullPolicy: imagePullPolicyIfNotPresent,
		},
		Daemon: test.DaemonConfig{
			Image:           "localhost:5000/pingcap/chaos-daemon",
			Tag:             "latest",
			ImagePullPolicy: imagePullPolicyIfNotPresent,
			Runtime:         "containerd",
			SocketPath:      "/run/containerd/containerd.sock",
		},
	}
}
