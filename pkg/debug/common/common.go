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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"

	kubectlscheme "k8s.io/kubectl/pkg/scheme"
)

var (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorCyan  = "\033[36m"
	scheme     = runtime.NewScheme()
)

type PodName struct {
	PodName              string
	PodNamespace         string
	ChaosDaemonName      string
	ChaosDaemonNamespace string
}

func init() {
	_ = v1alpha1.AddToScheme(scheme)
	_ = clientgoscheme.AddToScheme(scheme)
}

func UpperCaseChaos(str string) string {
	parts := regexp.MustCompile("(.*)(chaos)").FindStringSubmatch(str)
	return strings.Title(parts[1]) + strings.Title(parts[2])
}

// ExtractFromJson extract certain item from given string slice
// String means the key of yaml
// Number means the order in a slice
func ExtractFromJson(chaos runtime.Object, str []string) (interface{}, error) {
	resultMap := make(map[string]interface{})
	b, err := json.Marshal(chaos)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %s", err.Error())
	}
	err = json.Unmarshal(b, &resultMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal failed: %s", err.Error())
	}
	for i := 0; i < len(str)-1; i++ {
		var ok bool
		num, err := strconv.Atoi(str[i+1])
		if err == nil {
			resultMap, ok = resultMap[str[i]].([]interface{})[num].(map[string]interface{})
			i++
		} else {
			resultMap, ok = resultMap[str[i]].(map[string]interface{})
		}
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

func Debug(chaosType string, chaosName string, ns string) ([]string, error) {
	options := client.Options{
		Scheme: scheme,
	}
	c, err := client.New(config.GetConfigOrDie(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create client")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chaosType = UpperCaseChaos(strings.ToLower(chaosType))
	allKinds := v1alpha1.AllKinds()
	chaosListIntf := allKinds[chaosType].ChaosList

	if err := c.List(ctx, chaosListIntf, client.InNamespace(ns)); err != nil {
		return nil, fmt.Errorf("failed to get chaosList: %s", err.Error())
	}
	chaosList := chaosListIntf.ListChaos()
	if len(chaosList) == 0 {
		return nil, fmt.Errorf("no chaos is found, please check your input")
	}

	var retList []string
	chaosNum := 0
	for _, ch := range chaosList {
		if chaosName == "" || chaosName == ch.Name {
			retList = append(retList, ch.Name)
			chaosNum++
		}
	}
	if chaosNum == 0 {
		return nil, fmt.Errorf("no chaos is found, please check your input")
	}
	return retList, nil
}

func GetChaos(chaosType string, chaosName string, ns string) (runtime.Object, error) {
	options := client.Options{
		Scheme: scheme,
	}
	c, err := client.New(config.GetConfigOrDie(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create client")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get podName
	chaosType = UpperCaseChaos(strings.ToLower(chaosType))
	allKinds := v1alpha1.AllKinds()
	chaos := allKinds[chaosType].Chaos
	objectKey := client.ObjectKey{
		Namespace: ns,
		Name:      chaosName,
	}
	if err := c.Get(ctx, objectKey, chaos); err != nil {
		return nil, fmt.Errorf("failed to get chaos %s: %s", chaosName, err.Error())
	}
	return chaos, nil
}

func GetLog(pod string, ns string, tail int64) (string, error) {
	var podLogOpts corev1.PodLogOptions
	if tail < 0 {
		podLogOpts = corev1.PodLogOptions{}
	} else {
		podLogOpts = corev1.PodLogOptions{
			TailLines: func(i int64) *int64 { return &i }(tail),
		}
	}
	// runtime-controller not support getlog for now
	// use client-go client
	clientset, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		return "", fmt.Errorf("failed to access to K8S: %v", err.Error())
	}
	req := clientset.CoreV1().Pods(ns).GetLogs(pod, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		return "", fmt.Errorf("failed to open stream: %v", err.Error())
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("failed to copy information from podLogs to buf: %v", err.Error())
	}
	return buf.String(), nil
}

func ExecCommand(pod string, ns string, cmd string) (string, error) {
	// creates the clientset
	configK8S := config.GetConfigOrDie()
	clientset, err := kubernetes.NewForConfig(configK8S)
	if err != nil {
		return "", fmt.Errorf("error in getting access to K8S: %v", err.Error())
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Command: []string{"/bin/sh", "-c", cmd},
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, kubectlscheme.ParameterCodec)

	var stdout, stderr bytes.Buffer
	exec, err := remotecommand.NewSPDYExecutor(configK8S, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("error in creating NewSPDYExecutor: %v", err.Error())
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", fmt.Errorf("error in creating StreamOptions: %v", err.Error())
	}
	if stderr.String() != "" {
		return stdout.String(), fmt.Errorf(stderr.String())
	}
	return stdout.String(), nil
}

func GetPod(chaosType string, chaosName string, ns string) (*PodName, error) {
	options := client.Options{
		Scheme: scheme,
	}
	c, err := client.New(config.GetConfigOrDie(), options)
	if err != nil {
		return nil, fmt.Errorf("failed to create client")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// get podName
	chaos, err := GetChaos(chaosType, chaosName, ns)
	if err != nil {
		return nil, fmt.Errorf("failed to get chaos %s: %s", chaosName, err.Error())
	}

	failedMessageHier := []string{"status", "failedMessage"}
	failedMessage, err := ExtractFromJson(chaos, failedMessageHier)
	if err == nil {
		return nil, fmt.Errorf("chaos failed with: %s", failedMessage)
	}

	phaseHier := []string{"status", "experiment", "phase"}
	phase, err := ExtractFromJson(chaos, phaseHier)
	if err != nil {
		return nil, fmt.Errorf("failed to get chaos phase with: %s", err.Error())
	}

	nextStartHier := []string{"status", "scheduler", "nextStart"}
	nextStart, err := ExtractFromJson(chaos, nextStartHier)
	if err != nil {
		return nil, fmt.Errorf("failed to get chaos phase with: %s", err.Error())
	}

	if phase.(string) == "Waiting" {
		nextStartTime, err := time.Parse(time.RFC3339, nextStart.(string))
		if err != nil {
			return nil, fmt.Errorf("time parsing next start failed: %s", err.Error())
		}
		waitTime := nextStartTime.Sub(time.Now())
		fmt.Printf("Waiting for chaos to start, in %v\n", waitTime)
		time.Sleep(waitTime)
	}

	podNameHier := []string{"status", "experiment", "podRecords", "0", "name"}
	podName, err := ExtractFromJson(chaos, podNameHier)
	if err != nil {
		return nil, fmt.Errorf("get podName failed with: %s", err.Error())
	}
	podNamespaceHier := []string{"status", "experiment", "podRecords", "0", "namespace"}
	podNamespace, err := ExtractFromJson(chaos, podNamespaceHier)
	if err != nil {
		return nil, fmt.Errorf("get podNamespace with: %s", err.Error())
	}

	// get nodeName
	pod := &corev1.Pod{}
	objectKey := client.ObjectKey{
		Namespace: podNamespace.(string),
		Name:      podName.(string),
	}
	if err = c.Get(ctx, objectKey, pod); err != nil {
		return nil, fmt.Errorf("failed to get pod %s: %s", podName, err.Error())
	}
	nodeName := pod.Spec.NodeName

	// get chaos daemon
	podList := &corev1.PodList{}

	listOptions := (&client.ListOptions{}).ApplyOptions([]client.ListOption{
		client.MatchingFields{"spec.nodeName": nodeName},
		client.MatchingLabels{"app.kubernetes.io/component": "chaos-daemon"},
	})

	if err = c.List(ctx, podList, listOptions); err != nil || len(podList.Items) == 0 {
		return nil, fmt.Errorf("failed to get podList: %s", err.Error())
	}
	ChaosDaemonName := podList.Items[0].GetObjectMeta().GetName()
	ChaosDaemonNamespace := podList.Items[0].GetObjectMeta().GetNamespace()

	return &PodName{podName.(string), podNamespace.(string), ChaosDaemonName, ChaosDaemonNamespace}, nil
}

func PrintWithTab(s string) {
	fmt.Printf("\t%s\n", regexp.MustCompile("\n").ReplaceAllString(s, "\n\t"))
}
