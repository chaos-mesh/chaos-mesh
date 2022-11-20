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
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type Object struct {
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
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var (
		oldObj Object
		newObj Object
	)
	err = yaml.Unmarshal(data, &oldObj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	{
		var (
			schedule     yaml.MapSlice
			findSchedule bool
		)

		if isIncompatibleChaos(oldObj) {
			fmt.Printf("the define of %s is changed in v2.0, please modify it according to the document, refer to https://chaos-mesh.org/docs/\n", oldObj.Kind)
			os.Exit(1)
		}

		for _, item := range oldObj.Spec {
			if item.Key == "scheduler" {
				schedule = item.Value.(yaml.MapSlice)
				findSchedule = true
			}
		}

		if findSchedule {
			newObj = toScheduleObject(oldObj, schedule)
		} else {
			newObj = oldObj
			newObj.Spec = transformChaosSpec(oldObj)
		}
	}
	data, err = yaml.Marshal(newObj)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = os.WriteFile(os.Args[2], data, 0644)
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

func toScheduleObject(oldObj Object, schedule yaml.MapSlice) Object {
	var newObj Object
	var cron string
	for _, item := range schedule {
		if item.Key == "cron" {
			cron = item.Value.(string)
		}
	}
	newObj.ApiVersion = oldObj.ApiVersion
	newObj.Metadata = oldObj.Metadata
	newObj.Kind = "Schedule"
	newObj.Spec = append(newObj.Spec, yaml.MapItem{Key: "schedule", Value: cron})
	newObj.Spec = append(newObj.Spec, yaml.MapItem{Key: "type", Value: toNewKind(oldObj.Kind)})
	newObj.Spec = append(newObj.Spec, yaml.MapItem{Key: "historyLimit", Value: 5})
	newObj.Spec = append(newObj.Spec, yaml.MapItem{Key: "concurrencyPolicy", Value: "Forbid"})

	newSpec := transformChaosSpec(oldObj)
	newObj.Spec = append(newObj.Spec, yaml.MapItem{Key: getKeyName(oldObj.Kind), Value: newSpec})
	return newObj
}

func transformChaosSpec(obj Object) yaml.MapSlice {
	var (
		newSpec        yaml.MapSlice
		containerNames []string
	)
	for _, kv := range obj.Spec {
		if kv.Key == "scheduler" {
			continue
		}

		if kv.Key == "containerName" {
			containerNames = append(containerNames, kv.Value.(string))
			continue
		}

		if obj.Kind == "DNSChaos" {
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

func isIncompatibleChaos(obj Object) bool {
	incompatible := false
	switch obj.Kind {
	case "DNSChaos":
		for _, kv := range obj.Spec {
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
