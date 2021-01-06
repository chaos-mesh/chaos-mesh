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

package manager

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/node"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/kubernetesstuff"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/manager/sideeffect/resolver"
	"github.com/chaos-mesh/chaos-mesh/pkg/workflow/trigger"
)

func BootstrapManager(kubeclient client.Client, logger logr.Logger, controllerTrigger trigger.Trigger) (WorkflowManager, error) {
	pg := kubernetesstuff.NewKubernetesPlayground(kubeclient)
	repo := kubernetesstuff.NewKubernetesWorkflowRepo(kubeclient)
	eventTrigger := trigger.NewOperableTrigger()
	sideEffectsResolver, err := resolver.NewCompositeResolverWith(
		resolver.NewActorResolver(pg),
		resolver.NewCreateNewNodeResolver(repo),
		resolver.NewUpdateNodePhaseResolver(repo),
		resolver.NewNotifyNewEventResolver(eventTrigger),
	)
	if err != nil {
		return nil, err
	}
	manager := NewBasicManager("bootstrapped-manager", repo, logger, node.NewBasicNodeNameGenerator(), sideEffectsResolver, eventTrigger, controllerTrigger)
	return manager, nil
}
