package reinjection

import (
	"context"

	chaosmeshapi "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/go-logr/logr"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/flowcontrol"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ControllerConfig is the config of controller
type ControllerConfig struct {
	ChaosKinds []string `json:"chaosKinds" yaml:"chaosKinds"`
}

var (
	scheme               = runtime.NewScheme()
	controllerConfigPath string
)

const (
	relationshipKey              = "relationship"
	podToChaosConfigMapName      = "pod-to-chaos"
	podToChaosConfigMapNamespace = "chaos-mesh"
)

func Bootstrap(mgr ctrl.Manager, client client.Client, logger logr.Logger, b *chaosdaemon.ChaosDaemonClientBuilder) error {
	if !config.ShouldSpawnController("reinjection") {
		return nil
	}

	ctx := context.Background()
	log.Info("start to load config")
	controllerConfig := &ControllerConfig{ChaosKinds: []string{"StressChaos"}}
	//if err := readConfigFile(controllerConfig, controllerConfigPath); err != nil {
	//	log.Errorf("failed to read config file, err: %v", err)
	//	return err
	//}

	log.Info("setting up manager")
	cfg := ctrl.GetConfigOrDie()
	cfg.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(300, 300)
	mgr, err := manager.New(cfg, manager.Options{Scheme: scheme})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		return err
	}

	// Set up a new controller to reconcile ReplicaSets
	log.Infof("Setting up controller to watch these kind of chaos %v", controllerConfig.ChaosKinds)
	c, err := controller.New("reinject-controller", mgr, controller.Options{
		Reconciler: &reInjector{client: mgr.GetClient()},
	})
	if err != nil {
		log.Error(err, "unable to set up individual controller")
		return err

	}

	// Watch Pods and enqueue the Pod with restart count increased
	if err := c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{},
		buildPodPredicate(ctx, mgr.GetClient())); err != nil {
		log.Error(err, "unable to watch Pods")
		return err
	}

	// Watch ChaosTypes to maintain the podToChaosInfoMap
	// for update event, we need to update the podToChaosInfoMap from the chaos.status.record
	// for delete event, we should delete the chaosKey from the podToChaosInfoMap
	for _, chaosKindString := range controllerConfig.ChaosKinds {
		chaosKind, exists := chaosmeshapi.AllKinds()[chaosKindString]
		if !exists {
			log.Infof("chaosKind %s not found", chaosKindString)
			return err
		}
		chaos := chaosKind.SpawnObject()
		if err := c.Watch(&source.Kind{Type: chaos}, &handler.EnqueueRequestForObject{},
			buildChaosPredicate(ctx, mgr.GetClient(), chaosKindString)); err != nil {
			log.Error(err, "unable to watch chaosKind %s", chaosKindString)
			return err
		}
	}

	if err := createPodToChaosConfigMap(ctx, mgr.GetClient()); err != nil {
		log.Error(err, "unable to create podToChaosConfigMap")
		return err
	}

	log.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to run manager")
		return err
	}
	return nil
}
