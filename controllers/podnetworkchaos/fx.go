// Copyright 2021 Chaos Mesh Authors.
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

package podnetworkchaos

import (
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/types"
)

func NewController(mgr ctrl.Manager, client client.Client, reader client.Reader, logger logr.Logger) (types.Controller, error) {
	err := ctrl.NewControllerManagedBy(mgr).
		For(obj.Object).
		Named(obj.Name + "-podnetworkchaos").
		Complete(&Reconciler{
			Client:   client,
			Reader:   reader,
			Log:      logger.WithName("podnetworkchaos"),
			Recorder: mgr.GetEventRecorderFor("podnetworkchaos"),

			// TODO:
			AllowHostNetworkTesting: config.ControllerCfg.AllowHostNetworkTesting,
		})
	if err != nil {
		return "", err
	}

	return "podnetworkchaos", nil
}
