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

package common

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	chaosimpltypes "github.com/chaos-mesh/chaos-mesh/controllers/chaosimpl/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/common/pipeline"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

type Params struct {
	fx.In

	Mgr             ctrl.Manager
	Client          client.Client
	Logger          logr.Logger
	Selector        *selector.Selector
	RecorderBuilder *recorder.RecorderBuilder
	Impls           []*chaosimpltypes.ChaosImplPair `group:"impl"`
	Reader          client.Reader                   `name:"no-cache"`
	Steps           []pipeline.PipelineStep
}

func Bootstrap(params Params) error {
	logger := params.Logger
	pairs := params.Impls
	mgr := params.Mgr
	kubeclient := params.Client
	reader := params.Reader
	selector := params.Selector
	recorderBuilder := params.RecorderBuilder

	setupLog := logger.WithName("setup-common")
	for _, pair := range pairs {
		name := pair.Name + "-records"
		if !config.ShouldSpawnController(name) {
			return nil
		}

		setupLog.Info("setting up controller", "resource-name", pair.Name)

		builder := builder.Default(mgr).
			For(pair.Object).
			Named(pair.Name + "-pipeline")

		// for common CRDs, since we don't want to reconcile the object,
		// when we only change the object.status.experiment.records[].events
		predicaters := []predicate.Predicate{StatusRecordEventsChangePredicate{}}

		// Add owning resources
		if len(pair.Controlls) > 0 {
			pair := pair
			for _, obj := range pair.Controlls {
				builder.Watches(obj,
					handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
						reqs := []reconcile.Request{}
						objName := k8sTypes.NamespacedName{
							Namespace: obj.GetNamespace(),
							Name:      obj.GetName(),
						}

						list := pair.ObjectList.DeepCopyList()
						err := kubeclient.List(context.TODO(), list)
						if err != nil {
							setupLog.Error(err, "fail to list object")
						}

						items := reflect.ValueOf(list).Elem().FieldByName("Items")
						for i := 0; i < items.Len(); i++ {
							item := items.Index(i).Addr().Interface().(v1alpha1.InnerObjectWithSelector)
							for _, record := range item.GetStatus().Experiment.Records {
								namespacedName, err := controller.ParseNamespacedName(record.Id)
								if err != nil {
									setupLog.Error(err, "failed to parse record", "record", record.Id)
									continue
								}
								if namespacedName == objName {
									id := k8sTypes.NamespacedName{
										Namespace: item.GetNamespace(),
										Name:      item.GetName(),
									}
									setupLog.Info("mapping requests", "source", objName, "target", id)
									reqs = append(reqs, reconcile.Request{
										NamespacedName: id,
									})
								}
							}
						}
						return reqs
					}),
				)
			}
			predicaters = append(predicaters, PickChildCRDPredicate{})
		}

		pipe := pipeline.NewPipeline(&pipeline.PipelineContext{
			Logger: logger,
			Object: &types.Object{
				Name:   pair.Name,
				Object: pair.Object,
			},
			Impl:            pair.Impl,
			Mgr:             mgr,
			Client:          kubeclient,
			Reader:          reader,
			RecorderBuilder: recorderBuilder,
			Selector:        selector,
		})

		pipe.AddSteps(params.Steps...)
		builder = builder.WithEventFilter(predicate.And(predicate.Or(predicaters...), RemoteChaosPredicate{}))
		err := builder.Complete(pipe)
		if err != nil {
			return err
		}

	}

	return nil
}

// PickChildCRDPredicate allows events to trigger the Reconcile of Chaos CRD,
// for example:
// Reconcile of IOChaos could be triggered by changes on PodIOChaos.
// For now, we have PodHttpChaos/PodIOChaos/PodNetworkChaos which require to follow this pattern.
type PickChildCRDPredicate struct {
	predicate.Funcs
}

// Update implements UpdateEvent filter for child CRD.
func (PickChildCRDPredicate) Update(e event.UpdateEvent) bool {
	switch e.ObjectNew.(type) {
	case *v1alpha1.PodHttpChaos, *v1alpha1.PodIOChaos, *v1alpha1.PodNetworkChaos:
		return true
	}
	return false
}

// StatusRecordEventsChangePredicate skip the update event,
// when we Only update object.status.experiment.records[].events
type StatusRecordEventsChangePredicate struct {
	predicate.Funcs
}

// Update implements UpdateEvent filter for update to filter the events
// which we Only update object.status.experiment.records[].events
func (StatusRecordEventsChangePredicate) Update(e event.UpdateEvent) bool {
	objNew, ok := e.ObjectNew.DeepCopyObject().(v1alpha1.StatefulObject)
	if !ok {
		return false
	}
	objOld, ok := e.ObjectOld.DeepCopyObject().(v1alpha1.StatefulObject)
	if !ok {
		return false
	}
	statusNew := objNew.GetStatus()
	statusOld := objOld.GetStatus()
	if statusNew == nil || statusOld == nil {
		return true
	}
	objNew.SetGeneration(0)
	objOld.SetGeneration(0)
	objNew.SetResourceVersion("")
	objOld.SetResourceVersion("")
	for i := range statusNew.Experiment.Records {
		statusNew.Experiment.Records[i].Events = nil
	}
	for i := range statusOld.Experiment.Records {
		statusOld.Experiment.Records[i].Events = nil
	}
	return !reflect.DeepEqual(objNew, objOld)
}

type RemoteChaosPredicate struct {
	predicate.Funcs
}

func (RemoteChaosPredicate) Create(e event.CreateEvent) bool {
	obj, ok := e.Object.DeepCopyObject().(v1alpha1.RemoteObject)
	if !ok {
		return true
	}

	if obj.GetRemoteCluster() == "" {
		return true
	}

	return false
}

func (RemoteChaosPredicate) Update(e event.UpdateEvent) bool {
	obj, ok := e.ObjectNew.DeepCopyObject().(v1alpha1.RemoteObject)
	if !ok {
		return true
	}

	if obj.GetRemoteCluster() == "" {
		return true
	}

	return false
}

func (RemoteChaosPredicate) Delete(e event.DeleteEvent) bool {
	obj, ok := e.Object.DeepCopyObject().(v1alpha1.RemoteObject)
	if !ok {
		return true
	}

	if obj.GetRemoteCluster() == "" {
		return true
	}

	return false
}

func (RemoteChaosPredicate) Generic(e event.GenericEvent) bool {
	obj, ok := e.Object.DeepCopyObject().(v1alpha1.RemoteObject)
	if !ok {
		return true
	}

	if obj.GetRemoteCluster() == "" {
		return true
	}

	return false
}
