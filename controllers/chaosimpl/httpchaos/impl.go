// Copyright 2020 Chaos Mesh Authors.
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

package httpchaos

import (
	"context"
	"errors"
	"strings"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/httpchaos/podhttpchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/iochaos/podiochaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
)

const (
	waitForApplySync   v1alpha1.Phase = "Not Injected/Wait"
	waitForRecoverSync v1alpha1.Phase = "Injected/Wait"
)

type Impl struct {
	client.Client
	Log logr.Logger

	builder *podhttpchaosmanager.Builder
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	// The only possible phase to get in here is "Not Injected" or "Not Injected/Wait"

	impl.Log.Info("httpchaos Apply", "namespace", obj.GetObjectMeta().Namespace, "name", obj.GetObjectMeta().Name)
	httpchaos := obj.(*v1alpha1.HTTPChaos)
	if httpchaos.Status.Instances == nil {
		httpchaos.Status.Instances = make(map[string]int64)
	}

	record := records[index]
	phase := record.Phase

	if phase == waitForApplySync {
		podhttpchaos := &v1alpha1.PodHttpChaos{}
		err := impl.Client.Get(ctx, controller.ParseNamespacedName(record.Id), podhttpchaos)
		if err != nil {
			return waitForApplySync, err
		}

		if podhttpchaos.Status.FailedMessage != "" {
			return waitForApplySync, errors.New(podhttpchaos.Status.FailedMessage)
		}

		if podhttpchaos.Status.ObservedGeneration >= httpchaos.Status.Instances[record.Id] {
			return v1alpha1.Injected, nil
		}

		return waitForApplySync, nil
	}

	podId, _ := controller.ParseNamespacedNameContainer(records[index].Id)
	var pod v1.Pod
	err := impl.Client.Get(ctx, podId, &pod)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	source := httpchaos.Namespace + "/" + httpchaos.Name
	m := impl.builder.WithInit(source, types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	})

	m.T.Append(v1alpha1.PodHttpChaosRule{
		Source: m.Source,
		Port:   httpchaos.Spec.Port,
		PodHttpChaosBaseRule: v1alpha1.PodHttpChaosBaseRule{
			Target: httpchaos.Spec.Target,
			Selector: v1alpha1.PodHttpChaosSelector{
				Port:            &httpchaos.Spec.Port,
				Path:            httpchaos.Spec.Path,
				Method:          httpchaos.Spec.Method,
				Code:            httpchaos.Spec.Code,
				RequestHeaders:  httpchaos.Spec.RequestHeaders,
				ResponseHeaders: httpchaos.Spec.ResponseHeaders,
			},
			Actions: httpchaos.Spec.PodHttpChaosActions,
		},
	})
	generationNumber, err := m.Commit(ctx)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	// modify the custom status
	httpchaos.Status.Instances[record.Id] = generationNumber
	return waitForApplySync, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	// The only possible phase to get in here is "Injected" or "Injected/Wait"

	httpchaos := obj.(*v1alpha1.HTTPChaos)
	if httpchaos.Status.Instances == nil {
		httpchaos.Status.Instances = make(map[string]int64)
	}

	record := records[index]
	phase := record.Phase
	if phase == waitForRecoverSync {
		podhttpchaos := &v1alpha1.PodHttpChaos{}
		err := impl.Client.Get(ctx, controller.ParseNamespacedName(record.Id), podhttpchaos)
		if err != nil {
			// TODO: handle this error
			if k8sError.IsNotFound(err) {
				return v1alpha1.NotInjected, nil
			}

			if k8sError.IsForbidden(err) {
				if strings.Contains(err.Error(), "because it is being terminated") {
					return v1alpha1.NotInjected, nil
				}
			}

			return waitForRecoverSync, err
		}

		if podhttpchaos.Status.FailedMessage != "" {
			return waitForRecoverSync, errors.New(podhttpchaos.Status.FailedMessage)
		}

		if podhttpchaos.Status.ObservedGeneration >= httpchaos.Status.Instances[record.Id] {
			return v1alpha1.NotInjected, nil
		}

		return waitForRecoverSync, nil
	}

	podId, _ := controller.ParseNamespacedNameContainer(records[index].Id)
	var pod v1.Pod
	err := impl.Client.Get(ctx, podId, &pod)
	if err != nil {
		// TODO: handle this error
		if k8sError.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.NotInjected, err
	}

	source := httpchaos.Namespace + "/" + httpchaos.Name
	m := impl.builder.WithInit(source, types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	})

	generationNumber, err := m.Commit(ctx)
	if err != nil {
		if err == podiochaosmanager.ErrPodNotFound || err == podiochaosmanager.ErrPodNotRunning {
			return v1alpha1.NotInjected, nil
		}

		if k8sError.IsForbidden(err) {
			if strings.Contains(err.Error(), "because it is being terminated") {
				return v1alpha1.NotInjected, nil
			}
		}
		return v1alpha1.Injected, err
	}

	// Now modify the custom status and phase
	httpchaos.Status.Instances[record.Id] = generationNumber
	return waitForRecoverSync, nil
}

func NewImpl(c client.Client, b *podhttpchaosmanager.Builder, log logr.Logger) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "httpchaos",
		Object: &v1alpha1.HTTPChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("httpchaos"),
			builder: b,
		},
		ObjectList: &v1alpha1.HTTPChaosList{},
		Controlls:  []runtime.Object{&v1alpha1.PodHttpChaos{}},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	podhttpchaosmanager.NewBuilder,
)
