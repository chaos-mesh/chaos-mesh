// Copyright 2026 Chaos Mesh Authors.
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

package nodenetworkchaos

import (
	"context"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/ipset"
	"github.com/chaos-mesh/chaos-mesh/controllers/podnetworkchaos/iptable"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	chaosdaemonclient "github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/client"
)

// kubeletPort keeps the API server able to reach the kubelet (logs, exec)
// while the host network namespace is degraded.
const kubeletPort = 10250

// Reconciler applies nodenetworkchaos in the host network namespace of a node
type Reconciler struct {
	client.Client
	Recorder recorder.ChaosRecorder

	Log                      logr.Logger
	ChaosDaemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder
	ChaosDaemonPort          int
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &v1alpha1.NodeNetworkChaos{}
	if err := r.Client.Get(ctx, req.NamespacedName, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			r.Log.Error(err, "unable to get nodenetworkchaos")
		}
		return ctrl.Result{}, nil
	}

	if obj.Generation <= obj.Status.ObservedGeneration && obj.Status.FailedMessage == "" {
		r.Log.Info("the target node has been up to date", "node", obj.Name)
		return ctrl.Result{}, nil
	}

	r.Log.Info("updating nodenetworkchaos", "node", obj.Name, "spec", obj.Spec)

	err := r.apply(ctx, obj)
	if err != nil {
		r.Log.Error(err, "fail to apply nodenetworkchaos", "node", obj.Name)
		r.Recorder.Event(obj, recorder.Failed{
			Activity: "apply host network chaos",
			Err:      err.Error(),
		})
	}

	failedMessage := ""
	if err != nil {
		failedMessage = err.Error()
	}
	observedGeneration := obj.Generation
	updateError := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		latest := &v1alpha1.NodeNetworkChaos{}
		if err := r.Client.Get(ctx, req.NamespacedName, latest); err != nil {
			return err
		}
		latest.Status.FailedMessage = failedMessage
		latest.Status.ObservedGeneration = observedGeneration
		return r.Client.Status().Update(ctx, latest)
	})
	if updateError != nil {
		r.Log.Error(updateError, "fail to update nodenetworkchaos status")
		r.Recorder.Event(obj, recorder.Failed{
			Activity: "update status",
			Err:      updateError.Error(),
		})
	}

	if err != nil {
		return ctrl.Result{Requeue: true}, nil
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) apply(ctx context.Context, obj *v1alpha1.NodeNetworkChaos) error {
	pbClient, err := r.ChaosDaemonClientBuilder.BuildForNode(ctx, obj.Name, &types.NamespacedName{Name: obj.Name})
	if err != nil {
		return errors.Wrapf(err, "create chaos daemon client for node %s", obj.Name)
	}
	defer pbClient.Close()

	return r.applyWith(ctx, pbClient, obj)
}

func (r *Reconciler) applyWith(ctx context.Context, pbClient chaosdaemonclient.ChaosDaemonClientInterface, obj *v1alpha1.NodeNetworkChaos) error {
	chains, err := iptable.ToChains(obj.Spec.Iptables)
	if err != nil {
		return err
	}
	exemptPorts := r.exemptTCPPorts()
	for _, chain := range chains {
		chain.ExemptTcpPorts = exemptPorts
	}

	if err := ipset.FlushHostIPSets(ctx, pbClient, ipset.ToPBIPSets(obj.Spec.IPSets)); err != nil {
		return errors.Wrapf(err, "node %s", obj.Name)
	}
	if err := iptable.SetHostIptablesChains(ctx, pbClient, chains); err != nil {
		return errors.Wrapf(err, "node %s", obj.Name)
	}
	return nil
}

// exemptTCPPorts lists the ports whose traffic is never dropped in the host
// network namespace, so the controller can always reach chaos-daemon to recover.
func (r *Reconciler) exemptTCPPorts() string {
	ports := []string{strconv.Itoa(kubeletPort)}
	if r.ChaosDaemonPort != 0 && r.ChaosDaemonPort != kubeletPort {
		ports = append(ports, strconv.Itoa(r.ChaosDaemonPort))
	}
	return strings.Join(ports, ",")
}
