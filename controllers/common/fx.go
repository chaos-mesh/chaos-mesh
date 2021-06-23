// Copyright 2021 Chaos Mesh Authors.
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

package common

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"k8s.io/apimachinery/pkg/runtime"
	k8sTypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/chaos-mesh/chaos-mesh/controllers/types"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/builder"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/controller"
	"github.com/chaos-mesh/chaos-mesh/controllers/utils/recorder"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
)

type ChaosImplPair struct {
	Name   string
	Object InnerObjectWithSelector
	Impl   ChaosImpl

	ObjectList runtime.Object
	Controlls  []runtime.Object
}

type Params struct {
	fx.In

	Mgr             ctrl.Manager
	Client          client.Client
	Logger          logr.Logger
	Selector        *selector.Selector
	RecorderBuilder *recorder.RecorderBuilder
	Impls           []*ChaosImplPair `group:"impl"`
	Reader          client.Reader    `name:"no-cache"`
}

func NewController(params Params) (types.Controller, error) {
	logger := params.Logger
	pairs := params.Impls
	mgr := params.Mgr
	client := params.Client
	reader := params.Reader
	selector := params.Selector
	recorderBuilder := params.RecorderBuilder

	setupLog := logger.WithName("setup-common")
	for _, pair := range pairs {
		setupLog.Info("setting up controller", "resource-name", pair.Name)

		builder := builder.Default(mgr).
			For(pair.Object).
			Named(pair.Name + "-records")

		// Add owning resources
		if len(pair.Controlls) > 0 {
			pair := pair
			for _, obj := range pair.Controlls {
				builder = builder.Watches(&source.Kind{Type: obj}, &handler.EnqueueRequestsFromMapFunc{
					ToRequests: handler.ToRequestsFunc(func(obj handler.MapObject) []reconcile.Request {
						reqs := []reconcile.Request{}
						objName := k8sTypes.NamespacedName{
							Namespace: obj.Meta.GetNamespace(),
							Name:      obj.Meta.GetName(),
						}

						list := pair.ObjectList.DeepCopyObject()
						err := client.List(context.TODO(), list)
						if err != nil {
							setupLog.Error(err, "fail to list object")
						}

						items := reflect.ValueOf(list).Elem().FieldByName("Items")
						for i := 0; i < items.Len(); i++ {
							item := items.Index(i).Addr().Interface().(InnerObjectWithSelector)
							for _, record := range item.GetStatus().Experiment.Records {
								if controller.ParseNamespacedName(record.Id) == objName {
									id := k8sTypes.NamespacedName{
										Namespace: item.GetObjectMeta().Namespace,
										Name:      item.GetObjectMeta().Name,
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
				})
			}
		}

		err := builder.Complete(&Reconciler{
			Impl:     pair.Impl,
			Object:   pair.Object,
			Client:   client,
			Reader:   reader,
			Recorder: recorderBuilder.Build("records"),
			Selector: selector,
			Log:      logger.WithName("records"),
		})
		if err != nil {
			return "", err
		}

	}

	return "records", nil
}
