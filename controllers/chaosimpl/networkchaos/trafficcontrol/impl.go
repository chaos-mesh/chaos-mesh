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

package trafficcontrol

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/networkchaos/podnetworkchaosmanager"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
)

const (
	networkTcActionMsg    = "network traffic control action duration %s"
	networkChaosSourceMsg = "This is a source pod."
	networkChaosTargetMsg = "This is a target pod."

	targetIPSetPostFix = "tgt"
	sourceIPSetPostFix = "src"
)

type Impl struct {
	client.Client

	Log logr.Logger
}

func (impl *Impl) Apply(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	impl.Log.Info("traffic control Apply", "namespace", obj.GetObjectMeta().Namespace, "name", obj.GetObjectMeta().Name)
	networkchaos, ok := obj.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		impl.Log.Error(err, "chaos is not NetworkChaos", "chaos", obj)
		return v1alpha1.NotInjected, err
	}

	var pod v1.Pod
	record := records[index]
	err := impl.Get(ctx, controller.ParseNamespacedName(record.Id), &pod)
	if err != nil {
		// TODO: handle this error
		return v1alpha1.NotInjected, err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkchaosmanager.WithInit(source, impl.Log, impl.Client, types.NamespacedName{
		pod.Namespace,
		pod.Name,
	})

	if record.SelectorKey == "." {
		if networkchaos.Spec.Direction == v1alpha1.To || networkchaos.Spec.Direction == v1alpha1.Both {
			var targets []*v1alpha1.Record
			for _, record := range records {
				if record.SelectorKey == ".Target" {
					targets = append(targets, record)
				}
			}

			err := impl.ApplyTc(ctx, m, targets, networkchaos, targetIPSetPostFix)
			if err != nil {
				return v1alpha1.NotInjected, err
			}

			err = m.Commit(ctx)
			if err != nil {
				return v1alpha1.NotInjected, err
			}
		}

		return v1alpha1.Injected, nil
	} else if record.SelectorKey == ".Target" {
		if networkchaos.Spec.Direction == v1alpha1.From || networkchaos.Spec.Direction == v1alpha1.Both {
			var targets []*v1alpha1.Record
			for _, record := range records {
				if record.SelectorKey == "." {
					targets = append(targets, record)
				}
			}

			err := impl.ApplyTc(ctx, m, targets, networkchaos, sourceIPSetPostFix)
			if err != nil {
				return v1alpha1.NotInjected, err
			}

			err = m.Commit(ctx)
			if err != nil {
				return v1alpha1.NotInjected, err
			}
		}

		return v1alpha1.Injected, nil
	} else {
		impl.Log.Info("unknown selector key", "record", record)
		return v1alpha1.NotInjected, nil
	}
}

func (impl *Impl) Recover(ctx context.Context, index int, records []*v1alpha1.Record, obj v1alpha1.InnerObject) (v1alpha1.Phase, error) {
	networkchaos, ok := obj.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		impl.Log.Error(err, "chaos is not NetworkChaos", "chaos", obj)
		return v1alpha1.Injected, err
	}

	var pod v1.Pod
	record := records[index]
	err := impl.Get(ctx, controller.ParseNamespacedName(record.Id), &pod)
	if err != nil {
		// TODO: handle this error
		if k8sError.IsNotFound(err) {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	source := networkchaos.Namespace + "/" + networkchaos.Name
	m := podnetworkchaosmanager.WithInit(source, impl.Log, impl.Client, types.NamespacedName{
		pod.Namespace,
		pod.Name,
	})
	err = m.Commit(ctx)
	if err != nil {
		if err == podnetworkchaosmanager.ErrPodNotFound || err == podnetworkchaosmanager.ErrPodNotRunning {
			return v1alpha1.NotInjected, nil
		}
		return v1alpha1.Injected, err
	}

	return v1alpha1.NotInjected, nil
}

func (impl *Impl) ApplyTc(ctx context.Context, m *podnetworkchaosmanager.PodNetworkManager, targets []*v1alpha1.Record, networkchaos *v1alpha1.NetworkChaos, ipSetPostFix string) error {
	spec := networkchaos.Spec
	tcType := v1alpha1.Bandwidth
	switch spec.Action {
	case v1alpha1.NetemAction, v1alpha1.DelayAction, v1alpha1.DuplicateAction, v1alpha1.CorruptAction, v1alpha1.LossAction:
		tcType = v1alpha1.Netem
	case v1alpha1.BandwidthAction:
		tcType = v1alpha1.Bandwidth
	default:
		return fmt.Errorf("unknown action %s", spec.Action)
	}

	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		return err
	}

	if len(targets)+len(externalCidrs) == 0 {
		impl.Log.Info("apply traffic control", "sources", m.Source)
		m.T.Append(v1alpha1.RawTrafficControl{
			Type:        tcType,
			TcParameter: spec.TcParameter,
			Source:      m.Source,
		})
		return nil
	}

	targetPods := []v1.Pod{}
	for _, record := range targets {
		var pod v1.Pod
		err := impl.Get(ctx, controller.ParseNamespacedName(record.Id), &pod)
		if err != nil {
			// TODO: handle this error
			return err
		}
		targetPods = append(targetPods, pod)
	}
	dstIpset := ipset.BuildIPSet(targetPods, externalCidrs, networkchaos, string(tcType[0:2])+ipSetPostFix, m.Source)
	impl.Log.Info("apply traffic control with filter", "sources", m.Source, "ipset", dstIpset)

	m.T.Append(dstIpset)
	m.T.Append(v1alpha1.RawTrafficControl{
		Type:        tcType,
		TcParameter: spec.TcParameter,
		Source:      m.Source,
		IPSet:       dstIpset.Name,
	})

	return nil
}

func NewImpl(c client.Client, log logr.Logger) *Impl {
	return &Impl{
		Client: c,
		Log:    log.WithName("trafficcontrol"),
	}
}
