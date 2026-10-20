// Copyright 2026 Chaos Mesh Authors.
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

package nodescoped

import (
	"context"
	"sort"

	"go.uber.org/fx"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/config"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/generic"
	genericnamespace "github.com/chaos-mesh/chaos-mesh/pkg/selector/generic/namespace"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector/pod"
)

// NodeNameField is the pod field used to list the pods scheduled on a node
const NodeNameField = "spec.nodeName"

// Target is a pod or a node selected by a NodeScopedPodSelector
type Target interface {
	Id() string
}

// Node is the host network namespace of a node
type Node struct {
	Name string
}

func (n *Node) Id() string {
	return v1alpha1.NodeRecordIDPrefix + n.Name
}

type SelectImpl struct {
	c    client.Client
	r    client.Reader
	pods *pod.SelectImpl

	// ExcludedNamespaces are never selected, so chaos-mesh never degrades itself
	ExcludedNamespaces []string

	generic.Option
}

// Select returns every eligible pod on the nodes running the pods matched by
// the pod selector, followed by those nodes when IncludeHost is set.
func (impl *SelectImpl) Select(ctx context.Context, s *v1alpha1.NodeScopedPodSelector) ([]Target, error) {
	if s == nil {
		return []Target{}, nil
	}

	victims, err := impl.pods.Select(ctx, &s.PodSelector)
	if err != nil {
		return nil, err
	}

	nodeNames := map[string]struct{}{}
	for _, victim := range victims {
		if victim.Spec.NodeName != "" {
			nodeNames[victim.Spec.NodeName] = struct{}{}
		}
	}
	nodes := make([]string, 0, len(nodeNames))
	for name := range nodeNames {
		nodes = append(nodes, name)
	}
	sort.Strings(nodes)

	var targets []Target
	for _, node := range nodes {
		var pods v1.PodList
		if err := impl.r.List(ctx, &pods, client.MatchingFields{NodeNameField: node}); err != nil {
			return nil, err
		}

		sort.Slice(pods.Items, func(i, j int) bool {
			return pods.Items[i].Namespace+"/"+pods.Items[i].Name < pods.Items[j].Namespace+"/"+pods.Items[j].Name
		})
		for i := range pods.Items {
			p := pods.Items[i]
			ok, err := impl.eligible(ctx, &p)
			if err != nil {
				return nil, err
			}
			if ok {
				targets = append(targets, &pod.Pod{Pod: p})
			}
		}
	}

	if s.IncludeHost {
		for _, node := range nodes {
			targets = append(targets, &Node{Name: node})
		}
	}

	return targets, nil
}

func (impl *SelectImpl) eligible(ctx context.Context, p *v1.Pod) (bool, error) {
	if p.Status.Phase != v1.PodRunning || p.DeletionTimestamp != nil || p.Spec.HostNetwork {
		return false, nil
	}
	for _, ns := range impl.ExcludedNamespaces {
		if p.Namespace == ns {
			return false, nil
		}
	}
	if !impl.ClusterScoped && p.Namespace != impl.TargetNamespace {
		return false, nil
	}
	if impl.EnableFilterNamespace {
		return genericnamespace.IsAllowedNamespaces(ctx, impl.c, p.Namespace)
	}
	return true, nil
}

type Params struct {
	fx.In

	Client      client.Client
	Reader      client.Reader `name:"no-cache"`
	PodSelector *pod.SelectImpl
}

func New(params Params) *SelectImpl {
	return &SelectImpl{
		c:                  params.Client,
		r:                  params.Reader,
		pods:               params.PodSelector,
		ExcludedNamespaces: []string{config.ControllerCfg.Namespace},
		Option: generic.Option{
			ClusterScoped:         config.ControllerCfg.ClusterScoped,
			TargetNamespace:       config.ControllerCfg.TargetNamespace,
			EnableFilterNamespace: config.ControllerCfg.EnableFilterNamespace,
		},
	}
}

// NewForTest builds a selector without the global controller configuration
func NewForTest(c client.Client, r client.Reader, pods *pod.SelectImpl, option generic.Option, excludedNamespaces ...string) *SelectImpl {
	return &SelectImpl{
		c:                  c,
		r:                  r,
		pods:               pods,
		ExcludedNamespaces: excludedNamespaces,
		Option:             option,
	}
}
