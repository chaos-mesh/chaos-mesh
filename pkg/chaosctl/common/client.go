package common

import (
	"context"
	"fmt"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

func CreateClient(ctx context.Context, managerNamespace, managerSvc string) (*ctrlclient.CtrlClient, context.CancelFunc, error) {
	cancel, port, err := ctrlclient.ForwardCtrlServer(ctx, managerNamespace, managerSvc)
	if err != nil {
		return nil, nil, err
	}

	client, err := ctrlclient.NewCtrlClient(ctx, fmt.Sprintf("http://127.0.0.1:%d/query", port))
	if err != nil {
		return nil, nil, fmt.Errorf("fail to init ctrl client: %s", err)
	}

	return client, cancel, nil
}
