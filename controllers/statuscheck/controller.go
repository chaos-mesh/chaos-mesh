// Copyright Chaos Mesh Authors.
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

package statuscheck

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
)

type Reconciler struct {
	logger        logr.Logger
	kubeClient    client.Client
	eventRecorder recorder.ChaosRecorder
	manager       Manager
}

func NewReconciler(logger logr.Logger, kubeClient client.Client, eventRecorder recorder.ChaosRecorder, manager Manager) *Reconciler {
	return &Reconciler{
		kubeClient:    kubeClient,
		eventRecorder: eventRecorder,
		logger:        logger,
		manager:       manager,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	startTime := time.Now()
	obj := &v1alpha1.StatusCheck{}
	if err := r.kubeClient.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			// the StatusCheck is deleted, remove it from manger
			r.manager.Delete(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// // if status check was completed previously, we don't want to redo the termination
	if obj.IsCompleted() {
		return ctrl.Result{}, nil
	}

	result, ok := r.manager.Get(*obj)
	if !ok {
		// if nil, add status check to manager
		r.manager.Add(*obj)
		result, _ = r.manager.Get(*obj)
	}

	updateError := retry.RetryOnConflict(retry.DefaultBackoff, r.updateStatus(ctx, req, result, startTime))

	return ctrl.Result{}, client.IgnoreNotFound(updateError)
}

func (r *Reconciler) updateStatus(ctx context.Context, req ctrl.Request, result Result, startTime time.Time) func() error {
	return func() error {
		statusCheck := &v1alpha1.StatusCheck{}
		if err := r.kubeClient.Get(ctx, req.NamespacedName, statusCheck); err != nil {
			r.logger.Error(err, "unable to get status check")
			return err
		}

		if statusCheck.Status.StartTime == nil {
			statusCheck.Status.StartTime = &metav1.Time{Time: startTime}
		}
		statusCheck.Status.Count = result.Count
		statusCheck.Status.Records = result.Records

		conditions, err := generateConditions(*statusCheck)
		if err != nil {
			return err
		}
		statusCheck.Status.Conditions = conditions

		if statusCheck.IsCompleted() {
			if statusCheck.Status.CompletionTime == nil {
				statusCheck.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			}
			r.manager.Complete(*statusCheck)
			r.eventRecorder.Event(statusCheck, recorder.StatusCheckCompleted{})
			r.logger.Info("status check completed", "statuscheck", req.NamespacedName)
		}

		r.logger.V(1).Info("update status of status check", "statuscheck", req.NamespacedName)
		return r.kubeClient.Status().Update(ctx, statusCheck)
	}
}
