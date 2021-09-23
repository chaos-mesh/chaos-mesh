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

package utils

import (
	"os/exec"
	"reflect"
	"strings"
)

// ExecTag stand for the path of executable file in command.
// If we want this util works ,
// we must add Exec in the struct and use NewExec() to initialize it,
// because the default way to initialize Exec means None in code.
const ExecTag = "exec"

// SubCommandTag stand for the sub command in common command.
// We can use it in struct fields as a tag.
// Just like MatchExtension below
// type Iptables Struct {
// 	Exec
//	MatchExtension string `sub_command:""`
// }
// Field with SubcommandTag needs to be a struct with Exec.
const SubCommandTag = "sub_command"

// ParaTag stand for parameters in command.
// We can use it in struct fields as a tag.
// Just like Port below
// type Iptables Struct {
// 	Exec
//	Port string `para:"-p"`
// }
// If the field is not string type or []string type , it will be skipped.
// If the tag value like "-p" is empty string ,
// the para will just add the field value into the command just as some single value parameter in command.
// If the value of field is empty string or empty string slice or empty slice, the field and tag will all be skipped.
const ParaTag = "para"

type Exec struct {
	option string
}

func NewExec() Exec {
	return Exec{option: "OK"}
}

func ToCommand(i interface{}) *exec.Cmd {
	path, args := Unmarshal(i)
	return exec.Command(path, args...)
}

func Unmarshal(i interface{}) (string, []string) {
	value := reflect.ValueOf(i)
	return unmarshal(value)
}

func unmarshal(value reflect.Value) (string, []string) {
	//var options []string
	if path, ok := SearchKey(value); ok {
		// Field(0).String is Exec.Path

		if path == "" {
			return "", nil
		}
		args := make([]string, 0)
		for i := 0; i < value.NumField(); i++ {
			if _, ok := value.Type().Field(i).Tag.Lookup(SubCommandTag); ok {
				subPath, subArgs := unmarshal(value.Field(i))
				if subPath != "" {
					args = append(args, subPath)
				}
				args = append(args, subArgs...)
			}
			if paraName, ok := value.Type().Field(i).Tag.Lookup(ParaTag); ok {
				if value.Type().Field(i).Type.Name() == "string" {
					if value.Field(i).String() != "" {
						if paraName != "" {
							args = append(args, paraName)
						}
						args = append(args, value.Field(i).String())
					}
				}
				if value.Field(i).Kind() == reflect.Slice {
					if slicePara, ok := value.Field(i).Interface().([]string); ok {
						if strings.Join(slicePara, "") != "" {
							if paraName != "" {
								args = append(args, paraName)
							}
							args = append(args, slicePara...)
						}

					}
				}
			}
		}
		return path, args
	} else {
		return "", nil
	}
}

func SearchKey(value reflect.Value) (string, bool) {
	for i := 0; i < value.NumField(); i++ {
		if path, ok := value.Type().Field(i).Tag.Lookup(ExecTag); ok {
			if value.Field(i).Field(0).String() == "" {
				return "", false
			}
			return path, ok
		}
	}
	return "", false
}
