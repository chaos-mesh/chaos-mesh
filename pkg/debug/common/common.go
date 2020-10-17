// Copyright 2019 Chaos Mesh Authors.
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
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

var (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
)

func ExtractFromYaml(yamlStr string, str []string) (string, error) {
	resultMap := make(map[string]interface{})
	// remove parts that could not been parsed to yaml
	for {
		err := yaml.Unmarshal([]byte(yamlStr), &resultMap)
		if err == nil {
			break
		}
		errStr := err.Error()
		r, _ := regexp.Compile("(?:line )(.*)(?::)")
		lineIndex := r.FindStringSubmatch(errStr)[1]
		lineIndexInt, err := strconv.Atoi(lineIndex)
		if err != nil {
			return "", fmt.Errorf("could not avoid error to parse yaml")
		}
		rm := strings.Split(yamlStr, "\n")[lineIndexInt-1]
		yamlStr = string(bytes.Replace([]byte(yamlStr), []byte(rm+"\n"), []byte(""), 1))
	}
	for i := 0; i < len(str)-1; i++ {
		var ok bool
		resultMap, ok = resultMap[str[i]].(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("wrong hierarchy: %s", str[i])
		}
	}
	ret, ok := resultMap[str[len(str)-1]].(string)
	if !ok {
		return "", fmt.Errorf("wrong hierarchy: %s", str[len(str)-1])
	}
	return ret, nil
}

func ExtractFromGet(str string, column string) (string, error) {
	column += " " // avoid only get prefix
	lines := strings.Split(str, "\n")
	nodeIndex := strings.Index(lines[0], column)
	if nodeIndex == -1 {
		return "", fmt.Errorf("could not found column: %s", column)
	}
	if len(lines) <= 1 || len(lines[1]) < nodeIndex {
		return "", fmt.Errorf("could not found column: %s", column)
	}
	nodeName := strings.Split(string(lines[1][nodeIndex:]), " ")[0]
	return nodeName, nil
}

func Debug(chaosType string, chaos string, ns string) ([]string, error) {
	chaosType = strings.ToLower(chaosType)
	out, err := exec.Command("kubectl", "get", chaosType, "-n", ns).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run command 'kubectl get %s' failed with: %s", chaosType, err.Error())
	}

	lines := strings.Split(string(out), "\n")
	var chaosList []string
	if lines[1] == "" {
		fmt.Printf(string(out))
	} else {
		title := lines[0]
		chaosNum := 0
		for i := 1; i < len(lines)-1; i++ {
			chaosName, err := ExtractFromGet(title+"\n"+lines[i], "NAME")
			if err != nil {
				return nil, fmt.Errorf("ExtractFromGet failed with: %s", err.Error())
			}
			if chaos == "" || chaos == chaosName {
				chaosList = append(chaosList, chaosName)
				chaosNum++
			}
		}
		if chaosNum == 0 {
			return nil, fmt.Errorf("no chaos is found, please check your input")
		}
	}
	return chaosList, nil
}
