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

package cmd

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

func NewQueryCmd() *cobra.Command {
	var namespace string
	var resource string

	// queryCmd represents the query command
	var queryCmd = &cobra.Command{
		Use:   "get [QUERY]",
		Short: "get the target resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			// TODO: input ns by args
			cancel, port, err := common.ForwardCtrlServer(ctx, nil)
			if err != nil {
				return err
			}
			defer cancel()

			client, err := common.NewCtrlClient(ctx, fmt.Sprintf("http://127.0.0.1:%d/query", port))
			if err != nil {
				return fmt.Errorf("fail to init ctrl client: %s", err)
			}

			if client.Schema == nil {
				return errors.New("fail to fetch schema")
			}

			queryType, err := client.Schema.MustGetType(string(client.Schema.QueryType.Name))
			if err != nil {
				return err
			}

			var query *common.Query
			prefix := append([]string{"namespace"}, common.StandardizeQuery(namespace)...)

			if len(prefix) != 2 {
				return fmt.Errorf("invalid namepsace: %s", namespace)
			}

			if resource != "" {
				prefix = append(prefix, common.StandardizeQuery(resource)...)
			}

			for _, arg := range args {
				part := append(prefix, common.StandardizeQuery(arg)...)
				partQuery, err := client.Schema.ParseQuery(part, queryType)
				if err != nil {
					return err
				}

				query, err = query.Merge(partQuery)
				if err != nil {
					return err
				}
			}

			superQuery := common.NewQuery("query", queryType, nil)
			superQuery.Fields["namespace"] = query
			variables := common.NewVariables()

			queryStruct, err := client.Schema.Reflect(superQuery, variables)
			if err != nil {
				return err
			}

			queryValue := reflect.New(queryStruct.Elem()).Interface()
			err = client.Client.Query(ctx, queryValue, variables.GenMap())
			if err != nil {
				return err
			}

			prefixQuery, err := client.Schema.ParseQuery(prefix, queryType)
			if err != nil {
				return err
			}

			value, err := findResource(queryValue, prefixQuery)
			if err != nil {
				return err
			}

			data, err := yaml.Marshal(value)
			if err != nil {
				return err
			}

			fmt.Println(string(data))

			return nil
		},
	}

	queryCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "the kubenates namespace")
	queryCmd.Flags().StringVarP(&resource, "resource", "r", "", "the target resource")

	return queryCmd
}

func findResource(object interface{}, resource *common.Query) (interface{}, error) {
	if resource == nil {
		return object, nil
	}

	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Ptr:
		return findResource(value.Elem().Interface(), resource)
	case reflect.Struct:
		field := value.FieldByName(strcase.ToCamel(resource.Name))
		if field == *new(reflect.Value) {
			return nil, fmt.Errorf("cannot find field %s in object: %#v", resource.Name, object)
		}
		for _, f := range resource.Fields {
			return findResource(field.Interface(), f)
		}
		return findResource(field.Interface(), nil)
	case reflect.Slice:
		slice := make([]interface{}, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			elem, err := findResource(value.Index(i).Interface(), resource)
			if err != nil {
				return nil, err
			}
			slice = append(slice, elem)
		}
		return slice, nil
	default:
		return nil, fmt.Errorf("resource of %s kind is not supported", value.Kind())
	}
}
