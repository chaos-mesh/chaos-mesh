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

package client

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/iancoleman/strcase"
)

const NamespaceKey = "namespace"
const NamespaceType = "Namespace"

type CtrlClient struct {
	cancel context.CancelFunc

	Client             *graphql.Client
	SubscriptionClient *graphql.SubscriptionClient
	Schema             *Schema
}

type AutoCompleteContext struct {
	context.Context
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
	query = append(query, strings.Join(c.leaves, SeperatorLeaf))
	return strings.Join(query, SeperatorSegment)
}

func (c *ToComplete) TrimNamespaced(namespace string) string {
	return trimNamespace(c.ToQuery(), namespace)
}

type Completion []string

func NewAutoCompleteContext(ctx context.Context, namespace, toComplete string, completeLeaves bool) *AutoCompleteContext {
	query := []string{argField(NamespaceKey, namespace)}
	toCompleteSeg := append([]string{}, query...)
	toCompleteSeg = append(toCompleteSeg, strings.Split(toComplete, SeperatorSegment)...)

	return &AutoCompleteContext{
		Context:        ctx,
		namespace:      namespace,
		query:          query,
		completeLeaves: completeLeaves,
		toComplete: &ToComplete{
			root:   toCompleteSeg[:len(toCompleteSeg)-1],
			leaves: strings.Split(toCompleteSeg[len(toCompleteSeg)-1], SeperatorLeaf),
		},
	}
}

func (ctx *AutoCompleteContext) Next(fieldName, arg string) *AutoCompleteContext {
	query := append([]string{}, ctx.query...)
	field := fieldName
	if arg != "" {
		field = argField(field, arg)
	}
	query = append(query, field)

	return &AutoCompleteContext{
		Context:        ctx.Context,
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
		Client:             graphql.NewClient(url, nil),
		SubscriptionClient: graphql.NewSubscriptionClient(url),
	}

	schemaQuery := new(struct {
		Schema RawSchema `graphql:"__schema"`
	})

	err := client.Client.Query(ctx, schemaQuery, nil)
	if err != nil {
		return nil, err
	}

	client.Schema = NewSchema(&schemaQuery.Schema)
	return client, nil
}

func (c *CtrlClient) GetQueryType() (*Type, error) {
	return c.Schema.MustGetType(string(c.Schema.QueryType.Name))
}

// list tail arguments, expected queryStr: ["prefix1", "prefix2", "resource"]
func (c *CtrlClient) ListArguments(ctx context.Context, queryStr []string, argumentName string) ([]string, error) {
	queryType, err := c.GetQueryType()
	if err != nil {
		return nil, err
	}

	listQuery := append([]string{}, queryStr...)
	listQuery[len(listQuery)-1] = tagField(queryStr[len(queryStr)-1], TagNameAll)
	listQuery = append(listQuery, argumentName)
	query, err := c.Schema.ParseQuery(listQuery, queryType, false)
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
	err = c.Client.Query(ctx, queryValue, variables.GenMap())
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

func (c *CtrlClient) ListNamespace(ctx context.Context) ([]string, error) {
	namespaceQuery := new(struct {
		Namespace []struct {
			Ns string
		}
	})

	err := c.Client.Query(ctx, namespaceQuery, nil)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, ns := range namespaceQuery.Namespace {
		namespaces = append(namespaces, ns.Ns)
	}

	return namespaces, nil
}

func (c *CtrlClient) CompleteRoot(ctx context.Context, namespace, toComplete string) ([]string, error) {
	namespaceType, err := c.Schema.MustGetType(NamespaceType)
	if err != nil {
		return nil, err
	}

	completion, err := c.completeRoot(NewAutoCompleteContext(ctx, namespace, toComplete, false), namespaceType)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

func (c *CtrlClient) CompleteQuery(ctx context.Context, namespace, toComplete string) ([]string, error) {
	namespaceType, err := c.Schema.MustGetType(NamespaceType)
	if err != nil {
		return nil, err
	}

	completion, err := c.completeQuery(NewAutoCompleteContext(ctx, namespace, toComplete, true), namespaceType)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

func (c *CtrlClient) CompleteQueryBased(ctx context.Context, namespace, base, toComplete string) ([]string, error) {
	if base == "" {
		return c.CompleteQuery(ctx, namespace, toComplete)
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

	completeCtx := NewAutoCompleteContext(ctx, namespace, toComplete, true)
	completeCtx.query = query
	for len(root.Fields) != 0 {
		for _, field := range root.Fields {
			root = field
			break
		}
	}

	completion, err := c.completeQuery(completeCtx, root.Type)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

func (c *CtrlClient) completeObject(ctx *AutoCompleteContext, root *Type, completer func(*AutoCompleteContext, *Type) ([]string, error)) ([]string, error) {
	var completions []string
	for _, field := range root.Fields {
		subType, err := c.Schema.resolve(&field.Type)
		if err != nil {
			return nil, err
		}

		if len(field.Args) == 0 {
			subCompletions, err := completer(ctx.Next(string(field.Name), ""), subType)
			if err != nil {
				return nil, err
			}

			completions = append(completions, subCompletions...)
			continue
		}

		var args []string
		var tag string

		if typ, err := c.Schema.resolve(&field.Args[0].Type); err == nil && typ.Kind == EnumKind {
			for variant := range typ.EnumMap {
				args = append(args, variant)
			}
		} else {
			query := append([]string{}, ctx.query...)
			args, err = c.ListArguments(ctx, append(query, string(field.Name)), string(field.Args[0].Name))
			if err != nil {
				return nil, err
			}
			tag = TagNameAll
		}

		for _, arg := range args {
			subCompletions, err := completer(ctx.Next(string(field.Name), arg), subType)
			if err != nil {
				return nil, err
			}

			completions = append(completions, subCompletions...)
		}

		if tag != "" {
			subCompletions, err := completer(ctx.Next(tagField(string(field.Name), tag), ""), subType)
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
	currentQuery := strings.Join(ctx.query, SeperatorSegment)
	toCompleteRoot := strings.Join(ctx.toComplete.root, SeperatorSegment)

	if len(ctx.query) <= len(ctx.toComplete.root) {
		if !strings.HasPrefix(toCompleteRoot, currentQuery) {
			return nil, nil
		}

		return c.completeObject(ctx, root, c.completeQuery)
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
			subCompletes, err := c.completeObject(ctx, root, c.completeQuery)
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

func (c *CtrlClient) completeRoot(ctx *AutoCompleteContext, root *Type) ([]string, error) {
	if root.Kind != ObjectKind || len(ctx.toComplete.leaves) != 1 {
		return nil, nil
	}

	currentQuery := strings.Join(ctx.query, "/")
	toCompleteRoot := strings.Join(ctx.toComplete.root, "/")

	if len(ctx.query) <= len(ctx.toComplete.root) {
		if !strings.HasPrefix(toCompleteRoot, currentQuery) {
			return nil, nil
		}

		return c.completeObject(ctx, root, c.completeRoot)
	}

	if strings.HasPrefix(currentQuery, ctx.toComplete.ToQuery()) {
		completes := []string{trimNamespace(currentQuery, ctx.namespace)}
		if len(ctx.query)-len(ctx.toComplete.root) == 1 {
			subCompletes, err := c.completeObject(ctx, root, c.completeRoot)
			if err != nil {
				return nil, err
			}
			completes = append(completes, subCompletes...)
		}
		return completes, nil
	}

	return nil, nil
}

func trimNamespace(query, namespace string) string {
	newQuery := strings.TrimPrefix(query, argField(NamespaceKey, namespace))
	return strings.Trim(newQuery, SeperatorSegment)
}

func tagField(field, tag string) string {
	return fmt.Sprintf("%s%s%s", field, SeperatorTag, tag)
}

func argField(field, arg string) string {
	return fmt.Sprintf("%s%s%s", field, SeperatorArgument, arg)
}
