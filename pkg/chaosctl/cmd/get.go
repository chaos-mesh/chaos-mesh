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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	ctrlclient "github.com/chaos-mesh/chaos-mesh/pkg/ctrl/client"
)

const QueryKey = "query"

func NewQueryCmd(log logr.Logger) *cobra.Command {
	var namespace string
	var root string

	var joinPrefix = func() ([]string, error) {
		segment := []string{ctrlclient.NamespaceKey}
		if namespace != "" {
			segment = append(segment, namespace)
		}
		prefix := append([]string{}, strings.Join(segment, ":"))
		if root != "" {
			prefix = append(prefix, ctrlclient.StandardizeQuery(root)...)
		}
		return prefix, nil
	}

	// getCmd represents the get command
	var getCmd = &cobra.Command{
		Use:   "get [resources]",
		Short: "get the target resources",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			ctx := context.Background()
			client, cancel, err := createClient(ctx)
			if err != nil {
				log.Error(err, "fail to create client")
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			defer cancel()

			completion, err := client.CompleteQueryBased(namespace, root, toComplete)
			if err != nil {
				log.Error(err, "fail to complete query")
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return completion, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctrlclient.DisableRuntimeErrorHandler()
			ctx := context.Background()
			client, cancel, err := createClient(ctx)
			if err != nil {
				return err
			}
			defer cancel()

			queryType, err := client.GetQueryType()
			if err != nil {
				return err
			}

			var query *ctrlclient.Query

			prefix, err := joinPrefix()
			if err != nil {
				return err
			}

			for _, arg := range args {
				part := append(prefix, ctrlclient.StandardizeQuery(arg)...)
				partQuery, err := client.Schema.ParseQuery(part, queryType, false)
				if err != nil {
					return err
				}

				query, err = query.Merge(partQuery)
				if err != nil {
					return err
				}
			}

			if query == nil {
				return fmt.Errorf("resource list is empty")
			}

			superQuery := ctrlclient.NewQuery(QueryKey, queryType, nil)
			superQuery.Fields[ctrlclient.NamespaceKey] = query
			variables := ctrlclient.NewVariables()

			queryStruct, err := client.Schema.Reflect(superQuery, variables)
			if err != nil {
				return err
			}

			queryValue := reflect.New(queryStruct.Elem()).Interface()
			rawData, err := client.Client.QueryRaw(ctx, queryValue, variables.GenMap())
			if err != nil {
				return err
			}

			json.Unmarshal(*rawData, queryValue)
			prefixQuery, err := client.Schema.ParseQuery(prefix, queryType, true)
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

	getCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "the kubenates namespace")
	getCmd.Flags().StringVarP(&root, "root", "r", "", "the root resource")
	getCmd.RegisterFlagCompletionFunc("namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		client, cancel, err := createClient(ctx)
		if err != nil {
			log.Error(err, "fail to create client")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		defer cancel()

		completion, err := client.ListNamespace()
		if err != nil {
			log.Error(err, "fail to complete resource")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completion, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})
	getCmd.RegisterFlagCompletionFunc("root", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		ctx := context.Background()
		client, cancel, err := createClient(ctx)
		if err != nil {
			log.Error(err, "fail to create client")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		defer cancel()

		completion, err := client.CompleteRoot(namespace, toComplete)
		if err != nil {
			log.Error(err, "fail to complete resource")
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return completion, cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
	})

	return getCmd
}

func createClient(ctx context.Context) (*ctrlclient.CtrlClient, context.CancelFunc, error) {
	cancel, port, err := ctrlclient.ForwardCtrlServer(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	client, err := ctrlclient.NewCtrlClient(ctx, fmt.Sprintf("http://127.0.0.1:%d/query", port))
	if err != nil {
		return nil, nil, fmt.Errorf("fail to init ctrl client: %s", err)
	}

	if client.Schema == nil {
		return nil, nil, errors.New("fail to fetch schema")
	}

	return client, cancel, nil
}

func findResource(object interface{}, resource *ctrlclient.Query) (interface{}, error) {
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
