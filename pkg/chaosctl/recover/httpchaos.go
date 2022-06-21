// Copyright 2022 Chaos Mesh Authors.
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

	"github.com/pkg/errors"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

type httpRecoverer struct {
	tproxyCleaner Recoverer
}

func HTTPRecoverer(client *ctrlclient.CtrlClient) Recoverer {
	return &httpRecoverer{
		tproxyCleaner: newCleanProcessRecoverer(client, "tproxy"),
	}
}

func (r *httpRecoverer) Recover(ctx context.Context, pod *PartialPod) error {
	// TODO: need hostPath to store rules
	err := r.tproxyCleaner.Recover(ctx, pod)
	if err != nil {
		return errors.Wrap(err, "clean chaos-tproxy processes")
	}
	return nil
}
