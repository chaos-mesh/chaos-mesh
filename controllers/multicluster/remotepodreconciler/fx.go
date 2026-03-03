// Copyright 2022 Chaos Mesh Authors.
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

package remotepodreconciler

import (
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

func Bootstrap(mgr ctrl.Manager, podReconciler *Reconciler, logger logr.Logger) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			// TODO: in a newer version `controller-runtime`, it would be
			// possible to set a customized rate limiter.
			//
			// After upgrading `controller-runtime`, we could choose a better
			// rate limiter for error handling
			MaxConcurrentReconciles: 1,
		}).
		For(&corev1.Pod{}).
		Complete(podReconciler)
}

var Module = fx.Options(
	fx.Provide(
		fx.Annotate(
			New,
			fx.ParamTags(`name:"manage-client"`, `name:"cluster-name"`),
		),
	),
	fx.Invoke(Bootstrap),
)
