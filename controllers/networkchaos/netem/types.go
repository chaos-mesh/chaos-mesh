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

package netem

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pingcap/chaos-mesh/api/v1alpha1"
	"github.com/pingcap/chaos-mesh/controllers/common"
	"github.com/pingcap/chaos-mesh/controllers/networkchaos/ipset"
	"github.com/pingcap/chaos-mesh/controllers/networkchaos/tc"
	"github.com/pingcap/chaos-mesh/controllers/twophase"
	pb "github.com/pingcap/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/pingcap/chaos-mesh/pkg/utils"
)

const (
	networkNetemActionMsg = "network netem action duration %s"
	invalidNetemSpecMsg   = "invalid spec for netem action, at least one is required from delay, loss, duplicate, corrupt"

	ipsetPostFix = "netem"
)

// NetemSpec defines the interface to convert to a Netem protobuf
type NetemSpec interface {
	ToNetem() (*pb.Netem, error)
}

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
	r.Log.Info("netem Apply", "req", req, "chaos", chaos)

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

	// We should only apply filter when we specify targets
	if networkchaos.Spec.Target != nil {
		targets, err = utils.SelectAndFilterPods(ctx, r.Client, networkchaos.Spec.Target)
		if err != nil {
			r.Log.Error(err, "failed to select and filter pods")
			return err
		}
	}

	pods := append(sources, targets...)

	switch networkchaos.Spec.Direction {
	case v1alpha1.To:
		err = r.applyNetem(ctx, sources, targets, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply netem", "sources", sources, "targets", targets)
			return err
		}
	case v1alpha1.From:
		err = r.applyNetem(ctx, targets, sources, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply netem", "sources", targets, "targets", sources)
			return err
		}
	case v1alpha1.Both:
		err = r.applyNetem(ctx, pods, pods, networkchaos)
		if err != nil {
			r.Log.Error(err, "failed to apply netem", "sources", pods, "targets", pods)
			return err
		}
	}

	networkchaos.Status.Experiment.PodRecords = make([]v1alpha1.PodStatus, 0, len(pods))
	for _, pod := range pods {
		ps := v1alpha1.PodStatus{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			HostIP:    pod.Status.HostIP,
			PodIP:     pod.Status.PodIP,
			Action:    string(networkchaos.Spec.Action),
		}

		if networkchaos.Spec.Duration != nil {
			ps.Message = fmt.Sprintf(networkNetemActionMsg, *networkchaos.Spec.Duration)
		}

		networkchaos.Status.Experiment.PodRecords = append(networkchaos.Status.Experiment.PodRecords, ps)
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosInjected, "")
	return nil
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
		return err
	}
	r.Event(networkchaos, v1.EventTypeNormal, utils.EventChaosRecovered, "")
	return nil
}

func (r *Reconciler) cleanFinalizersAndRecover(ctx context.Context, networkchaos *v1alpha1.NetworkChaos) error {
	var result error

	for _, key := range networkchaos.Finalizers {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
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
			if !k8serror.IsNotFound(err) {
				result = multierror.Append(result, err)
				continue
			}

			r.Log.Info("Pod not found", "namespace", ns, "name", name)
			networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
			continue
		}

		err = r.recoverPod(ctx, &pod, networkchaos)
		if err != nil {
			result = multierror.Append(result, err)
			continue
		}

		networkchaos.Finalizers = utils.RemoveFromFinalizer(networkchaos.Finalizers, key)
	}

	if networkchaos.Annotations[common.AnnotationCleanFinalizer] == common.AnnotationCleanFinalizerForced {
		r.Log.Info("Force cleanup all finalizers", "chaos", networkchaos)
		networkchaos.Finalizers = make([]string, 0)
		return nil
	}

	return result
}

func (r *Reconciler) recoverPod(ctx context.Context, pod *v1.Pod, _ *v1alpha1.NetworkChaos) error {
	r.Log.Info("Try to recover pod", "namespace", pod.Namespace, "name", pod.Name)

	pbClient, err := utils.NewChaosDaemonClient(ctx, r.Client, pod, common.Cfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	_, err = pbClient.DeleteNetem(ctx, &pb.NetemRequest{
		ContainerId: containerID,
		Netem:       nil,
	})

	if err != nil {
		r.Log.Error(err, "recover pod error", "namespace", pod.Namespace, "name", pod.Name)
	} else {
		r.Log.Info("Recover pod finished", "namespace", pod.Namespace, "name", pod.Name)
	}

	return err
}

func (r *Reconciler) applyPod(ctx context.Context, pod *v1.Pod, networkchaos *v1alpha1.NetworkChaos, parent, handle *pb.TcHandle) error {
	r.Log.Info("Try to apply netem on pod", "namespace", pod.Namespace, "name", pod.Name)

	var (
		netem *pb.Netem
		err   error
	)
	switch networkchaos.Spec.Action {
	case v1alpha1.NetemAction:
		netem, err = mergeNetem(networkchaos.Spec)
	default:
		action := strings.Title(string(networkchaos.Spec.Action))
		spec, ok := reflect.Indirect(reflect.ValueOf(networkchaos.Spec)).FieldByName(action).Interface().(NetemSpec)
		if !ok {
			return fmt.Errorf("spec %s is not a NetemSpec", action)
		}
		netem, err = spec.ToNetem()
	}
	if err != nil {
		return err
	}

	pbClient, err := utils.NewChaosDaemonClient(ctx, r.Client, pod, common.Cfg.ChaosDaemonPort)
	if err != nil {
		return err
	}
	defer pbClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	containerID := pod.Status.ContainerStatuses[0].ContainerID

	netem.Parent = parent
	netem.Handle = handle

	_, err = pbClient.SetNetem(ctx, &pb.NetemRequest{
		ContainerId: containerID,
		Netem:       netem,
	})

	return err
}

// mergeNetem calls ToNetem on all non nil network emulation specs and merges them into one request.
func mergeNetem(spec v1alpha1.NetworkChaosSpec) (*pb.Netem, error) {
	// NOTE: a cleaner way like
	// emSpecs = []NetemSpec{spec.Delay, spec.Loss} won't work.
	// Because in the for _, spec := range emSpecs loop,
	// spec != nil would always be true.
	// See https://stackoverflow.com/questions/13476349/check-for-nil-and-nil-interface-in-go
	// And https://groups.google.com/forum/#!topic/golang-nuts/wnH302gBa4I/discussion
	// > In short: If you never store (*T)(nil) in an interface, then you can reliably use comparison against nil
	var emSpecs []NetemSpec
	if spec.Delay != nil {
		emSpecs = append(emSpecs, spec.Delay)
	}
	if spec.Loss != nil {
		emSpecs = append(emSpecs, spec.Loss)
	}
	if spec.Duplicate != nil {
		emSpecs = append(emSpecs, spec.Duplicate)
	}
	if spec.Corrupt != nil {
		emSpecs = append(emSpecs, spec.Corrupt)
	}
	if len(emSpecs) == 0 {
		return nil, errors.New(invalidNetemSpecMsg)
	}

	merged := &pb.Netem{}
	for _, spec := range emSpecs {
		em, err := spec.ToNetem()
		if err != nil {
			return nil, err
		}
		merged = utils.MergeNetem(merged, em)
	}
	return merged, nil
}

func (r *Reconciler) applyNetem(ctx context.Context, sources, targets []v1.Pod, networkchaos *v1alpha1.NetworkChaos) error {

	g := errgroup.Group{}

	for index := range sources {
		pod := &sources[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}

		networkchaos.Finalizers = utils.InsertFinalizer(networkchaos.Finalizers, key)
	}

	// if we don't specify targets, then sources pods apply netem on all egress traffic
	if len(targets) == 0 {
		r.Log.Info("apply netem", "sources", sources)
		for index := range sources {
			pod := &sources[index]

			g.Go(func() error {
				parent := &pb.TcHandle{
					Major: 1,
					Minor: 0,
				}

				handle := &pb.TcHandle{
					Major: 1,
					Minor: 0,
				}
				return r.applyPod(ctx, pod, networkchaos, parent, handle)
			})
		}
		return g.Wait()
	}

	// create ipset contains all target ips
	dstIpset := ipset.BuildIpSet(targets, networkchaos, ipsetPostFix)
	r.Log.Info("apply netem with filter", "sources", sources, "ipset", dstIpset)

	for index := range sources {
		pod := &sources[index]

		g.Go(func() error {
			// NOTE: this is core logic of filtering:
			// we need handle both filtered traffic with netem actions like delay, loss etc,
			// and keep rest traffic as normal.
			// To filter traffic, use tc filter to classify traffic and route to netem qdisc.
			// Because the default qdisc pfifo_fast is classless, which we can't use here,
			// using prio with proper priomap as a replacement of pfifo_fast.
			//
			// Here's the qdisc tree:
			//           root 1:
			//         /  |   |  \
			//       1:1 1:2 1:3 1:4
			//        |   |   |   |
			//       10: 20: 30: 40:
			//
			// The code logic is equivalent with raw command line commands like following
			// $ipset create myset hash:ip
			// $ipset flush
			// $ipset add myset 180.101.49.11
			// $tc qdisc add dev eth0 root handle 1: prio bands 4 priomap 1 2 2 2 1 2 0 0 1 1 1 1 1 1 1 1
			// $tc qdisc add dev eth0 parent 1:1 handle 10: sfq
			// $tc qdisc add dev eth0 parent 1:2 handle 20: sfq
			// $tc qdisc add dev eth0 parent 1:3 handle 30: sfq
			// $tc qdisc add dev eth0 parent 1:4 handle 40: netem delay 10ms
			// $tc filter add dev eth0 parent 1: basic match 'ipset(myset dst)' classid 1:4

			err := ipset.FlushIpSet(ctx, r.Client, pod, dstIpset)
			if err != nil {
				return err
			}

			err = tc.AddQdisc(ctx, r.Client, pod, &pb.Qdisc{
				Parent: &pb.TcHandle{
					Major: 1,
					Minor: 0,
				},
				Handle: &pb.TcHandle{
					Major: 1,
					Minor: 0,
				},
				Type: "prio",
				// NOTE: priomap is the same as pfifo_fast qdisc,
				// so that it keeps the same behavior when handling non-classified traffic.
				// http://tldp.org/HOWTO/Adv-Routing-HOWTO/lartc.qdisc.classless.html
				// bands 4 = 3 + 1:
				// 3 is for default bands setting, similar with priomap,
				// 1 is for holding netem qdisc.
				Args: []string{"bands", "4", "priomap", "1", "2", "2", "2", "1", "2", "0", "0", "1", "1", "1", "1", "1", "1", "1", "1"},
			})
			if err != nil {
				return err
			}

			err = tc.AddQdisc(ctx, r.Client, pod, &pb.Qdisc{
				Parent: &pb.TcHandle{
					Major: 1,
					Minor: 1,
				},
				Handle: &pb.TcHandle{
					Major: 10,
					Minor: 0,
				},
				Type: "sfq",
			})
			if err != nil {
				return err
			}

			err = tc.AddQdisc(ctx, r.Client, pod, &pb.Qdisc{
				Parent: &pb.TcHandle{
					Major: 1,
					Minor: 2,
				},
				Handle: &pb.TcHandle{
					Major: 20,
					Minor: 0,
				},
				Type: "sfq",
			})
			if err != nil {
				return err
			}

			err = tc.AddQdisc(ctx, r.Client, pod, &pb.Qdisc{
				Parent: &pb.TcHandle{
					Major: 1,
					Minor: 3,
				},
				Handle: &pb.TcHandle{
					Major: 30,
					Minor: 0,
				},
				Type: "sfq",
			})
			if err != nil {
				return err
			}

			parent := &pb.TcHandle{
				Major: 1,
				Minor: 4,
			}
			handle := &pb.TcHandle{
				Major: 40,
				Minor: 0,
			}
			err = r.applyPod(ctx, pod, networkchaos, parent, handle)
			if err != nil {
				return err
			}

			return tc.AddEmatchFilter(ctx, r.Client, pod, &pb.EmatchFilter{
				Match: fmt.Sprintf("ipset(%s dst)", dstIpset.Name),
				Parent: &pb.TcHandle{
					Major: 1,
					Minor: 0,
				},
				Classid: &pb.TcHandle{
					Major: 1,
					Minor: 4,
				},
			})
		})
	}

	return g.Wait()
}
