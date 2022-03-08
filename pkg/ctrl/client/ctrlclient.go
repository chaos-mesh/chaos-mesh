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

package client

import (
	"context"

	"github.com/hasura/go-graphql-client"
	"k8s.io/apimachinery/pkg/types"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/ctrl/server/model"
)

type CtrlClient struct {
	QueryClient        *graphql.Client
	SubscriptionClient *graphql.SubscriptionClient
}

func NewCtrlClient(url string) *CtrlClient {
	return &CtrlClient{
		QueryClient:        graphql.NewClient(url, nil),
		SubscriptionClient: graphql.NewSubscriptionClient(url),
	}
}

func (c *CtrlClient) ListNamespace(ctx context.Context) ([]string, error) {
	namespaceQuery := new(struct {
		Namespace []struct {
			Ns string
		}
	})

	err := c.QueryClient.Query(ctx, namespaceQuery, nil)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, ns := range namespaceQuery.Namespace {
		namespaces = append(namespaces, ns.Ns)
	}

	return namespaces, nil
}

func (c *CtrlClient) SelectPods(ctx context.Context, selector v1alpha1.PodSelectorSpec) ([]types.NamespacedName, error) {
	selectInput := model.PodSelectorInput{
		Pods:                map[string]interface{}{},
		NodeSelectors:       map[string]interface{}{},
		Nodes:               selector.Nodes,
		PodPhaseSelectors:   selector.PodPhaseSelectors,
		Namespaces:          selector.Namespaces,
		FieldSelectors:      map[string]interface{}{},
		LabelSelectors:      map[string]interface{}{},
		AnnotationSelectors: map[string]interface{}{},
	}

	for k, v := range selector.Pods {
		selectInput.Pods[k] = v
	}

	for k, v := range selector.NodeSelectors {
		selectInput.NodeSelectors[k] = v
	}

	for k, v := range selector.FieldSelectors {
		selectInput.FieldSelectors[k] = v
	}

	for k, v := range selector.LabelSelectors {
		selectInput.LabelSelectors[k] = v
	}

	for k, v := range selector.AnnotationSelectors {
		selectInput.AnnotationSelectors[k] = v
	}

	podsQuery := new(struct {
		Pods []struct {
			Namespace, Name string
		} `graphql:"pods(selector: $selector)"`
	})

	err := c.QueryClient.Query(ctx, podsQuery, map[string]interface{}{"selector": selectInput})
	if err != nil {
		return nil, err
	}

	var pods []types.NamespacedName
	for _, pod := range podsQuery.Pods {
		pods = append(pods, types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name})
	}

	return pods, nil
}
