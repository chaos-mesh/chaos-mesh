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
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/pingcap/chaos-mesh/controllers"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/go-logr/logr"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"

	v1 "k8s.io/api/core/v1"
	k8sError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	networkPartitionActionMsg = "part network for %s"

	sourceIpSetPostFix = "src"
	targetIpSetPostFix = "tgt"
)

func NewReconciler(c client.Client, log logr.Logger, req ctrl.Request) twophase.Reconciler {
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

// Apply is a functions used to apply partition chaos.
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
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	targets, err := utils.SelectAndGeneratePods(ctx, r.Client, &networkchaos.Spec.Target)

	if err != nil {
		r.Log.Error(err, "failed to select and generate pods")
		return err
	}

	sourceSet := r.generateSet(sources, networkchaos, sourceIpSetPostFix)
	targetSet := r.generateSet(targets, networkchaos, targetIpSetPostFix)

	allPods := append(sources, targets...)

	// Set up ipset in every related pods
	g := errgroup.Group{}
	for index := range allPods {
		pod := allPods[index]
		r.Log.Info("PODS", "name", pod.Name, "namespace", pod.Namespace)
		g.Go(func() error {
			err := r.flushPodIPSet(ctx, &pod, sourceSet, networkchaos)
			if err != nil {
				return err
			}

			r.Log.Info("flush ipset on pod", "name", pod.Name, "namespace", pod.Namespace)
			return r.flushPodIPSet(ctx, &pod, targetSet, networkchaos)
		})
	}

	if err = g.Wait(); err != nil {
		r.Log.Error(err, "flush pod ipset error")
		return err
	}

	if networkchaos.Spec.Direction == v1alpha1.To || networkchaos.Spec.Direction == v1alpha1.Both {
		if err := r.BlockSet(ctx, sources, targetSet, pb.Rule_OUTPUT, networkchaos); err != nil {
			r.Log.Error(err, "set source iptables failed")
			return err
		}

		if err := r.BlockSet(ctx, targets, sourceSet, pb.Rule_INPUT, networkchaos); err != nil {
			r.Log.Error(err, "set target iptables failed")
			return err
		}
	}

	if networkchaos.Spec.Direction == v1alpha1.From || networkchaos.Spec.Direction == v1alpha1.Both {
		if err := r.BlockSet(ctx, sources, targetSet, pb.Rule_INPUT, networkchaos); err != nil {
			r.Log.Error(err, "set source iptables failed")
			return err
		}

		if err := r.BlockSet(ctx, targets, sourceSet, pb.Rule_OUTPUT, networkchaos); err != nil {
			r.Log.Error(err, "set target iptables failed")
			return err
		}
	}

	networkchaos.Status.Experiment.Pods = []v1alpha1.PodStatus{}

	for _, pod := range allPods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
			Message:   fmt.Sprintf(networkPartitionActionMsg, *networkchaos.Spec.Duration),
		}

		networkchaos.Status.Experiment.Pods = append(networkchaos.Status.Experiment.Pods, ps)
	}

	return nil
}

func (r *Reconciler) BlockSet(ctx context.Context, pods []v1.Pod, set pb.IpSet, direction pb.Rule_Direction, networkchaos *v1alpha1.NetworkChaos) error {
	g := errgroup.Group{}
	sourceRule := r.generateIPTables(pb.Rule_ADD, direction, set.Name)

	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		switch direction {
		case pb.Rule_INPUT:
			networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, "input-"+key)
		case pb.Rule_OUTPUT:
			networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, "output"+key)
		}

		g.Go(func() error {
			return r.sendIPTables(ctx, pod, sourceRule, networkchaos)
		})
	}
	return g.Wait()
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
		r.Log.Error(err, "cleanFinalizersAndRecover failed")
		return err
	}

	return nil
}

func (r *Reconciler) generateSetName(networkchaos *v1alpha1.NetworkChaos, namePostFix string) string {
	r.Log.Info("generating name for chaos", "name", networkchaos.Name)
	originalName := networkchaos.Name

	var ipsetName string
	if len(originalName) < 6 {
		ipsetName = originalName + "_" + namePostFix
	} else {
		namePrefix := originalName[0:5]
		nameRest := originalName[5:]

		hasher := sha1.New()
		hasher.Write([]byte(nameRest))
		hashValue := fmt.Sprintf("%x", hasher.Sum(nil))

		// keep the length does not exceed 27
		ipsetName = namePrefix + "_" + hashValue[0:17] + "_" + namePostFix
	}

	r.Log.Info("name generated", "ipsetName", ipsetName)
	return ipsetName
}

func (r *Reconciler) generateSet(pods []v1.Pod, networkchaos *v1alpha1.NetworkChaos, namePostFix string) pb.IpSet {
	name := r.generateSetName(networkchaos, namePostFix)
	ips := make([]string, 0, len(pods))

	for _, pod := range pods {
		if len(pod.Status.PodIP) > 0 {
			ips = append(ips, pod.Status.PodIP)
		}
	}

	r.Log.Info("creating ipset", "name", name, "ips", ips)
	return pb.IpSet{
		Name: name,
		Ips:  ips,
	}
}

func (r *Reconciler) generateIPTables(action pb.Rule_Action, direction pb.Rule_Direction, set string) pb.Rule {
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
	for _, key := range networkchaos.Finalizers {
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
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
			continue
		}

		var rule pb.Rule

		if networkchaos.Spec.Direction != v1alpha1.From {
			switch direction {
			case "output":
				set := r.generateSetName(networkchaos, targetIpSetPostFix)
				rule = r.generateIPTables(pb.Rule_DELETE, pb.Rule_OUTPUT, set)
			case "input-":
				set := r.generateSetName(networkchaos, sourceIpSetPostFix)
				rule = r.generateIPTables(pb.Rule_DELETE, pb.Rule_INPUT, set)
			}

			err = r.sendIPTables(ctx, &pod, rule, networkchaos)
			if err != nil {
				r.Log.Error(err, "error while deleting iptables rules")
				return err
			}
		}

		if networkchaos.Spec.Direction != v1alpha1.To {
			switch direction {
			case "output":
				set := r.generateSetName(networkchaos, sourceIpSetPostFix)
				rule = r.generateIPTables(pb.Rule_DELETE, pb.Rule_OUTPUT, set)
			case "input-":
				set := r.generateSetName(networkchaos, targetIpSetPostFix)
				rule = r.generateIPTables(pb.Rule_DELETE, pb.Rule_INPUT, set)
			}

			err = r.sendIPTables(ctx, &pod, rule, networkchaos)
			if err != nil {
				r.Log.Error(err, "error while deleting iptables rules")
				return err
			}
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}
	r.Log.Info("after recovering", "finalizers", networkchaos.Finalizers)

	return nil
}

func (r *Reconciler) flushPodIPSet(ctx context.Context, pod *v1.Pod, ipset pb.IpSet, networkchaos *v1alpha1.NetworkChaos) error {
	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	ctx, cancel := context.WithTimeout(ctx,
		time.Duration(controllers.RPCTimeout)*time.Millisecond)
	defer cancel()

	_, err = pbClient.FlushIpSet(ctx, &pb.IpSetRequest{
		Ipset:       &ipset,
		ContainerId: containerID,
	})
	return err
}

func (r *Reconciler) sendIPTables(ctx context.Context, pod *v1.Pod, rule pb.Rule, networkchaos *v1alpha1.NetworkChaos) error {
	c, err := utils.CreateGrpcConnection(ctx, r.Client, pod)
	if err != nil {
		return err
	}
	defer c.Close()

	pbClient := pb.NewChaosDaemonClient(c)

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	ctx, cancel := context.WithTimeout(ctx,
		time.Duration(controllers.RPCTimeout)*time.Millisecond)
	defer cancel()

	_, err = pbClient.FlushIptables(ctx, &pb.IpTablesRequest{
		Rule:        &rule,
		ContainerId: containerID,
	})
	return err
}
