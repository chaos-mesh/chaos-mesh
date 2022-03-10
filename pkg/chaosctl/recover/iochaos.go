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

package recover

import (
	"context"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type ioRecover struct {
	client *ctrlclient.CtrlClient
}

func IORecover(client *ctrlclient.CtrlClient) Recover {
	return &ioRecover{
		client: client,
	}
}

func (r *ioRecover) Recover(ctx context.Context, pod *ctrlclient.PartialPod) error {
	// TODO: need hostPath to store replaced fds
	return nil
}
