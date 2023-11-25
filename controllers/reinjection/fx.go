// Copyright 2023 Chaos Mesh Authors.
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

package reinjection

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"

	chaosmeshapi "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
)

// ControllerConfig is the config of controller
type ControllerConfig struct {
	ChaosKinds []string `json:"chaosKinds" yaml:"chaosKinds"`
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	// register chaosmesh api, os that we can watch chaosmesh cr
	_ = chaosmeshapi.AddToScheme(scheme)
}

// the struct of this map is map[podKey][chaosKey]containerName
var podToChaosInfoMap = sync.Map{}

func Bootstrap(mgr ctrl.Manager, client client.Client, logger logr.Logger, b *chaosdaemon.ChaosDaemonClientBuilder) error {
	if !config.ShouldSpawnController(config.ReInjectControllerName) {
		return nil
	}

	logger.Info("start to load config")
	controllerConfig := &ControllerConfig{
		ChaosKinds: config.ControllerCfg.ReInjectChaosKinds,
	}

	// Set up a new controller to reconcile ReplicaSets
	logger.Info("Setting up controller to watch these kind of chaos", "chaosKinds", controllerConfig.ChaosKinds)
	c, err := controller.New("reinject-controller", mgr, controller.Options{
		Reconciler: &reInjector{client: mgr.GetClient(), logger: logger},
	})
	if err != nil {
		logger.Error(err, "unable to set up individual controller")
		return err
	}

	// Watch Pods and enqueue the Pod with restart count increased
	if err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{},
		buildPodPredicate()); err != nil {
		logger.Error(err, "unable to watch Pods")
		return err
	}

	// Watch ChaosTypes to maintain the podToChaosInfoMap
	// for update event, we need to update the podToChaosInfoMap from the chaos.status.record
	// for delete event, we should delete the chaosKey from the podToChaosInfoMap
	for _, chaosKindString := range controllerConfig.ChaosKinds {
		chaosKind, exists := chaosmeshapi.AllKinds()[chaosKindString]
		if !exists {
			logger.Info("chaosKind not found", chaosKindString)
			return err
		}
		chaos := chaosKind.SpawnObject()
		if err := c.Watch(&source.Kind{Type: chaos}, &handler.EnqueueRequestForObject{},
			buildChaosPredicate(chaosKindString)); err != nil {
			logger.Error(err, "unable to watch chaosKind", chaosKindString)
			return err
		}
	}

	err = mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		logger.Info("start to run the syncChaos")
		if syncErr := syncChaos(mgr.GetClient(), controllerConfig); syncErr != nil {
			logger.Error(err, "unable to sync chaos")
		}
		<-ctx.Done()
		return nil
	}))
	if err != nil {
		logger.Error(err, "unable to add runnable function")
		return err
	}

	return nil
}
