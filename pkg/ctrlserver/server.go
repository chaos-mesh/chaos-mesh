package ctrlserver

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/go-logr/logr"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
)

func Handler(logger logr.Logger, client client.Client, clientset kubernetes.Interface) http.Handler {
	return handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{Log: logger, Client: client, Clientset: clientset}}))
}
