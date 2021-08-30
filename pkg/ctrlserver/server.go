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

package ctrlserver

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
)

func Handler(logger logr.Logger, client client.Client, clientset *kubernetes.Clientset, daemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder) http.Handler {
	resolvers := &graph.Resolver{
		DaemonHelper: &graph.DaemonHelper{Builder: daemonClientBuilder},
		Log:          logger,
		Client:       client,
		Clientset:    clientset,
	}
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolvers}))
}
