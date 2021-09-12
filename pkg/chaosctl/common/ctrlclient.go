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
	prmt "github.com/gitchander/permutation"
	comb "github.com/gitchander/permutation/combination"
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
	maxRecurLevel int
	visitedTypes  map[string]bool
	query         []string
	leaves        int
}

type Completion []string

func NewAutoCompleteContext(namespace string, level, leaves int) *AutoCompleteContext {
	return &AutoCompleteContext{
		maxRecurLevel: level,
		visitedTypes:  make(map[string]bool),
		query:         []string{NamespaceKey, namespace},
		leaves:        leaves,
	}
}

func (ctx *AutoCompleteContext) IsComplete() bool {
	return ctx.maxRecurLevel == 0
}

func (ctx *AutoCompleteContext) Visited(typ *Type) bool {
	if typ.Kind != ObjectKind {
		return false
	}
	return ctx.visitedTypes[string(typ.Name)]
}

func (ctx *AutoCompleteContext) Next(typename, fieldName, arg string) *AutoCompleteContext {
	types := map[string]bool{
		typename: true,
	}

	query := make([]string, 0)

	for name := range ctx.visitedTypes {
		types[name] = true
	}

	query = append(query, ctx.query...)
	query = append(query, fieldName)
	if arg != "" {
		query = append(query, arg)
	}

	return &AutoCompleteContext{
		maxRecurLevel: ctx.maxRecurLevel - 1,
		visitedTypes:  types,
		leaves:        ctx.leaves,
		query:         query,
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

func (c *CtrlClient) CompleteQuery(namespace string, leaves int) ([]string, error) {
	namespaceType, err := c.Schema.MustGetType(NamespaceType)
	if err != nil {
		return nil, err
	}

	completion, err := c.completeQuery(NewAutoCompleteContext(namespace, 6, leaves), namespaceType)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

func (c *CtrlClient) CompleteQueryBased(namespace string, base string, leaves int) ([]string, error) {
	if base == "" {
		return c.CompleteQuery(namespace, leaves)
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

	ctx := NewAutoCompleteContext(namespace, 6, leaves)
	ctx.query = query
	for len(root.Fields) != 0 {
		ctx.visitedTypes[string(root.Type.Name)] = true
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

func (c *CtrlClient) CompleteResource(namespace string) ([]string, error) {
	namespaceType, err := c.Schema.MustGetType(NamespaceType)
	if err != nil {
		return nil, err
	}

	completion, err := c.completeResource(NewAutoCompleteContext(namespace, 6, 0), namespaceType)
	if err != nil {
		return nil, err
	}

	return completion, nil
}

// accepts ScalarKind, EnumKind and ObjectKind
func (c *CtrlClient) completeQuery(ctx *AutoCompleteContext, root *Type) ([]string, error) {
	if ctx.IsComplete() {
		return nil, nil
	}

	switch root.Kind {
	case ScalarKind, EnumKind:
		return nil, nil
	case ListKind, NonNullKind:
		return nil, fmt.Errorf("type is not supported to complete: %#v", root)
	}

	var trunks, leaves []string
	for _, field := range root.Fields {
		subType, err := c.Schema.resolve(&field.Type)
		if err != nil {
			return nil, err
		}

		if ctx.Visited(subType) {
			continue
		}

		if len(field.Args) == 0 {
			subQueries, err := c.completeQuery(ctx.Next(string(subType.Name), string(field.Name), ""), subType)
			if err != nil {
				return nil, err
			}

			if subQueries == nil {
				// this field is a leaf
				// or rearching the max recursion levels
				leaves = append(leaves, string(field.Name))
				continue
			}

			for _, subQuery := range subQueries {
				trunks = append(trunks, strings.Join([]string{string(field.Name), subQuery}, "/"))
			}
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
			subQueries, err := c.completeQuery(ctx.Next(string(subType.Name), string(field.Name), arg), subType)
			if err != nil {
				return nil, err
			}

			for _, subQuery := range subQueries {
				trunks = append(trunks, strings.Join([]string{string(field.Name), arg, subQuery}, "/"))
			}
			continue
		}
	}

	var queries []string
	for _, leafPrmt := range fullPermutation(leaves, ctx.leaves) {
		queries = append(queries, strings.Join(leafPrmt, ","))
	}

	queries = append(queries, trunks...)
	return queries, nil
}

// accepts ObjectKind only
func (c *CtrlClient) completeResource(ctx *AutoCompleteContext, root *Type) ([]string, error) {
	if ctx.IsComplete() {
		return nil, nil
	}

	var resources []string
	for _, field := range root.Fields {
		subType, err := c.Schema.resolve(&field.Type)
		if err != nil {
			return nil, err
		}

		if ctx.Visited(subType) {
			continue
		}

		if subType.Kind != ObjectKind {
			continue
		}

		if len(field.Args) == 0 {
			resources = append(resources, string(field.Name))
			subResources, err := c.completeResource(ctx.Next(string(subType.Name), string(field.Name), ""), subType)
			if err != nil {
				return nil, err
			}

			for _, subResource := range subResources {
				resources = append(resources, strings.Join([]string{string(field.Name), subResource}, "/"))
			}
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
			resources = append(resources, strings.Join([]string{string(field.Name), arg}, "/"))
			subResources, err := c.completeResource(ctx.Next(string(subType.Name), string(field.Name), arg), subType)
			if err != nil {
				return nil, err
			}

			for _, subResource := range subResources {
				resources = append(resources, strings.Join([]string{string(field.Name), arg, subResource}, "/"))
			}
			continue
		}
	}

	return resources, nil
}

func fullPermutation(strs []string, leaves int) [][]string {
	var results [][]string

	maxLeaves := len(strs)
	if leaves < maxLeaves {
		maxLeaves = leaves
	}

	for i := 1; i <= maxLeaves; i++ {
		substrs := make([]string, i)
		var (
			n = len(strs)    // length of set
			k = len(substrs) // length of subset
		)

		c := comb.New(n, k)
		p := prmt.New(prmt.StringSlice(substrs))

		for c.Next() {
			// fill substrs by indexes
			for subsetIndex, setIndex := range c.Indexes() {
				substrs[subsetIndex] = strs[setIndex]
			}

			for p.Next() {
				results = append(results, append(make([]string, 0, len(substrs)), substrs...))
			}
		}
	}

	return results
}
