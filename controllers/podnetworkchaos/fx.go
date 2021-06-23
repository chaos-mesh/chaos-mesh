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
	"reflect"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

func NewController(mgr ctrl.Manager, client client.Client, logger logr.Logger, b *chaosdaemon.ChaosDaemonClientBuilder, recorderBuilder *recorder.RecorderBuilder) (types.Controller, error) {
	err := builder.Default(mgr).
		For(&v1alpha1.PodNetworkChaos{}).
		Named("podnetworkchaos").
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldObj := e.ObjectOld.(*v1alpha1.PodNetworkChaos)
				newObj := e.ObjectNew.(*v1alpha1.PodNetworkChaos)

				return !reflect.DeepEqual(oldObj.Spec, newObj.Spec)
			},
		}).
		Complete(&Reconciler{
			Client:   client,
			Log:      logger.WithName("podnetworkchaos"),
			Recorder: recorderBuilder.Build("podnetworkchaos"),

			// TODO:
			AllowHostNetworkTesting:  config.ControllerCfg.AllowHostNetworkTesting,
			ChaosDaemonClientBuilder: b,
		})
	if err != nil {
		return "", err
	}

	return "podnetworkchaos", nil
}
