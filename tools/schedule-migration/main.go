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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
)

type Item struct {
	ApiVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   yaml.MapSlice `yaml:"metadata"`
	Spec       yaml.MapSlice `yaml:"spec"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println(len(os.Args))
		fmt.Println("migrator <old-yaml> <new-yaml>")
		os.Exit(1)
	}
	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var (
		old Item
		new Item
	)
	err = yaml.Unmarshal(data, &old)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	{
		schedule, find := old.Spec.ToMap()["scheduler"]
		if !find {
			new = old
		} else {
			new.ApiVersion = old.ApiVersion
			new.Metadata = old.Metadata
			new.Kind = "Schedule"
			new.Spec = append(new.Spec, yaml.MapItem{Key: "schedule", Value: schedule.(map[string]interface{})["cron"]})
			new.Spec = append(new.Spec, yaml.MapItem{Key: "type", Value: old.Kind})
			new.Spec = append(new.Spec, yaml.MapItem{Key: "historyLimit", Value: 5})
			new.Spec = append(new.Spec, yaml.MapItem{Key: "concurrencyPolicy", Value: "Forbid"})
			var newSpec yaml.MapSlice
			for _, item := range old.Spec {
				if item.Key != "scheduler" {
					newSpec = append(newSpec, item)
				}
			}
			new.Spec = append(new.Spec, yaml.MapItem{Key: getKeyName(old.Kind), Value: newSpec})
		}

	}
	data, err = yaml.Marshal(new)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile(os.Args[2], data, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getKeyName(name string) string {
	s := strings.ToLower(name)
	return strings.ReplaceAll(s, "chaos", "_chaos")
}
