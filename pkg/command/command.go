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

package command

import (
	"os/exec"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// ExecTag stands for the path of executable file in command.
// If we want this util works ,
// we must add Exec in the struct and use NewExec() to initialize it,
// because the default way to initialize Exec means None in code.
const ExecTag = "exec"

// SubCommandTag stands for the sub command in common command.
// We can use it in struct fields as a tag.
// Just like MatchExtension below
//
//	type Iptables Struct {
//		Exec
//		MatchExtension Match `sub_command:""`
//	}
//
//	type Match Struct {
//		Exec
//		Port string `para:"-p"`
//	}
//
// Field with SubcommandTag needs to be a struct with Exec.
const SubCommandTag = "sub_command"

// ParaTag stands for parameters in command.
// We can use it in struct fields as a tag.
// Just like Port below
//
//	type Iptables Struct {
//		Exec
//		Port string `para:"-p"`
//	}
//
// If the field is not string type or []string type , it will bring an error.
// If the tag value like "-p" is empty string ,
// the para will just add the field value into the command just as some single value parameter in command.
// If the value of field is empty string or empty string slice or empty slice, the field and tag will all be skipped.
const ParaTag = "para"

// Exec is the interface of a command.
// We need to inherit it in the struct of command.
// User must add ExecTag as the tag of Exec field.
// Example:
//
//	type Iptables struct {
//		Exec           `exec:"iptables"`
//		Tables         string `para:"-t"`
//	}
type Exec struct {
	active bool
}

func NewExec() Exec {
	return Exec{active: true}
}

func ToCommand(i interface{}) (*exec.Cmd, error) {
	path, args, err := Marshal(i)
	if err != nil {
		return nil, err
	}
	return exec.Command(path, args...), nil
}

func Marshal(i interface{}) (string, []string, error) {
	value := reflect.ValueOf(i)
	return marshal(value)
}

func marshal(value reflect.Value) (string, []string, error) {
	//var options []string
	if path, ok := SearchKey(value); ok {
		// Field(0).String is Exec.Path

		if path == "" {
			return "", nil, nil
		}
		args := make([]string, 0)
		for i := 0; i < value.NumField(); i++ {
			if _, ok := value.Type().Field(i).Tag.Lookup(SubCommandTag); ok {
				subPath, subArgs, err := marshal(value.Field(i))
				if err != nil {
					return "", nil, err
				}
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
				} else if value.Field(i).Kind() == reflect.Slice {
					if slicePara, ok := value.Field(i).Interface().([]string); ok {
						if strings.Join(slicePara, "") != "" {
							if paraName != "" {
								args = append(args, paraName)
							}
							args = append(args, slicePara...)
						}
					} else {
						return "", nil, errors.Errorf("invalid parameter slice type %s :parameter slice must be string slice", value.Field(i).String())
					}
				} else {
					return "", nil, errors.Errorf("invalid parameter type %s : parameter must be string or string slice", value.Type().Field(i).Type.Name())
				}
			}
		}
		return path, args, nil
	}
	return "", nil, nil
}

func SearchKey(value reflect.Value) (string, bool) {
	for i := 0; i < value.NumField(); i++ {
		if path, ok := value.Type().Field(i).Tag.Lookup(ExecTag); ok {
			if value.Field(i).Field(0).Bool() {
				return path, ok
			}
		}
	}
	return "", false
}
