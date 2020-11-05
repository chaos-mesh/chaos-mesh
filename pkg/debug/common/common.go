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
	"log"
	"regexp"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
	e2econfig "github.com/chaos-mesh/chaos-mesh/test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util/portforward"

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

// Exec executes certain command and returns the result
// runtime-controller only support CRUD， use client-go client
func Exec(pod string, ns string, cmd string, c *kubernetes.Clientset) (string, error) {
	req := c.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
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

	var chaosDaemons []v1.Pod
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
		chaosDaemons = append(chaosDaemons, daemons[0])
	}

	return pods, chaosDaemons, nil
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

// GetPidFromPod returns pid given containerd ID in pod
func GetPidFromPod(ctx context.Context, pod v1.Pod, daemon v1.Pod) (int, error) {
	pfCancel, localPort, err := forwardPorts(ctx, daemon, uint16(common.ControllerCfg.ChaosDaemonPort))
	if err != nil {
		return 0, fmt.Errorf("forward ports failed: %s", err.Error())
	}

	daemonClient, err := utils.NewChaosDaemonClientLocally(int(localPort))
	if err != nil {
		return 0, fmt.Errorf("new chaos daemon client failed: %s", err.Error())
	}
	defer daemonClient.Close()

	if len(pod.Status.ContainerStatuses) == 0 {
		return 0, fmt.Errorf("%s %s can't get the state of container", pod.Namespace, pod.Name)
	}

	res, err := daemonClient.ContainerGetPid(ctx, &pb.ContainerRequest{
		Action: &pb.ContainerAction{
			Action: pb.ContainerAction_GETPID,
		},
		ContainerId: pod.Status.ContainerStatuses[0].ContainerID,
	})
	if err != nil {
		return 0, fmt.Errorf("container get pid failed: %s", err.Error())
	}
	if pfCancel != nil {
		pfCancel()
	}
	return int(res.Pid), nil
}

func forwardPorts(ctx context.Context, pod v1.Pod, port uint16) (context.CancelFunc, uint16, error) {
	clientRawConfig, err := e2econfig.LoadClientRawConfig()
	if err != nil {
		log.Fatal("failed to load raw config", err.Error())
	}
	fw, err := portforward.NewPortForwarder(ctx, e2econfig.NewSimpleRESTClientGetter(clientRawConfig))
	if err != nil {
		log.Fatal("failed to create port forwarder", err.Error())
	}
	_, localPort, pfCancel, err := portforward.ForwardOnePort(fw, pod.Namespace, pod.Name, port, false)
	return pfCancel, localPort, err
}
