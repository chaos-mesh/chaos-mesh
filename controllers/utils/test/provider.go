package test

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
