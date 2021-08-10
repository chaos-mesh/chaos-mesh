package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/generated"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrlserver/graph/model"
)

func (r *queryResolver) Namepsace(ctx context.Context, ns *string) (*model.Namespace, error) {
	panic(fmt.Errorf("not implemented"))
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
