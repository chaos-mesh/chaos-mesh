package utils

import (
	"os/exec"
	"reflect"
	"strings"
)

const ExecTag = "exec"
const SubCommandTag = "sub_command"
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
							args = append(args, paraName)
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
