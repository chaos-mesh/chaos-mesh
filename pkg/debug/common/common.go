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
	"time"

	"github.com/ghodss/yaml"
)

var (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorCyan  = "\033[36m"
)

type PodName struct {
	PodName                 string
	PodNamespace            string
	ChaosDaemonPodName      string
	ChaosDaemonPodNamespace string
}

func ExtractFromYaml(yamlStr string, str []string) (interface{}, error) {
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
	ret, ok := resultMap[str[len(str)-1]]
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
	cmd := fmt.Sprintf("kubectl get %s -n %s", chaosType, ns)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
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
				return nil, fmt.Errorf("extractFromGet failed with: %s", err.Error())
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

func GetPod(chaosType string, chaos string, ns string) (*PodName, error) {
	// get podName
	cmd := fmt.Sprintf("kubectl describe %s %s -n %s", chaosType, chaos, ns)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}

	failedMessage := regexp.MustCompile("(?:Failed Message:\\s*)(.*)").FindStringSubmatch(string(out))
	if len(failedMessage) != 0 {
		return nil, fmt.Errorf("chaos failed with: %s", failedMessage)
	}

	isRunning := regexp.MustCompile("(?:Next Recover:\\s)(.*)").MatchString(string(out))
	nextStart := regexp.MustCompile("(?:Next Start:\\s*)(.*)").FindStringSubmatch(string(out))[1]
	if isRunning == false {
		nextStartTime, err := time.Parse(time.RFC3339, nextStart)
		if err != nil {
			return nil, fmt.Errorf("time parsing next start failed: %s", err.Error())
		}
		waitTime := nextStartTime.Sub(time.Now())
		fmt.Printf("Waiting for chaos to start, for %v\n", waitTime)
		time.Sleep(waitTime)
	}

	podHier := []string{"Status", "Experiment", "Pod Records", "Name"}
	podName, err := ExtractFromYaml(string(out), podHier)
	if err != nil {
		return nil, fmt.Errorf("get podName from '%s' failed with: %s", cmd, err.Error())
	}
	podHier = []string{"Status", "Experiment", "Pod Records", "Namespace"}
	podNamespace, err := ExtractFromYaml(string(out), podHier)
	if err != nil {
		return nil, fmt.Errorf("get podNamespace from '%s' with: %s", cmd, err.Error())
	}

	// get nodeName
	cmd = fmt.Sprintf("kubectl get pods -o wide %s -n %s", podName, podNamespace)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	nodeName, err := ExtractFromGet(string(out), "NODE")
	if err != nil {
		return nil, fmt.Errorf("get nodeName from '%s' failed with: %s", cmd, err.Error())
	}

	// get chaos daemon
	cmd = "kubectl get pods -A -o wide"
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	title := strings.Split(string(out), "\n")[0]
	cmd = fmt.Sprintf("kubectl get pods -A -o wide | grep chaos-daemon | grep %s", nodeName)
	out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	chaosDaemonPodName, err := ExtractFromGet(title+"\n"+string(out), "NAME")
	if err != nil {
		return nil, fmt.Errorf("get chaos daemon name from '%s' failed with: %s", cmd, err.Error())
	}
	chaosDaemonPodNamespace, err := ExtractFromGet(title+"\n"+string(out), "NAMESPACE")
	if err != nil {
		return nil, fmt.Errorf("get chaos daemon namespace failed from '%s' with: %s", cmd, err.Error())
	}

	return &PodName{podName.(string), podNamespace.(string), chaosDaemonPodName, chaosDaemonPodNamespace}, nil
}

func PrintWithTab(s string) {
	fmt.Printf("\t%s\n", regexp.MustCompile("\n").ReplaceAllString(s, "\n\t"))
}
