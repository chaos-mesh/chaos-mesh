package ctrlserver

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
)

func Handler() http.Handler {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))
}
