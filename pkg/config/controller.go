package config

import (
	"github.com/kelseyhightower/envconfig"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
)

var controllerLog = ctrl.Log.WithName("conf")

//ControllerCfg is a global variable to keep the configuration for Chaos Controller
var ControllerCfg *ChaosControllerConfig

func init() {
	conf, err := EnvironChaosController()
	if err != nil {
		controllerLog.Error(err, "Chaos Controller: invalid environment configuration")
		os.Exit(1)
	}
	ControllerCfg = &conf
}

// ChaosControllerConfig defines the configuration for Chaos Controller
type ChaosControllerConfig struct {
	ChaosDaemonPort int `envconfig:"CHAOS_DAEMON_PORT" default:"31767"`
	BPFKIPort       int `envconfig:"BPFKI_PORT" default:"50051"`
}

// EnvironChaosController returns the settings from the environment.
func EnvironChaosController() (ChaosControllerConfig, error) {
	cfg := ChaosControllerConfig{}
	err := envconfig.Process("", &cfg)
	return cfg, err
}

