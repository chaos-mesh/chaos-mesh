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

	return ctrlclient.NewCtrlClient(fmt.Sprintf("http://127.0.0.1:%d/query", port)), cancel, nil
}
