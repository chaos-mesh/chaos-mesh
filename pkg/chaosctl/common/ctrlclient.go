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
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/hasura/go-graphql-client"
	"github.com/iancoleman/strcase"
)

type CtrlClient struct {
	ctx context.Context

	Client             *graphql.Client
	SubscriptionClient *graphql.SubscriptionClient
	Schema             *Schema
}

func NewCtrlClient(ctx context.Context, url string) (*CtrlClient, error) {
	client := &CtrlClient{
		ctx:                ctx,
		Client:             graphql.NewClient(url, nil),
		SubscriptionClient: graphql.NewSubscriptionClient(url),
	}

	schemaQuery := new(struct {
		Schema RawSchema `graphql:"__schema"`
	})

	err := client.Client.Query(client.ctx, schemaQuery, nil)
	if err != nil {
		return nil, err
	}

	client.Schema = NewSchema(&schemaQuery.Schema)
	return client, nil
}

func (c *CtrlClient) GetQueryType() (*Type, error) {
	return c.Schema.MustGetType(string(c.Schema.QueryType.Name))
}

// list tail arguments, expected queryStr: ["prefix1", "prefix2", "resource", "<some value> can be empty"]
func (c *CtrlClient) ListArguments(queryStr []string, argumentName string) ([]string, error) {
	queryType, err := c.GetQueryType()
	if err != nil {
		return nil, err
	}

	listQuery := queryStr[:len(queryStr)-1]
	helper := pluralize.NewClient()
	listQuery[len(listQuery)-1] = helper.Plural(listQuery[len(listQuery)-1])
	listQuery = append(listQuery, argumentName)
	query, err := c.Schema.ParseQuery(listQuery, queryType)
	if err != nil {
		return nil, err
	}

	superQuery := NewQuery("query", queryType, nil)
	superQuery.Fields["namespace"] = query
	variables := NewVariables()

	queryStruct, err := c.Schema.Reflect(superQuery, variables)
	if err != nil {
		return nil, err
	}

	queryValue := reflect.New(queryStruct.Elem()).Interface()
	err = c.Client.Query(c.ctx, queryValue, variables.GenMap())
	if err != nil {
		return nil, err
	}

	arguments, err := listArguments(queryValue, query, queryStr[len(queryStr)-1])
	if err != nil {
		return nil, err
	}

	return arguments, err
}

func listArguments(object interface{}, resource *Query, startWith string) ([]string, error) {
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Ptr:
		return listArguments(value.Elem().Interface(), resource, startWith)
	case reflect.Struct:
		field := value.FieldByName(strcase.ToCamel(resource.Name))
		if field == *new(reflect.Value) {
			return nil, fmt.Errorf("cannot find field %s in object: %#v", resource.Name, object)
		}
		for _, f := range resource.Fields {
			return listArguments(field.Interface(), f, startWith)
		}
		return listArguments(field.Interface(), nil, startWith)
	case reflect.Slice:
		slice := make([]string, 0)
		for i := 0; i < value.Len(); i++ {
			arguments, err := listArguments(value.Index(i).Interface(), resource, startWith)
			if err != nil {
				return nil, err
			}
			slice = append(slice, arguments...)
		}
		return slice, nil
	default:
		if resource != nil {
			return nil, fmt.Errorf("resource of %s kind is not supported", value.Kind())
		}
	}

	var val string

	switch v := object.(type) {
	case graphql.String:
		val = string(v)
	case graphql.Boolean:
		val = strconv.FormatBool(bool(v))
	case graphql.Int:
		val = fmt.Sprintf("%d", v)
	case graphql.Float:
		val = fmt.Sprintf("%f", v)
	default:
		return nil, fmt.Errorf("unsupported value: %#v", v)
	}

	if strings.HasPrefix(val, startWith) {
		return []string{val}, nil
	}
	return nil, nil
}
