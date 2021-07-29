// Copyright 2019 Chaos Mesh Authors.
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

package iochaos

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

	builder *podiochaosmanager.Builder
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	// The only possible phase to get in here is "Not Injected" or "Not Injected/Wait"

	impl.Log.Info("iochaos Apply", "namespace", obj.GetObjectMeta().Namespace, "name", obj.GetObjectMeta().Name)
	iochaos := obj.(*v1alpha1.IOChaos)
	if iochaos.Status.Instances == nil {
		iochaos.Status.Instances = make(map[string]int64)
	}

	record := records[index]
	phase := record.Phase

	if phase == waitForApplySync {
		podiochaos := &v1alpha1.PodIOChaos{}
		err := impl.Client.Get(ctx, controller.ParseNamespacedName(record.Id), podiochaos)
		if err != nil {
			if k8sError.IsNotFound(err) {
				return v1alpha1.NotInjected, nil
			}

			if k8sError.IsForbidden(err) {
				if strings.Contains(err.Error(), "because it is being terminated") {
					return v1alpha1.NotInjected, nil
				}
			}

			return waitForApplySync, err
		}

		if podiochaos.Status.FailedMessage != "" {
			return waitForApplySync, errors.New(podiochaos.Status.FailedMessage)
		}

		if podiochaos.Status.ObservedGeneration >= iochaos.Status.Instances[record.Id] {
			return v1alpha1.Injected, nil
		}

		return waitForApplySync, nil
	}

	podId, containerName := controller.ParseNamespacedNameContainer(records[index].Id)
	var pod v1.Pod
	err := impl.Client.Get(ctx, podId, &pod)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	source := iochaos.Namespace + "/" + iochaos.Name
	m := impl.builder.WithInit(source, types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	})

	m.T.SetVolumePath(iochaos.Spec.VolumePath)
	m.T.SetContainer(containerName)

	m.T.Append(v1alpha1.IOChaosAction{
		Type: iochaos.Spec.Action,
		Filter: v1alpha1.Filter{
			Path:    iochaos.Spec.Path,
			Percent: iochaos.Spec.Percent,
			Methods: iochaos.Spec.Methods,
		},
		Faults: []v1alpha1.IoFault{
			{
				Errno:  iochaos.Spec.Errno,
				Weight: 1,
			},
		},
		Latency:          iochaos.Spec.Delay,
		AttrOverrideSpec: iochaos.Spec.Attr,
		MistakeSpec:      iochaos.Spec.Mistake,
		Source:           m.Source,
	})
	generationNumber, err := m.Commit(ctx, iochaos)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	// modify the custom status
	iochaos.Status.Instances[record.Id] = generationNumber
	return waitForApplySync, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	// The only possible phase to get in here is "Injected" or "Injected/Wait"

	iochaos := obj.(*v1alpha1.IOChaos)
	if iochaos.Status.Instances == nil {
		iochaos.Status.Instances = make(map[string]int64)
	}

	record := records[index]
	phase := record.Phase
	if phase == waitForRecoverSync {
		podiochaos := &v1alpha1.PodIOChaos{}
		err := impl.Client.Get(ctx, controller.ParseNamespacedName(record.Id), podiochaos)
		if err != nil {
			// TODO: handle this error
			if k8sError.IsNotFound(err) {
				return v1alpha1.NotInjected, nil
			}
			return waitForRecoverSync, err
		}

		if podiochaos.Status.FailedMessage != "" {
			return waitForRecoverSync, errors.New(podiochaos.Status.FailedMessage)
		}

		if podiochaos.Status.ObservedGeneration >= iochaos.Status.Instances[record.Id] {
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

	source := iochaos.Namespace + "/" + iochaos.Name
	m := impl.builder.WithInit(source, types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	})

	generationNumber, err := m.Commit(ctx, iochaos)
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
	iochaos.Status.Instances[record.Id] = generationNumber
	return waitForRecoverSync, nil
}

func NewImpl(c client.Client, b *podiochaosmanager.Builder, log logr.Logger) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "iochaos",
		Object: &v1alpha1.IOChaos{},
		Impl: &Impl{
			Client:  c,
			Log:     log.WithName("iochaos"),
			builder: b,
		},
		ObjectList: &v1alpha1.IOChaosList{},
		Controlls:  []runtime.Object{&v1alpha1.PodIOChaos{}},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
	podiochaosmanager.NewBuilder,
)
