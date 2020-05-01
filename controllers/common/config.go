package common

import "os"

var Cfg *Config

func init() {
	Cfg = NewConfig()
}

type Config struct {
	ChaosDaemonPort string
	BPFKIPort       string
}

func NewConfig() *Config {
	return &Config{
		ChaosDaemonPort: os.Getenv("CHAOS_DAEMON_PORT"),
		BPFKIPort:       os.Getenv("BPFKI_PORT"),
	}
}
