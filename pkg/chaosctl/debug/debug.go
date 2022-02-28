package debug

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type Debugger interface {
	Run(ctx context.Context, namespace, chaosName string) ([]*common.ChaosResult, error)
	List(ctx context.Context, namespace string) ([]string, error)
}

type Debug func(client *ctrlclient.CtrlClient) Debugger
