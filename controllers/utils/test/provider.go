package test

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/fx"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	ccfg "github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/go-logr/logr"
	ginkgoConfig "github.com/onsi/ginkgo/config"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
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

func NewTestOption(logger logr.Logger) *ctrl.Options {
	setupLog := logger.WithName("setup")
	scheme := runtime.NewScheme()

	clientgoscheme.AddToScheme(scheme)

	v1alpha1.AddToScheme(scheme)
	fmt.Println("Bind to port:", 9443+ginkgoConfig.GinkgoConfig.ParallelNode)
	options := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: ":" + fmt.Sprint(10080+ginkgoConfig.GinkgoConfig.ParallelNode),
		LeaderElection:     ccfg.ControllerCfg.EnableLeaderElection,
		Port:               9443 + ginkgoConfig.GinkgoConfig.ParallelNode,
	}

	if ccfg.ControllerCfg.ClusterScoped {
		setupLog.Info("Chaos controller manager is running in cluster scoped mode.")
		// will not specific a certain namespace
	} else {
		setupLog.Info("Chaos controller manager is running in namespace scoped mode.", "targetNamespace", ccfg.ControllerCfg.TargetNamespace)
		options.Namespace = ccfg.ControllerCfg.TargetNamespace
	}

	return &options
}
