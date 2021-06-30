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

package manager

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
)

func NewTestManager(lc fx.Lifecycle, options *ctrl.Options, cfg *rest.Config) (ctrl.Manager, error) {
	if ccfg.ControllerCfg.QPS > 0 {
		cfg.QPS = ccfg.ControllerCfg.QPS
	}
	if ccfg.ControllerCfg.Burst > 0 {
		cfg.Burst = ccfg.ControllerCfg.Burst
	}
	ch := make(chan struct{})

	mgr, err := ctrl.NewManager(cfg, *options)

	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			fmt.Println("Starting manager")
			go func() {
				if err := mgr.Start(ch); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Stopping manager")
			close(ch)
			return nil
		},
	})
	return mgr, nil
}
