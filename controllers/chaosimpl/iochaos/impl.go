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

	"go.uber.org/fx"

	"github.com/chaos-mesh/chaos-mesh/controllers/common"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/iochaos/podiochaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Impl struct {
	client.Client
	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	podId, containerName := controller.ParseNamespacedNameContainer(records[index].Id)
	var pod v1.Pod
	err := impl.Client.Get(ctx, podId, &pod)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	iochaos := obj.(*v1alpha1.IoChaos)

	source := iochaos.Namespace + "/" + iochaos.Name
	m := podiochaosmanager.WithInit(source, impl.Log, impl.Client, types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	})

	m.T.SetVolumePath(iochaos.Spec.VolumePath)
	m.T.SetContainer(containerName)

	m.T.Append(v1alpha1.IoChaosAction{
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
		Source:           m.Source,
	})
	err = m.Commit(ctx)
	if err != nil {
		return v1alpha1.NotInjected, err
	}

	return v1alpha1.Injected, nil
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
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

	iochaos := obj.(*v1alpha1.IoChaos)

	source := iochaos.Namespace + "/" + iochaos.Name
	m := podiochaosmanager.WithInit(source, impl.Log, impl.Client, types.NamespacedName{
		Name:      pod.Name,
		Namespace: pod.Namespace,
	})

	err = m.Commit(ctx)
	if err != nil {
		if err == podnetworkchaosmanager.ErrPodNotFound || err == podnetworkchaosmanager.ErrPodNotRunning {
			return v1alpha1.NotInjected, err
		}
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func NewImpl(c client.Client, log logr.Logger) *common.ChaosImplPair {
	return &common.ChaosImplPair{
		Name:   "iochaos",
		Object: &v1alpha1.IoChaos{},
		Impl: &Impl{
			Client: c,
			Log:    log.WithName("iochaos"),
		},
	}
}

var Module = fx.Provide(
	fx.Annotated{
		Group:  "impl",
		Target: NewImpl,
	},
)
