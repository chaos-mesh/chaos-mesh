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
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"google.golang.org/grpc/grpclog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	kubectlscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/controllers/common"
	"github.com/chaos-mesh/chaos-mesh/pkg/chaosdaemon/pb"
	"github.com/chaos-mesh/chaos-mesh/pkg/utils"
	e2econfig "github.com/chaos-mesh/chaos-mesh/test/e2e/config"
	"github.com/chaos-mesh/chaos-mesh/test/e2e/util/portforward"
)

var (
	colorFunc = map[string]func(string, ...interface{}){
		"Blue":  color.Blue,
		"Red":   color.Red,
		"Green": color.Green,
		"Cyan":  color.Cyan,
	}
	scheme = runtime.NewScheme()
)

// ClientSet contains two different clients
type ClientSet struct {
	CtrlCli client.Client
	KubeCli *kubernetes.Clientset
}

type ChaosResult struct {
	Name string
	Pods []PodResult
}

type PodResult struct {
	Name  string
	Items []ItemResult
}

const (
	ItemSuccess = iota + 1
	ItemFailure
)

type ItemResult struct {
	Name    string
	Value   string
	Status  int    `json:",omitempty"`
	SucInfo string `json:",omitempty"`
	ErrInfo string `json:",omitempty"`
}

func init() {
	_ = v1alpha1.AddToScheme(scheme)
	_ = clientgoscheme.AddToScheme(scheme)
}

func upperCaseChaos(str string) string {
	parts := regexp.MustCompile("(.*)(chaos)").FindStringSubmatch(str)
	return strings.Title(parts[1]) + strings.Title(parts[2])
}

func print(s string, num int, color string) {
	var tabStr string
	for i := 0; i < num; i++ {
		tabStr += "\t"
	}
	str := fmt.Sprintf("%s%s\n\n", tabStr, regexp.MustCompile("\n").ReplaceAllString(s, "\n"+tabStr))
	if color != "" {
		if cfunc, ok := colorFunc[color]; !ok {
			fmt.Printf("COLOR NOT SUPPORTED")
		} else {
			cfunc(str)
		}
	} else {
		fmt.Printf(str)
	}
}

// PrintResult prints result to users in prettier format
func PrintResult(result []ChaosResult) {
	for _, chaos := range result {
		print("[Chaos]: "+chaos.Name, 0, "Blue")
		for _, pod := range chaos.Pods {
			print("[Pod]: "+pod.Name, 0, "Blue")
			for i, item := range pod.Items {
				print(fmt.Sprintf("%d. [%s]", i+1, item.Name), 1, "Cyan")
				print(item.Value, 1, "")
				if item.Status == ItemSuccess {
					if item.SucInfo != "" {
						print(item.SucInfo, 1, "Green")
					} else {
						print("Execute as expected", 1, "Green")
					}
				} else if item.Status == ItemFailure {
					print(fmt.Sprintf("Failed: %s ", item.ErrInfo), 1, "Red")
				}
			}
		}
	}
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
	kubeClient, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		return nil, fmt.Errorf("error in getting access to K8S: %s", err.Error())
	}
	return &ClientSet{ctrlClient, kubeClient}, nil
}

// Exec executes certain command and returns the result
// runtime-controller only support CRUD， use client-go client
func Exec(ctx context.Context, pod v1.Pod, daemon v1.Pod, cmd string, c *kubernetes.Clientset) (string, error) {
	out, err := exec(ctx, pod, daemon, cmd, c)

	if err != nil {
		// use daemon to enter namespace and execute command if command not found (which stream would failed)
		if strings.Contains(err.Error(), "streaming remotecommand") {
			outNs, errNs := nsEnterExec(ctx, err.Error(), pod, daemon, cmd, c)
			if errNs == nil {
				return outNs, nil
			}
			err = fmt.Errorf("%s\nnsenter also failed with: %s", err.Error(), errNs.Error())
		}
		return "", err
	}

	return out, nil
}

func exec(ctx context.Context, pod v1.Pod, daemon v1.Pod, cmd string, c *kubernetes.Clientset) (string, error) {
	name := pod.GetObjectMeta().GetName()
	namespace := pod.GetObjectMeta().GetNamespace()
	// TODO: if `containerNames` is set and specific container is injected chaos,
	// need to use THE name rather than the first one.
	// till 20/11/10 only podchaos and kernelchaos support `containerNames`, so not set it for now
	containerName := pod.Spec.Containers[0].Name

	req := c.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   []string{"/bin/sh", "-c", cmd},
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
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
		return "", fmt.Errorf("error in streaming remotecommand: %s", err.Error())
	}
	if stderr.String() != "" {
		return "", fmt.Errorf(stderr.String())
	}
	return stdout.String(), nil
}

func nsEnterExec(ctx context.Context, stderr string, pod v1.Pod, daemon v1.Pod, cmd string, c *kubernetes.Clientset) (string, error) {
	cmdSubSlice := strings.Fields(cmd)
	if len(cmdSubSlice) == 0 {
		return "", fmt.Errorf("command should not be empty")
	}
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
	pid, err := GetPidFromPod(ctx, pod, daemon)
	if err != nil {
		return "", err
	}
	switch cmdSubSlice[0] {
	case "ps":
		nsenterPath := "-p/proc/" + strconv.Itoa(pid) + "/ns/pid"
		cmdArguments := strings.Join(cmdSubSlice[1:], " ")
		nsCmd := fmt.Sprintf("mount -t proc proc /proc && ps %s && umount proc", cmdArguments)
		newCmd := fmt.Sprintf("/usr/bin/nsenter %s -- /bin/bash -c '%s'", nsenterPath, nsCmd)
		return exec(ctx, daemon, daemon, newCmd, c)
	case "cat", "ls":
		// we need to enter mount namespace to get file related infomation
		// but enter mnt ns would prevent us to access `cat`/`ls` in daemon
		// so use `nsexec` to achieve using nsenter and cat together
		if len(cmdSubSlice) < 2 {
			return "", fmt.Errorf("%s should have one argument at least", cmdSubSlice[0])
		}
		newCmd := fmt.Sprintf("/usr/local/bin/nsexec %s %s", strconv.Itoa(pid), cmd)
		return exec(ctx, daemon, daemon, newCmd, c)
	default:
		return "", fmt.Errorf("command not supported for nsenter")
	}
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

	if phase == v1alpha1.ExperimentPhaseWaiting {
		waitTime := nextStart.Sub(time.Now())
		fmt.Printf("Waiting for chaos to start, in %s\n", waitTime)
		time.Sleep(waitTime)
	}

	pods, err := utils.SelectPods(ctx, c, c, selector)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to SelectPods with: %s", err.Error())
	}
	if len(pods) == 0 {
		return nil, nil, fmt.Errorf("no pods found for selector: %s", selector)
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
		if err != nil {
			return nil, nil, fmt.Errorf("failed to SelectPods with: %s", err.Error())
		}
		if len(daemons) == 0 {
			return nil, nil, fmt.Errorf("no daemons found for selector: %s", daemonSelector)
		}
		chaosDaemons = append(chaosDaemons, daemons[0])
	}

	return pods, chaosDaemons, nil
}

// GetChaosList returns chaos list limited by input
func GetChaosList(ctx context.Context, chaosType string, chaosName string, ns string, c client.Client) ([]runtime.Object, []string, error) {
	chaosType = upperCaseChaos(strings.ToLower(chaosType))
	allKinds := v1alpha1.AllKinds()
	chaosListInterface := allKinds[chaosType].ChaosList

	if err := c.List(ctx, chaosListInterface, client.InNamespace(ns)); err != nil {
		return nil, nil, fmt.Errorf("failed to get chaosList: %s", err.Error())
	}
	chaosList := chaosListInterface.ListChaos()
	if len(chaosList) == 0 {
		return nil, nil, fmt.Errorf("no chaos is found, please check your input")
	}

	var retList []runtime.Object
	var nameList []string
	for _, ch := range chaosList {
		if chaosName == "" || chaosName == ch.Name {
			chaos, err := getChaos(ctx, chaosType, ch.Name, ns, c)
			if err != nil {
				return nil, nil, err
			}
			retList = append(retList, chaos)
			nameList = append(nameList, ch.Name)
		}
	}
	if len(retList) == 0 {
		return nil, nil, fmt.Errorf("no chaos is found, please check your input")
	}

	return retList, nameList, nil
}

func getChaos(ctx context.Context, chaosType string, chaosName string, ns string, c client.Client) (runtime.Object, error) {
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
	_, localPort, pfCancel, err := portforward.ForwardOnePort(fw, pod.Namespace, pod.Name, port)
	return pfCancel, localPort, err
}
