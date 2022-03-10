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
