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
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	clusterName string

	logger       logr.Logger
	manageClient client.Client
	localClient  client.Client
}

func New(manageClient client.Client, clusterName string, localClient client.Client, logger logr.Logger) *Reconciler {
	return &Reconciler{
		clusterName:  clusterName,
		logger:       logger.WithName("example-pod"),
		manageClient: manageClient,
		localClient:  localClient,
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger.Info("reconcile pod", "clusterName", r.clusterName, "namespace", req.Namespace, "name", req.Name)
	return ctrl.Result{}, nil
}
