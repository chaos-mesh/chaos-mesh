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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	chaosmeshapi "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type reInjector struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
	logger logr.Logger
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reInjector{}

// Reconcile will re-inject chaos in restart pods
func (r *reInjector) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	podKey := fmt.Sprintf("%s/%s", request.Namespace, request.Name)
	// find the chaosKey in podToChaosInfoMap
	chaosToContainerMapValue, ok := podToChaosInfoMap.Load(podKey)
	if !ok {
		r.logger.Info("pod not found in podToChaosInfoMap", "podKey", podKey)
		return reconcile.Result{}, nil
	}

	chaosToContainerMap := chaosToContainerMapValue.(map[string]string)
	for chaosKey := range chaosToContainerMap {
		if err := r.reInjectChaos(ctx, chaosKey, r.logger); err != nil {
			r.logger.Error(err, "reinject chaos failed", chaosKey)
			continue
		}
	}
	return reconcile.Result{}, nil
}

func (r *reInjector) reInjectChaos(ctx context.Context, chaosKey string, logger logr.Logger) error {
	logger.Info("start to re-inject chaos")
	chaosInfo, err := parseChaosKey(chaosKey)
	if err != nil {
		logger.Error(err, "parse chaosKey failed", "chaosKey", chaosKey)
		return err
	}
	kind, namespace, name := chaosInfo.Kind, chaosInfo.Namespace, chaosInfo.Name
	chaosKind, exists := chaosmeshapi.AllKinds()[kind]
	if !exists {
		logger.Info("chaosKind not found", "chaosKind", kind)
		return nil
	}
	chaos := chaosKind.SpawnObject()
	namespacedName := types.NamespacedName{Namespace: namespace, Name: name}
	if err = r.client.Get(ctx, namespacedName, chaos); err != nil {
		return err
	}

	if err = r.pauseChaos(ctx, chaos); err != nil {
		return err
	}

	timeoutTicker := time.NewTicker(30 * time.Second)
	durationTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-timeoutTicker.C:
			return nil
		case <-durationTicker.C:
			if err = r.client.Get(ctx, namespacedName, chaos); err != nil {
				return err
			}
			statefulChaos := chaos.(chaosmeshapi.StatefulObject)
			allNotInjected := true
			for _, record := range statefulChaos.GetStatus().Experiment.Records {
				if record.Phase != chaosmeshapi.NotInjected {
					allNotInjected = false
				}
			}
			if allNotInjected {
				if err = r.unPauseChaos(ctx, chaos); err != nil {
					return err
				}
				logger.Info("re-inject chaos successfully", "chaosKey", chaosKey)
				return nil
			}
		}
	}
}

func (r *reInjector) pauseChaos(ctx context.Context, chaos client.Object) error {
	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{chaosmeshapi.PauseAnnotationKey: "true"},
		},
	})
	return r.client.Patch(ctx, chaos, client.RawPatch(types.MergePatchType, mergePatch))
}

func (r *reInjector) unPauseChaos(ctx context.Context, chaos client.Object) error {
	var mergePatch []byte
	mergePatch, _ = json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{chaosmeshapi.PauseAnnotationKey: "false"},
		},
	})
	return r.client.Patch(ctx, chaos, client.RawPatch(types.MergePatchType, mergePatch))
}
