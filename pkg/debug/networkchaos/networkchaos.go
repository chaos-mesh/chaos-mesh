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

package networkchaos

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/grpc/grpclog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/debug/common"
)

func Debug(ctx context.Context, chaos runtime.Object, c *cm.ClientSet) error {
	networkChaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		return fmt.Errorf("chaos is not network")
	}
	chaosStatus := networkChaos.Status.ChaosStatus
	chaosSelector := networkChaos.Spec.GetSelector()

	pods, daemons, err := cm.GetPods(ctx, chaosStatus, chaosSelector, c.CtrlClient)
	if err != nil {
		return err
	}

	for i := range pods {
		podName := pods[i].GetObjectMeta().GetName()
		cm.Print("[Pod]: "+podName, 0, cm.ColorBlue)
		err := debugEachPod(ctx, pods[i], daemons[i], networkChaos, c)
		if err != nil {
			return fmt.Errorf("for %s: %s", podName, err.Error())
		}
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.NetworkChaos, c *cm.ClientSet) error {
	podName := pod.GetObjectMeta().GetName()
	podNamespace := pod.GetObjectMeta().GetNamespace()
	daemonName := daemon.GetObjectMeta().GetName()
	daemonNamespace := daemon.GetObjectMeta().GetNamespace()

	// To disable printing irrelevant log from grpc/clientconn.go
	// see grpc/grpc-go#3918 for detail. could be resolved in the future
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
	pid, err := cm.GetPidFromPod(ctx, pod, daemon)
	if err != nil {
		return err
	}
	nsenterPath := "-n/proc/" + strconv.Itoa(pid) + "/ns/net"

	// print out debug info
	cmd := fmt.Sprintf("/usr/bin/nsenter %s -- ipset list", nsenterPath)
	out, err := cm.Exec(daemonName, daemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	cm.Print("1. [ipset list]", 1, cm.ColorCyan)
	cm.Print(string(out), 1, "")

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- tc qdisc list", nsenterPath)
	out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("run command '%s' failed with: %s", cmd, err.Error())
	}
	cm.Print("2. [tc qdisc list]", 1, cm.ColorCyan)
	cm.Print(string(out), 1, "")

	action := chaos.Spec.Action
	var netemExpect string
	switch action {
	case "delay":
		latency := chaos.Spec.Delay.Latency
		jitter := chaos.Spec.Delay.Jitter
		correlation := chaos.Spec.Delay.Correlation
		netemExpect = fmt.Sprintf("%v %v %v %v%%", action, latency, jitter, correlation)
	default:
		return fmt.Errorf("chaos not supported")
	}

	netemCurrent := regexp.MustCompile("(?:limit 1000)(.*)").FindStringSubmatch(string(out))
	if len(netemCurrent) == 0 {
		return fmt.Errorf("No NetworkChaos is applied")
	}
	for i, netem := range strings.Fields(netemCurrent[1]) {
		itemCurrent := netem
		itemExpect := strings.Fields(netemExpect)[i]
		if itemCurrent != itemExpect {
			r := regexp.MustCompile("([0-9]*[.])?[0-9]+")
			numCurrent, err := strconv.ParseFloat(r.FindString(itemCurrent), 64)
			if err != nil {
				return fmt.Errorf("parse itemCurrent failed: %s", err.Error())
			}
			numExpect, err := strconv.ParseFloat(r.FindString(itemExpect), 64)
			if err != nil {
				return fmt.Errorf("parse itemExpect failed: %s", err.Error())
			}
			if numCurrent == numExpect {
				continue
			}
			alpCurrent := regexp.MustCompile("[[:alpha:]]+").FindString(itemCurrent)
			alpExpect := regexp.MustCompile("[[:alpha:]]+").FindString(itemExpect)
			if alpCurrent == alpExpect {
				continue
			}
			errInfo := fmt.Sprintf("NetworkChaos didn't execute as expected, expect: %s, got: %s", netemExpect, netemCurrent)
			cm.Print(errInfo, 1, cm.ColorRed)
			return nil
		}
	}
	cm.Print("NetworkChaos execute as expected", 1, cm.ColorGreen)

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- iptables --list", nsenterPath)
	out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}
	cm.Print("3. [iptables list]", 1, cm.ColorCyan)
	cm.Print(string(out), 1, "")

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- iptables --list", nsenterPath)
	out, err = cm.Exec(daemonName, daemonNamespace, cmd, c.K8sClient)
	if err != nil {
		return fmt.Errorf("cmd.Run() failed with: %s", err.Error())
	}

	podNetworkChaos := &v1alpha1.PodNetworkChaos{}
	objectKey := client.ObjectKey{
		Namespace: podNamespace,
		Name:      podName,
	}

	if err = c.CtrlClient.Get(ctx, objectKey, podNetworkChaos); err != nil {
		return fmt.Errorf("failed to get chaos: %s", err.Error())
	}
	cm.Print("4. [podnetworkchaos]", 1, cm.ColorCyan)
	mar, err := cm.MarshalChaos(podNetworkChaos.Spec)
	if err != nil {
		return err
	}
	cm.Print(mar, 1, "")

	return nil
}
