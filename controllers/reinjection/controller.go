package reinjection

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	chaosmeshapi "github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type reInjector struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reInjector{}

// Reconcile will re-inject chaos in restart pods
func (r *reInjector) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	podKey := fmt.Sprintf("%s/%s", request.Namespace, request.Name)
	logger := log.WithField("pod", podKey)
	// find the chaosKey in podToChaosInfoMap
	chaosToContainerMap, err := readRelationship(ctx, r.client, podKey)
	if err != nil || chaosToContainerMap == nil {
		logger.Infof("pod %s not found in podToChaosInfoMap", podKey)
		return reconcile.Result{}, nil
	}

	for chaosKey, _ := range chaosToContainerMap {
		if err := r.reInjectChaos(ctx, chaosKey, logger); err != nil {
			logger.Errorf("reinject chaos %s failed, err: %s", chaosKey, err.Error())
			continue
		}
	}
	return reconcile.Result{}, nil
}

func (r *reInjector) reInjectChaos(ctx context.Context, chaosKey string, logger *log.Entry) error {
	reInjectLog := logger.WithField("chaosKey", chaosKey)
	reInjectLog.Info("start to re-inject chaos")
	chaosInfo, err := parseChaosKey(chaosKey)
	if err != nil {
		reInjectLog.Errorf("parse chaosKey %s failed, err: %s", chaosKey, err.Error())
		return err
	}
	kind, namespace, name := chaosInfo.Kind, chaosInfo.Namespace, chaosInfo.Name
	chaosKind, exists := chaosmeshapi.AllKinds()[kind]
	if !exists {
		reInjectLog.Infof("chaosKind %s not found", kind)
		return nil
	}
	chaos := chaosKind.SpawnObject()
	namespacedName := types.NamespacedName{Namespace: namespace, Name: name}
	if err = r.client.Get(ctx, namespacedName, chaos); err != nil {
		return err
	}

	reInjectLog.Infof("start to pause chaos %s/%s", chaos.GetNamespace(), chaos.GetName())
	if err = r.pauseChaos(ctx, chaos); err != nil {
		return err
	}

	timeoutTicker := time.NewTicker(30 * time.Second)
	durationTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-timeoutTicker.C:
			reInjectLog.Infof("timeout to pause chaos %s/%s", chaos.GetNamespace(), chaos.GetName())
			return nil
		case <-durationTicker.C:
			reInjectLog.Infof("start to check chaos %s/%s", chaos.GetNamespace(), chaos.GetName())
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
				reInjectLog.Infof("start to unpause chaos %s/%s", chaos.GetNamespace(), chaos.GetName())
				if err = r.unPauseChaos(ctx, chaos); err != nil {
					return err
				}
				reInjectLog.Infof("unpause chaos %s/%s successfully", chaos.GetNamespace(), chaos.GetName())
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
