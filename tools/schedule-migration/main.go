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
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Item struct {
	ApiVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   yaml.MapSlice `yaml:"metadata"`
	Spec       yaml.MapSlice `yaml:"spec"`
}

func main() {
	if len(os.Args) != 3 {
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
		var (
			schedule     yaml.MapSlice
			findSchedule bool
		)

		if isIncompatibleChaos(old) {
			fmt.Printf("the define of %s is changed in v2.0, please modify it according to the document, refer to https://chaos-mesh.org/docs/\n", old.Kind)
			os.Exit(1)
		}

		for _, item := range old.Spec {
			if item.Key == "scheduler" {
				schedule = item.Value.(yaml.MapSlice)
				findSchedule = true
			}
		}

		if findSchedule {
			new = toScheduleObject(old, schedule)
		} else {
			new = old
			new.Spec = transformChaosSpec(old)
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
	return strings.ReplaceAll(s, "chaos", "Chaos")
}

func toNewKind(kind string) string {
	if kind == "IoChaos" {
		return "IOChaos"
	}
	return kind
}

func toScheduleObject(old Item, schedule yaml.MapSlice) Item {
	var new Item
	var cron string
	for _, item := range schedule {
		if item.Key == "cron" {
			cron = item.Value.(string)
		}
	}
	new.ApiVersion = old.ApiVersion
	new.Metadata = old.Metadata
	new.Kind = "Schedule"
	new.Spec = append(new.Spec, yaml.MapItem{Key: "schedule", Value: cron})
	new.Spec = append(new.Spec, yaml.MapItem{Key: "type", Value: toNewKind(old.Kind)})
	new.Spec = append(new.Spec, yaml.MapItem{Key: "historyLimit", Value: 5})
	new.Spec = append(new.Spec, yaml.MapItem{Key: "concurrencyPolicy", Value: "Forbid"})

	newSpec := transformChaosSpec(old)
	new.Spec = append(new.Spec, yaml.MapItem{Key: getKeyName(old.Kind), Value: newSpec})
	return new
}

func transformChaosSpec(item Item) yaml.MapSlice {
	var (
		newSpec        yaml.MapSlice
		containerNames []string
	)
	for _, kv := range item.Spec {
		if kv.Key == "scheduler" {
			continue
		}

		if kv.Key == "containerName" {
			containerNames = append(containerNames, kv.Value.(string))
			continue
		}

		if item.Kind == "DNSChaos" {
			// 'scope' is obsolete in v2.0, and instead with 'patterns'
			// if 'scope' is 'all', means chaos applies to all the host
			// so it equal to pattern "*". otherwise, can't transform.
			if kv.Key == "scope" && kv.Value == "all" {
				patterns := []string{"*"}
				kv = yaml.MapItem{Key: "patterns", Value: patterns}
			}
		}

		newSpec = append(newSpec, kv)
	}

	if len(containerNames) != 0 {
		newSpec = append(newSpec, yaml.MapItem{Key: "containerNames", Value: containerNames})
	}

	return newSpec
}

func isIncompatibleChaos(item Item) bool {
	incompatible := false
	switch item.Kind {
	case "DNSChaos":
		for _, kv := range item.Spec {
			// 'scope' is obsolete in v2.0, and instead with 'patterns'
			// if 'scope' is 'all', means chaos applies to all the host
			// so it equal to pattern "*". otherwise, can't transform.
			if kv.Key == "scope" && kv.Value != "all" {
				incompatible = true
			}
		}
	default:
	}

	return incompatible
}
