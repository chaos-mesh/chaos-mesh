// Copyright 2021 Chaos Mesh Authors.
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

package collector

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/jinzhu/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/dashboard/core"
)

// ChaosCollector is used to collect chaos experiments into DB.
type ChaosCollector struct {
	client.Client
	Log        logr.Logger
	apiType    runtime.Object
	experiment core.ExperimentStore
	event      core.EventStore
}

// Setup setups ChaosCollector by Manager.
func (r *ChaosCollector) Setup(mgr ctrl.Manager, apiType client.Object) error {
	r.apiType = apiType

	return ctrl.NewControllerManagedBy(mgr).
		For(apiType).
		Complete(r)
}

// Reconcile reconciles ChaosCollector.
func (r *ChaosCollector) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		chaosMeta metav1.Object
		ok        bool
		isManaged = false
	)

	if r.apiType == nil {
		r.Log.Error(nil, "apiType has not been initialized")
		return ctrl.Result{}, nil
	}

	obj, ok := r.apiType.DeepCopyObject().(v1alpha1.InnerObject)
	if !ok {
		r.Log.Error(nil, "it's not a stateful object")
		return ctrl.Result{}, nil
	}

	err := r.Get(ctx, req.NamespacedName, obj)
	if err != nil {
		r.Log.Error(err, "failed to get chaos object", "request", req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if chaosMeta, ok = obj.(metav1.Object); !ok {
		r.Log.Error(nil, "failed to get chaos meta information")
	}

	if chaosMeta.GetLabels()[v1alpha1.LabelManagedBy] != "" {
		isManaged = true
	}

	// Ignore errors because logging is already done in the function
	r.createOrUpdateExperiment(req, obj)

	if obj.IsDeleted() && isManaged {
		if err = r.event.DeleteByUID(ctx, string(chaosMeta.GetUID())); err != nil {
			r.Log.Error(err, "failed to delete experiment related events")
		}

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *ChaosCollector) createOrUpdateExperiment(req ctrl.Request, obj v1alpha1.InnerObject) error {
	var (
		chaosMeta metav1.Object
		ok        bool
	)

	if chaosMeta, ok = obj.(metav1.Object); !ok {
		r.Log.Error(nil, "failed to get chaos meta information")
	}

	uid, fieldAction, action :=
		string(chaosMeta.GetUID()),
		reflect.ValueOf(obj).Elem().FieldByName("Spec").FieldByName("Action"),
		""
	if fieldAction.IsValid() {
		action = fieldAction.String()
	}
	exp := &core.Experiment{
		ExperimentMeta: core.ExperimentMeta{
			UID:       uid,
			Namespace: req.Namespace,
			Name:      req.Name,
			Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
			Action:    action,
		},
	}

	exp.CreatedAt = chaosMeta.GetCreationTimestamp().Time
	if chaosMeta.GetDeletionTimestamp() != nil {
		exp.DeletedAt = &chaosMeta.GetDeletionTimestamp().Time
	}

	data, err := json.Marshal(chaosMeta)
	if err != nil {
		r.Log.Error(err, "failed to marshal chaos", "namespace", exp.Namespace, "name", exp.Name)
		return err
	}

	exp.Experiment = string(data)

	find, err := r.experiment.FindByUID(context.Background(), uid)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		r.Log.Error(err, "failed to find experiment", "UID", uid)
		return err
	}

	if find != nil {
		exp.ID = find.ID
	}

	if err := r.experiment.Save(context.Background(), exp); err != nil {
		r.Log.Error(err, "failed to create or update an experiment in db", "experiment", exp)
		return err
	}

	return nil
}
