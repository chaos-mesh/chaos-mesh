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

package common

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

type ChaosImplPair struct {
	Name   string
	Object InnerObjectWithSelector
	Impl   ChaosImpl
}

type ChaosImplPairs struct {
	fx.In

	Impls []*ChaosImplPair `group:"impl"`
}

func NewController(mgr ctrl.Manager, client client.Client, reader client.Reader, logger logr.Logger, selector *selector.Selector, pairs ChaosImplPairs) (types.Controller, error) {
	setupLog := logger.WithName("setup")
	for _, pair := range pairs.Impls {
		setupLog.Info("setting up controller", "resource-name", pair.Name)

		err := ctrl.NewControllerManagedBy(mgr).
			For(pair.Object).
			Named(pair.Name + "-records").
			Complete(&Reconciler{
				Impl:     pair.Impl,
				Object:   pair.Object,
				Client:   client,
				Reader:   reader,
				Selector: selector,
				Log:      logger,
			})
		if err != nil {
			return "", err
		}

	}

	return "records", nil
}
