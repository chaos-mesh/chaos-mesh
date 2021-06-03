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
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/grpc/grpclog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	kubectlscheme "k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/chaos-mesh/chaos-mesh/pkg/bpm"
)

// Exec executes certain command and returns the result
// Only commands in chaos-mesh components should use this way
// for target pod, use ExecBypass
func Exec(ctx context.Context, pod v1.Pod, cmd string, c *kubernetes.Clientset) (string, error) {
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
		return "", errors.Wrapf(err, "error in creating NewSPDYExecutor for pod %s/%s", pod.GetNamespace(), pod.GetName())
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		if stderr.String() != "" {
			return "", fmt.Errorf("error: %s\npod: %s\ncommand: %s", strings.TrimSuffix(stderr.String(), "\n"), pod.Name, cmd)
		}
		return "", errors.Wrapf(err, "error in streaming remotecommand: pod: %s/%s, command: %s", pod.GetNamespace(), pod.Name, cmd)
	}
	if stderr.String() != "" {
		return "", fmt.Errorf("error of command %s: %s", cmd, stderr.String())
	}
	return stdout.String(), nil
}

// ExecBypass use chaos-daemon to enter namespace and execute command in target pod
func ExecBypass(ctx context.Context, pod v1.Pod, daemon v1.Pod, cmd string, c *kubernetes.Clientset) (string, error) {
	// To disable printing irrelevant log from grpc/clientconn.go
	// See grpc/grpc-go#3918 for detail. Could be resolved in the future
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
	pid, err := GetPidFromPod(ctx, pod, daemon)
	if err != nil {
		return "", err
	}
	// enter all possible namespaces needed, since there's no bad effect to do so
	cmdBuilder := bpm.DefaultProcessBuilder(cmd).SetNS(pid, bpm.MountNS).SetNS(pid, bpm.PidNS).SetContext(ctx)
	return Exec(ctx, daemon, cmdBuilder.Build().Cmd.String(), c)
}
