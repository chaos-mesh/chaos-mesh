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

package ctrl

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-logr/logr"
	"go.uber.org/fx"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/controllers/utils/chaosdaemon"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/generated"
)

type ServerParams struct {
	fx.In

	NoCacheReader       client.Reader `name:"no-cache"`
	Logger              logr.Logger
	Client              client.Client
	Clientset           *kubernetes.Clientset
	DaemonClientBuilder *chaosdaemon.ChaosDaemonClientBuilder
}

func New(param ServerParams) *handler.Server {
	resolvers := &server.Resolver{
		DaemonHelper:  &server.DaemonHelper{Builder: param.DaemonClientBuilder},
		Log:           param.Logger.WithName("ctrl-server"),
		Client:        param.Client,
		Clientset:     param.Clientset,
		NoCacheReader: param.NoCacheReader,
	}
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: resolvers}))
}
