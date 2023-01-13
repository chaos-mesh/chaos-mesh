// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package config

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/chaos-mesh/chaos-mesh/pkg/config"
)

// ControllerCfg is a global variable to keep the configuration for Chaos Controller
var ControllerCfg *config.ChaosControllerConfig

var log = ctrl.Log.WithName("config")

func init() {
	conf, err := config.EnvironChaosController()
	if err != nil {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
		log.Error(err, "Chaos Controller: invalid environment configuration")
		os.Exit(1)
	}

	err = validate(&conf)
	if err != nil {
		ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
		log.Error(err, "Chaos Controller: invalid configuration")
		os.Exit(1)
	}

	ControllerCfg = &conf
}

func validate(config *config.ChaosControllerConfig) error {

	if !config.ClusterScoped {
		if strings.TrimSpace(config.TargetNamespace) == "" {
			return errors.New("no target namespace specified with namespace scoped mode")
		}
	}

	return nil
}
