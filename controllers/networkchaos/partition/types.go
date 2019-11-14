// Copyright 2019 PingCAP, Inc.
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
	"github.com/go-logr/logr"
	"github.com/pingcap/chaos-operator/api/v1alpha1"
	"github.com/pingcap/chaos-operator/controllers/twophase"
	pb "github.com/pingcap/chaos-operator/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-operator/pkg/utils"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	networkPartitionActionMsg = "part network for %s"
)

func NewConciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
	return twophase.Reconciler{
		InnerReconciler: &Reconciler{
			Client: c,
			Log:    log,
		},
		Client: c,
		Log:    log,
	}
}

type Reconciler struct {
	client.Client
	Log logr.Logger
}

func (r *Reconciler) Object() twophase.InnerObject {
	return &v1alpha1.NetworkChaos{}
}

func (r *Reconciler) Apply(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	r.Log.Info("applying network partition")

	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	sources, err := utils.SelectAndGeneratePods(ctx, r.Client, &networkchaos.Spec)

	if err != nil {
		r.Log.Error(err, "fail to select and generate pods")
		return err
	}

	targets, err := utils.SelectAndGeneratePods(ctx, r.Client, &networkchaos.Spec.Target)

	if err != nil {
		r.Log.Error(err, "fail to select and generate pods")
		return err
	}

	sourceSet := r.generateSet(sources, networkchaos, "source")
	targetSet := r.generateSet(targets, networkchaos, "target")

	allPods := append(sources, targets...)

	{
		g := errgroup.Group{}
		for _, pod := range allPods {
			pod := pod
			r.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)
			g.Go(func() error {
				err := r.flushPodIpSet(ctx, &pod, sourceSet, networkchaos)
				if err != nil {
					return err
				}

				r.Log.Info("flush ipset on pod", "name", pod.Name, "namespace", pod.Namespace)
				return r.flushPodIpSet(ctx, &pod, targetSet, networkchaos)
			})
		}

		if err = g.Wait(); err != nil {
			r.Log.Error(err, "flush pod ipset error")
			return err
		}
	}

	{
		g := errgroup.Group{}
		sourceRule := r.generateIpTables(pb.Rule_ADD, pb.Rule_OUTPUT, targetSet.Name)
		for _, pod := range sources {
			pod := pod
			g.Go(func() error {
				key, err := cache.MetaNamespaceKeyFunc(&pod)
				if err != nil {
					return err
				}

				networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, "source"+key)
				return r.sendIpTables(ctx, &pod, sourceRule, networkchaos)
			})
		}

		if err = g.Wait(); err != nil {
			r.Log.Error(err, "set source iptables failed")
			return err
		}
	}

	{
		g := errgroup.Group{}
		targetRule := r.generateIpTables(pb.Rule_ADD, pb.Rule_INPUT, sourceSet.Name)
		for _, pod := range targets {
			pod := pod
			g.Go(func() error {
				key, err := cache.MetaNamespaceKeyFunc(&pod)
				if err != nil {
					return err
				}

				networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, "target"+key)
				return r.sendIpTables(ctx, &pod, targetRule, networkchaos)
			})
		}

		if err = g.Wait(); err != nil {
			r.Log.Error(err, "set target iptables failed")
			return err
		}
	}

	return nil
}

func (r *Reconciler) Recover(ctx context.Context, req ctrl.Request, chaos twophase.InnerObject) error {
	networkchaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		err := errors.New("chaos is not NetworkChaos")
		r.Log.Error(err, "chaos is not NetworkChaos", "chaos", chaos)

		return err
	}

	err := r.cleanFinalizersAndRecover(ctx, networkchaos)
	if err != nil {
		return err
	}

	networkchaos.Status.Experiment.EndTime = &metav1.Time{
		Time: time.Now(),
	}
	networkchaos.Status.Experiment.Phase = v1alpha1.ExperimentPhaseFinished

	return nil
}

func (r *Reconciler) generateSetName(networkchaos *v1alpha1.NetworkChaos, namePostFix string) string {
	return networkchaos.Name + "_" + namePostFix
}

func (r *Reconciler) generateSet(pods []v1.Pod, networkchaos *v1alpha1.NetworkChaos, namePostFix string) pb.IpSet {
	name := r.generateSetName(networkchaos, namePostFix)
	ips := make([]string, len(pods))

	for index, pod := range pods {
		ips[index] = pod.Status.PodIP
	}

	r.Log.Info("creating ipset", "name", name, "ips", ips)
	return pb.IpSet{
		Name: name,
		Ips:  ips,
	}
}

func (r *Reconciler) generateIpTables(action pb.Rule_Action, direction pb.Rule_Direction, set string) pb.Rule {
	return pb.Rule{
		Action:    action,
		Direction: direction,
		Set:       set,
	}
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	if len(networkchaos.Finalizers) == 0 {
		return nil
	}
	for index, key := range networkchaos.Finalizers {
		direction := key[0:6]

		podKey := key[6:]
		ns, name, err := cache.SplitMetaNamespaceKey(podKey)
		if err != nil {
			return err
		}

		var pod v1.Pod
		err = r.Get(ctx, types.NamespacedName{
			Namespace: ns,
			Name:      name,
		}, &pod)

		if err != nil {
			if !k8sError.IsNotFound(err) {
				return err
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, index)
			continue
		}

		var rule pb.Rule

		switch direction {
		case "source":
			set := r.generateSetName(networkchaos, "target")
			rule = r.generateIpTables(pb.Rule_DELETE, pb.Rule_OUTPUT, set)
		case "target":
			set := r.generateSetName(networkchaos, "source")
			rule = r.generateIpTables(pb.Rule_DELETE, pb.Rule_INPUT, set)
		}

		err = r.sendIpTables(ctx, &pod, rule, networkchaos)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, index)
	}

	return nil
}

func (r *Reconciler) flushPodIpSet(ctx context.Context, pod *v1.Pod, ipset pb.IpSet, networkchaos *v1alpha1.NetworkChaos) error {
	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	containerId := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.FlushIpSet(ctx, &pb.IpSetRequest{
		Ipset:       &ipset,
		ContainerId: containerId,
	})
	return err
}

func (r *Reconciler) sendIpTables(ctx context.Context, pod *v1.Pod, rule pb.Rule, networkchaos *v1alpha1.NetworkChaos) error {
	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	containerId := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.FlushIptables(ctx, &pb.IpTablesRequest{
		Rule:        &rule,
		ContainerId: containerId,
	})
	return err
}
