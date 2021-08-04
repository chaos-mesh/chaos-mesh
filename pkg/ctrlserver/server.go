package ctrlserver

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
)

func Handler(client client.Client) http.Handler {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Client: client}}))
}
