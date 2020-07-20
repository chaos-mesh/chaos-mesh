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

package partition

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/controllers/networkchaos/netutils"
	"github.com/chaos-mesh/chaos-mesh/controllers/twophase"
	pb "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
)

const (
	networkPartitionActionMsg = "partition network duration %s"

	sourceIpSetPostFix = "src"
	targetIpSetPostFix = "tgt"
)

func newReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client:        c,
			EventRecorder: recorder,
			Log:           log,
		},
		Client: c,
		Log:    log,
	}
}

// NewTwoPhaseReconciler would create Reconciler for twophase package
func NewTwoPhaseReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *twophase.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return twophase.NewReconciler(r, r.Client, r.Log)
}

// NewCommonReconciler would create Reconciler for common package
func NewCommonReconciler(c client.Client, log logr.Logger, req ctrl.Request,
	recorder record.EventRecorder) *common.Reconciler {
	r := newReconciler(c, log, req, recorder)
	return common.NewReconciler(r, r.Client, r.Log)
}

type Reconciler struct {
	client.Client
	record.EventRecorder
	Log logr.Logger
}

// Object implements the reconciler.InnerReconciler.Object
func (r *Reconciler) Object() v1alpha1.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

// Apply implements the reconciler.InnerReconciler.Apply
func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	r.Log.Info("Applying network partition")

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	sources, err := utils.SelectAndFilterPods(ctx, r.Client, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "failed to select and filter pods")
		return err
	}

	var targets []v1.Pod

	if networkchaos.Spec.Target != nil {
		targets, err = utils.SelectAndFilterPods(ctx, r.Client, networkchaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter pods")
			return err
		}
	}

	sourceSet := ipset.BuildIpSet(sources, []string{}, networkchaos, sourceIpSetPostFix)
	externalCidrs, err := netutils.ResolveCidrs(networkchaos.Spec.ExternalTargets)
	if err != nil {
		r.Log.Error(err, "failed to resolve external targets")
		return err
	}
	targetSet := ipset.BuildIpSet(targets, externalCidrs, networkchaos, targetIpSetPostFix)

	allPods := append(sources, targets...)

	// Set up ipset in every related pods
	g := errgroup.Group{}
	for index := range allPods {
		pod := allPods[index]
		r.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)
		g.Go(func() error {
			err = ipset.FlushIpSets(ctx, r.Client, &pod, []*pb.IpSet{&sourceSet})
			if err != nil {
				return err
			}

			r.Log.Info("Flush ipset on pod", "name", pod.Name, "namespace", pod.Namespace)
			return ipset.FlushIpSets(ctx, r.Client, &pod, []*pb.IpSet{&targetSet})
		})
	}

	if err = g.Wait(); err != nil {
		r.Log.Error(err, "flush pod ipset error")
		return err
	}

	if networkchaos.Spec.Direction == v1alpha1.To || networkchaos.Spec.Direction == v1alpha1.Both {
		if err := r.BlockSet(ctx, sources, &targetSet, pb.Chain_OUTPUT, networkchaos); err != nil {
			r.Log.Error(err, "set source iptables failed")
			return err
		}

		if err := r.BlockSet(ctx, targets, &sourceSet, pb.Chain_INPUT, networkchaos); err != nil {
			r.Log.Error(err, "set target iptables failed")
			return err
		}
	}

	if networkchaos.Spec.Direction == v1alpha1.From || networkchaos.Spec.Direction == v1alpha1.Both {
		if err := r.BlockSet(ctx, sources, &targetSet, pb.Chain_INPUT, networkchaos); err != nil {
			r.Log.Error(err, "set source iptables failed")
			return err
		}

		if err := r.BlockSet(ctx, targets, &sourceSet, pb.Chain_OUTPUT, networkchaos); err != nil {
			r.Log.Error(err, "set target iptables failed")
			return err
		}
	}

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(allPods))
	for _, pod := range allPods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkPartitionActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}

	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
}

// BlockSet blocks ipset for pods
func (r *Reconciler) BlockSet(ctx context.Context, pods []v1.Pod, set *pb.IpSet, direction pb.Chain_Direction, networkchaos *v1alpha1.NetworkChaos) error {
	g := errgroup.Group{}

	var chainName string
	switch direction {
	case pb.Chain_INPUT:
		chainName = "INPUT/" + netutils.CompressName(networkchaos.Name, 21, "")
	case pb.Chain_OUTPUT:
		chainName = "OUTPUT/" + netutils.CompressName(networkchaos.Name, 20, "")
	}

	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		switch direction {
		case pb.Chain_INPUT:
			networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, "input-"+key)
		case pb.Chain_OUTPUT:
			networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, "output"+key)
		}

		g.Go(func() error {
			return iptable.FlushIptablesChains(ctx, r.Client, pod, chainName, direction, []string{set.Name})
		})
	}
	return g.Wait()
}

// Recover implements the reconciler.InnerReconciler.Recover
func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	if err := r.cleanFinalizersAndRecover(ctx, networkchaos); err != nil {
		r.Log.Error(err, "cleanFinalizersAndRecover failed")
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")

	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	var result error

	for _, key := range networkchaos.Finalizers {
		direction := key[0:6]

		podKey := key[6:]
		ns, name, err := cache.SplitMetaNamespaceKey(podKey)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8sError.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
			continue
		}

		var chainName string
		var chainDirection pb.Chain_Direction
		switch direction {
		case "input-":
			chainName = "INPUT/" + netutils.CompressName(networkchaos.Name, 21, "")
			chainDirection = pb.Chain_INPUT
		case "output":
			chainName = "OUTPUT/" + netutils.CompressName(networkchaos.Name, 20, "")
			chainDirection = pb.Chain_OUTPUT
		}

		err = iptable.FlushIptablesChains(ctx, r.Client, &pod, chainName, chainDirection, nil)
		if err != nil {
			r.Log.Error(err, "error while deleting iptables rules")
			result = multierror.Append(result, err)
			continue
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}
	r.Log.Info("After recovering", "finalizers", networkchaos.Finalizers)

	if networkchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", networkchaos)
		networkchaos.Finalizers = networkchaos.Finalizers[:0]
		return nil
	}

	return result
}
