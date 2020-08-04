// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package rawpodnetworkchaos

import (
	"context"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Handler struct {
	client.Client
	Log logr.Logger
}

func (h *Handler) Apply(ctx context.Context, chaos *v1alpha1.RawPodNetworkChaos) error {
	h.Log.Info("updating network chaos", "pod", chaos.Namespace+"/"+chaos.Name, "spec", chaos.Spec)

	return nil
}
