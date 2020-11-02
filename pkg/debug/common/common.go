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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"

	kubectlscheme "k8s.io/kubectl/pkg/scheme"
)

var (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorGreen = "\033[32m"
	ColorCyan  = "\033[36m"
	ColorBlue  = "\033[34m"
	scheme     = runtime.NewScheme()
)

// ClientSet contains two different clients
type ClientSet struct {
	CtrlClient client.Client
	K8sClient  *kubernetes.Clientset
}

func init() {
	_ = v1alpha1.AddToScheme(scheme)
	_ = clientgoscheme.AddToScheme(scheme)
}

func upperCaseChaos(str string) string {
	parts := regexp.MustCompile("(.*)(chaos)").FindStringSubmatch(str)
	return strings.Title(parts[1]) + strings.Title(parts[2])
}

// Print prints result to users in prettier format, with number of tabs and color specified
func Print(s string, num int, color string) {
	var tabStr string
	for i := 0; i < num; i++ {
		tabStr += "\t"
	}
	s = string(color) + s + string(ColorReset)
	fmt.Printf("%s%s\n\n", tabStr, regexp.MustCompile("\n").ReplaceAllString(s, "\n"+tabStr))
}

// MarshalChaos returns json in readable format
func MarshalChaos(s interface{}) (string, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal indent: %s", err.Error())
	}
	return string(b), nil
}

// InitClientSet inits two different clients that would be used
func InitClientSet() (*ClientSet, error) {
	ctrlClient, err := client.New(config.GetConfigOrDie(), client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client")
	}
	k8sClient, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		return nil, fmt.Errorf("error in getting access to K8S: %s", err.Error())
	}
	return &ClientSet{ctrlClient, k8sClient}, nil
}

// Log gets information from log and returns the result
// runtime-controller only support CRUD， use client-go client
func Log(pod string, ns string, tail int64, c *kubernetes.Clientset) (string, error) {
	var podLogOpts corev1.PodLogOptions
	//use negative tail to indicate no tail limit is needed
	if tail < 0 {
		podLogOpts = corev1.PodLogOptions{}
	} else {
		podLogOpts = corev1.PodLogOptions{
			TailLines: func(i int64) *int64 { return &i }(tail),
		}
	}
	req := c.CoreV1().Pods(ns).GetLogs(pod, &podLogOpts)
	podLogs, err := req.Stream()
	if err != nil {
		return "", fmt.Errorf("failed to open stream: %s", err.Error())
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("failed to copy information from podLogs to buf: %s", err.Error())
	}
	return buf.String(), nil
}

// Exec executes certain command and returns the result
func Exec(pod string, ns string, cmd string, c *kubernetes.Clientset) (string, error) {
	req := c.CoreV1().RESTClient().Post().
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
	exec, err := remotecommand.NewSPDYExecutor(config.GetConfigOrDie(), "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("error in creating NewSPDYExecutor: %s", err.Error())
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", fmt.Errorf("error in creating StreamOptions: %s", err.Error())
	}
	if stderr.String() != "" {
		return stdout.String(), fmt.Errorf(stderr.String())
	}
	return stdout.String(), nil
}

// GetChaos returns the chaos that will do type-assertion
func GetChaos(ctx context.Context, chaosType string, chaosName string, ns string, c client.Client) (runtime.Object, error) {
	// get podName
	chaosType = upperCaseChaos(strings.ToLower(chaosType))
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

// GetPods returns pod list and corresponding chaos daemon
func GetPods(ctx context.Context, status v1alpha1.ChaosStatus, selector v1alpha1.SelectorSpec, c client.Client) ([]v1.Pod, []v1.Pod, error) {
	// get podName
	failedMessage := status.FailedMessage
	if failedMessage != "" {
		return nil, nil, fmt.Errorf("chaos failed with: %s", failedMessage)
	}

	phase := status.Experiment.Phase
	nextStart := status.Scheduler.NextStart

	if phase == "Waiting" {
		waitTime := nextStart.Sub(time.Now())
		fmt.Printf("Waiting for chaos to start, in %s\n", waitTime)
		time.Sleep(waitTime)
	}

	// TODO: failed and maybe not appropirate to create manager here, to get client.Reader
	// So could not parse fieldSelector for now
	pods, err := utils.SelectPods(ctx, c, nil, selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to SelectPods with: %s", err.Error())
	}

	var chaosDaemons []corev1.Pod
	// get chaos daemon
	for _, pod := range pods {
		nodeName := pod.Spec.NodeName
		daemonSelector := v1alpha1.SelectorSpec{
			Nodes:          []string{nodeName},
			LabelSelectors: map[string]string{"app.kubernetes.io/component": "chaos-daemon"},
		}
		daemons, err := utils.SelectPods(ctx, c, nil, daemonSelector)
		if err != nil || len(daemons) == 0 {
			return nil, nil, fmt.Errorf("fail to get daemon with: %s", err.Error())
		}
		// TODO: not sure about this, if only get one daemon is enough
		chaosDaemons = append(chaosDaemons, daemons[0])
	}

	return pods, chaosDaemons, nil
}

// GetChaosList returns chaos list limited by input
func GetChaosList(ctx context.Context, chaosType string, chaosName string, ns string, c client.Client) ([]string, error) {
	chaosType = upperCaseChaos(strings.ToLower(chaosType))
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
