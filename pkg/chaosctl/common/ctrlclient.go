// Copyright 2021 Chaos Mesh Authors.
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

package common

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type CtrlClient struct {
	ctx                context.Context
	client             *graphql.Client
	subscriptionClient *graphql.SubscriptionClient

	Schema *Schema
}

func NewCtrlClient(ctx context.Context, url string) (*CtrlClient, error) {
	client := &CtrlClient{
		ctx:                ctx,
		client:             graphql.NewClient(url, nil),
		subscriptionClient: graphql.NewSubscriptionClient(url),
	}

	schemaQuery := new(struct {
		Schema Schema `graphql:"__schema"`
	})

	err := client.client.Query(client.ctx, schemaQuery, nil)
	if err != nil {
		return nil, err
	}

	client.Schema = &schemaQuery.Schema
	return client, nil
}
