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

const NamespaceKey = "namespace"
const NamespaceType = "Namespace"

type CtrlClient struct {
	ctx context.Context

	Client             *graphql.Client
	SubscriptionClient *graphql.SubscriptionClient
	Schema             *Schema
}

type AutoCompleteContext struct {
	namespace      string
	query          []string
	completeLeaves bool
	toComplete     *ToComplete
}

type ToComplete struct {
	root   []string
	leaves []string
}

func (c *ToComplete) Clone() *ToComplete {
	return &ToComplete{
		root:   append([]string{}, c.root...),
		leaves: append([]string{}, c.leaves...),
	}
}

func (c *ToComplete) ToQuery() string {
	var query []string
	query = append(query, c.root...)
	query = append(query, strings.Join(c.leaves, ","))
	return strings.Join(query, "/")
}

func (c *ToComplete) TrimNamespaced(namespace string) string {
	return trimNamespace(c.ToQuery(), namespace)
}

type Completion []string

func NewAutoCompleteContext(namespace, toComplete string, completeLeaves bool) *AutoCompleteContext {
	query := []string{NamespaceKey, namespace}
	toCompleteSeg := append([]string{}, query...)
	toCompleteSeg = append(toCompleteSeg, strings.Split(toComplete, "/")...)

	return &AutoCompleteContext{
		namespace:      namespace,
		query:          query,
		completeLeaves: completeLeaves,
		toComplete: &ToComplete{
			root:   toCompleteSeg[:len(toCompleteSeg)-1],
			leaves: strings.Split(toCompleteSeg[len(toCompleteSeg)-1], ","),
		},
	}
}

func (ctx *AutoCompleteContext) Next(fieldName, arg string) *AutoCompleteContext {
	query := append([]string{}, ctx.query...)
	query = append(query, fieldName)
	if arg != "" {
		query = append(query, arg)
	}

	return &AutoCompleteContext{
		namespace:      ctx.namespace,
		query:          query,
		completeLeaves: ctx.completeLeaves,
		toComplete:     ctx.toComplete,
	}
}

func (c Completion) Len() int {
	return len(c)
}

func (c Completion) Less(i, j int) bool {
	return len(c[i]) < len(c[j])
}

func (c Completion) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
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
	query, err := c.Schema.ParseQuery(listQuery, queryType, false)
	if err != nil {
		switch e := err.(type) {
		case *EnumValueNotFound:
			if e.Target != queryStr[len(queryStr)-1] {
				return nil, err
			}
			return e.Variants, nil
		case *FieldNotFound:
			if e.Target != queryStr[len(queryStr)-1] {
				return nil, err
			}
			return e.Fields, err
		case *LeafRequireArgument:
			if e.Leaf != queryStr[len(queryStr)-1] {
				return nil, err
			}
			return c.ListArguments(append(queryStr, ""), e.Argument)

		default:
			return nil, err
		}
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

func (c *CtrlClient) CompleteQuery(namespace, toComplete string, completeLeaves bool) ([]string, error) {
	namespaceType, err := c.Schema.MustGetType(NamespaceType)
	if err != nil {
		return nil, err
	}

	completion, err := c.completeQuery(NewAutoCompleteContext(namespace, toComplete, completeLeaves), namespaceType)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

func (c *CtrlClient) CompleteQueryBased(namespace, base, toComplete string, completeLeaves bool) ([]string, error) {
	if base == "" {
		return c.CompleteQuery(namespace, toComplete, completeLeaves)
	}

	queryType, err := c.GetQueryType()
	if err != nil {
		return nil, err
	}

	query := append([]string{NamespaceKey, namespace}, strings.Split(base, "/")...)

	root, err := c.Schema.ParseQuery(query, queryType, true)
	if err != nil {
		return nil, err
	}

	ctx := NewAutoCompleteContext(namespace, toComplete, completeLeaves)
	ctx.query = query
	for len(root.Fields) != 0 {
		for _, field := range root.Fields {
			root = field
			break
		}
	}

	completion, err := c.completeQuery(ctx, root.Type)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

func (c *CtrlClient) completeQueryObject(ctx *AutoCompleteContext, root *Type) ([]string, error) {
	var completions []string
	for _, field := range root.Fields {
		subType, err := c.Schema.resolve(&field.Type)
		if err != nil {
			return nil, err
		}

		if len(field.Args) == 0 {
			subCompletions, err := c.completeQuery(ctx.Next(string(field.Name), ""), subType)
			if err != nil {
				return nil, err
			}

			completions = append(completions, subCompletions...)
			continue
		}

		var args []string

		if typ, err := c.Schema.resolve(&field.Args[0].Type); err == nil && typ.Kind == EnumKind {
			for variant := range typ.EnumMap {
				args = append(args, variant)
			}
		} else {
			args, err = c.ListArguments(append(ctx.query, string(field.Name), ""), string(field.Args[0].Name))
			if err != nil {
				return nil, err
			}
		}

		for _, arg := range args {
			subCompletions, err := c.completeQuery(ctx.Next(string(field.Name), arg), subType)
			if err != nil {
				return nil, err
			}

			completions = append(completions, subCompletions...)
		}
	}
	return completions, nil
}

// accepts ScalarKind, EnumKind and ObjectKind
func (c *CtrlClient) completeQuery(ctx *AutoCompleteContext, root *Type) ([]string, error) {
	currentQuery := strings.Join(ctx.query, "/")
	toCompleteRoot := strings.Join(ctx.toComplete.root, "/")

	if len(ctx.query) <= len(ctx.toComplete.root) {
		if !strings.HasPrefix(toCompleteRoot, currentQuery) {
			return nil, nil
		}

		return c.completeQueryObject(ctx, root)
	}

	if !strings.HasPrefix(currentQuery, toCompleteRoot) {
		return nil, nil
	}

	if len(ctx.query)-len(ctx.toComplete.root) == 1 {
		var completes, leaves []string
		leaf := ctx.query[len(ctx.query)-1]
		complete := ctx.toComplete.Clone()
		leaves = append(leaves, complete.leaves[:len(complete.leaves)-1]...)

		leafMap := make(map[string]bool)
		for _, l := range leaves {
			leafMap[l] = true
		}

		if !leafMap[leaf] && strings.HasPrefix(leaf, complete.leaves[len(complete.leaves)-1]) {
			leaves = append(leaves, leaf)
		}
		if ctx.completeLeaves && root.Kind != ObjectKind && len(leaves) != 0 {
			complete.leaves = leaves
			completes = append(completes, complete.TrimNamespaced(ctx.namespace))
		} else if len(leaves) == 1 && root.Kind == ObjectKind {
			subCompletes, err := c.completeQueryObject(ctx, root)
			if err != nil {
				return nil, err
			}
			completes = append(completes, subCompletes...)
		}

		return completes, nil
	}

	if strings.HasPrefix(currentQuery, ctx.toComplete.ToQuery()) {
		return []string{trimNamespace(currentQuery, ctx.namespace)}, nil
	}

	return nil, nil
}

func trimNamespace(query, namespace string) string {
	newQuery := strings.TrimPrefix(query, strings.Join([]string{NamespaceKey, namespace}, "/"))
	return strings.Trim(newQuery, "/")
}
