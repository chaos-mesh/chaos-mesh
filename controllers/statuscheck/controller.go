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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
			r.logger.Info("status check is deleted", "statuscheck", req.NamespacedName)
			// the StatusCheck is deleted, remove it from manger
			r.manager.Delete(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, errors.Wrapf(err, "get status check '%s'", req.NamespacedName.String())
	}

	// if status check was completed previously, we don't want to redo the termination
	if obj.IsCompleted() {
		r.logger.V(1).Info("status check is already completed", "statuscheck", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	result, ok := r.manager.Get(*obj)
	if !ok {
		// if nil, add status check to manager
		err := r.manager.Add(*obj)
		if err != nil {
			return ctrl.Result{}, errors.Wrapf(err, "add status check '%s' to manager", req.NamespacedName.String())
		}
		result, _ = r.manager.Get(*obj)
	}

	updateError := retry.RetryOnConflict(retry.DefaultBackoff, r.updateStatus(ctx, req, result, startTime))

	return ctrl.Result{RequeueAfter: time.Duration(obj.Spec.IntervalSeconds) * time.Second}, client.IgnoreNotFound(updateError)
}

func (r *Reconciler) updateStatus(ctx context.Context, req ctrl.Request, result Result, startTime time.Time) func() error {
	return func() error {
		statusCheck := &v1alpha1.StatusCheck{}
		if err := r.kubeClient.Get(ctx, req.NamespacedName, statusCheck); err != nil {
			return errors.Wrapf(err, "get status check '%s'", req.NamespacedName.String())
		}

		if statusCheck.Status.StartTime == nil {
			statusCheck.Status.StartTime = &metav1.Time{Time: startTime}
		}
		statusCheck.Status.Count = result.Count
		statusCheck.Status.Records = result.Records

		conditions, err := r.generateConditions(*statusCheck)
		if err != nil {
			return errors.Wrapf(err, "generate conditions for status check '%s'", req.NamespacedName.String())
		}

		if conditions.isCompleted() {
			if statusCheck.Status.CompletionTime == nil {
				statusCheck.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			}
			r.manager.Complete(*statusCheck)
			r.eventRecorder.Event(statusCheck, recorder.StatusCheckCompleted{Msg: conditions[v1alpha1.StatusCheckConditionCompleted].Reason})
			r.logger.Info("status check is completed", "statuscheck", req.NamespacedName)
		}
		if conditions.isDurationExceed() {
			r.eventRecorder.Event(statusCheck, recorder.StatusCheckDurationExceed{})
		}
		if conditions.isSuccessThresholdExceed() {
			r.eventRecorder.Event(statusCheck, recorder.StatusCheckSuccessThresholdExceed{})
		}
		if conditions.isFailureThresholdExceed() {
			r.eventRecorder.Event(statusCheck, recorder.StatusCheckFailureThresholdExceed{})
		}

		statusCheck.Status.Conditions = toConditionList(conditions)

		r.logger.V(1).Info("update status of status check", "statuscheck", req.NamespacedName)
		return r.kubeClient.Status().Update(ctx, statusCheck)
	}
}

func (r *Reconciler) generateConditions(statusCheck v1alpha1.StatusCheck) (conditionMap, error) {
	conditions := toConditionMap(statusCheck.Status.Conditions)

	if err := setDurationExceedCondition(statusCheck, conditions); err != nil {
		return nil, errors.Wrapf(err, "set duration exceed condition for status check '%s'", fmt.Sprintf("%s/%s", statusCheck.Namespace, statusCheck.Name))
	}
	setFailureThresholdExceedCondition(statusCheck, conditions)
	setSuccessThresholdExceedCondition(statusCheck, conditions)

	// this condition must be placed after the above three conditions
	setCompletedCondition(statusCheck, conditions)

	return conditions, nil
}
