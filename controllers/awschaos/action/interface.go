package action

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type Recoverer interface {
	Recover(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error
}

type Snapshotter interface {
	Snapshot(ctx context.Context) (*v1alpha1.AWSStatusSnapshot, error)
}

type Stopper interface {
	Snapshotter
	Stop(ctx context.Context, snapshot *v1alpha1.AWSStatusSnapshot) error
}
