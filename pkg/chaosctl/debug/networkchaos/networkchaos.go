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

	"github.com/pkg/errors"
	"google.golang.org/grpc/grpclog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	cm "github.com/chaos-mesh/chaos-mesh/pkg/chaosctl/common"
)

// Debug get chaos debug information
func Debug(ctx context.Context, chaos runtime.Object, c *cm.ClientSet, result *cm.ChaosResult) error {
	networkChaos, ok := chaos.(*v1alpha1.NetworkChaos)
	if !ok {
		return fmt.Errorf("chaos is not network")
	}
	chaosStatus := networkChaos.Status.ChaosStatus
	chaosSelector := networkChaos.Spec.Selector

	pods, daemons, err := cm.GetPods(ctx, networkChaos.GetName(), chaosStatus, chaosSelector, c.CtrlCli)
	if err != nil {
		return err
	}

	for i := range pods {
		podName := pods[i].Name
		podResult := cm.PodResult{Name: podName}
		err = debugEachPod(ctx, pods[i], daemons[i], networkChaos, c, &podResult)
		if err != nil {
			fmt.Println(err)
		}
		result.Pods = append(result.Pods, podResult)
		// TODO: V(4) log when err != nil, wait for #1433
	}
	return nil
}

func debugEachPod(ctx context.Context, pod v1.Pod, daemon v1.Pod, chaos *v1alpha1.NetworkChaos, c *cm.ClientSet, result *cm.PodResult) error {
	// To disable printing irrelevant log from grpc/clientconn.go
	// see grpc/grpc-go#3918 for detail. could be resolved in the future
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
	pid, err := cm.GetPidFromPod(ctx, pod, daemon)
	if err != nil {
		return err
	}
	nsenterPath := fmt.Sprintf("-n/proc/%d/ns/net", pid)

	// print out debug info
	cmd := fmt.Sprintf("/usr/bin/nsenter %s -- ipset list", nsenterPath)
	out, err := cm.Exec(ctx, daemon, cmd, c.KubeCli)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("run command '%s' failed", cmd))
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "ipset list", Value: string(out)})

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- tc qdisc list", nsenterPath)
	out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("run command '%s' failed", cmd))
	}
	itemResult := cm.ItemResult{Name: "tc qdisc list", Value: string(out)}

	// A demo for comparison with expected. A bit messy actually, don't know if we still need this
	action := chaos.Spec.Action
	var netemExpect string
	switch action {
	case "delay":
		latency := chaos.Spec.Delay.Latency
		jitter := chaos.Spec.Delay.Jitter
		correlation := chaos.Spec.Delay.Correlation
		netemExpect = fmt.Sprintf("%v %v %v %v%%", action, latency, jitter, correlation)

		netemCurrent := regexp.MustCompile("(?:limit 1000)(.*)").FindStringSubmatch(string(out))
		if len(netemCurrent) == 0 {
			return fmt.Errorf("no NetworkChaos is applied")
		}
		for i, netem := range strings.Fields(netemCurrent[1]) {
			itemCurrent := netem
			itemExpect := strings.Fields(netemExpect)[i]
			if itemCurrent != itemExpect {
				r := regexp.MustCompile("([0-9]*[.])?[0-9]+")
				// digit could be different, so parse string to float
				numCurrent, err := strconv.ParseFloat(r.FindString(itemCurrent), 64)
				if err != nil {
					return errors.Wrap(err, "parse itemCurrent failed")
				}
				numExpect, err := strconv.ParseFloat(r.FindString(itemExpect), 64)
				if err != nil {
					return errors.Wrap(err, "parse itemExpect failed")
				}
				if numCurrent == numExpect {
					continue
				}
				// alphabetic characters
				alpCurrent := regexp.MustCompile("[[:alpha:]]+").FindString(itemCurrent)
				alpExpect := regexp.MustCompile("[[:alpha:]]+").FindString(itemExpect)
				if alpCurrent == alpExpect {
					continue
				}
				itemResult.Status = cm.ItemFailure
				itemResult.ErrInfo = fmt.Sprintf("expect: %s, got: %v", netemExpect, netemCurrent)
			}
		}
		if itemResult.Status != cm.ItemFailure {
			itemResult.Status = cm.ItemSuccess
		}
	}
	result.Items = append(result.Items, itemResult)

	cmd = fmt.Sprintf("/usr/bin/nsenter %s -- iptables --list", nsenterPath)
	out, err = cm.Exec(ctx, daemon, cmd, c.KubeCli)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("run command %s failed", cmd))
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "iptables list", Value: string(out)})

	podNetworkChaos := &v1alpha1.PodNetworkChaos{}
	objectKey := client.ObjectKey{
		Namespace: pod.Namespace,
		Name:      pod.Name,
	}

	if err = c.CtrlCli.Get(ctx, objectKey, podNetworkChaos); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to get network chaos %s/%s", podNetworkChaos.GetNamespace(), podNetworkChaos.GetName()))
	}
	output, err := cm.MarshalChaos(podNetworkChaos.Spec)
	if err != nil {
		return err
	}
	result.Items = append(result.Items, cm.ItemResult{Name: "podnetworkchaos", Value: output})

	return nil
}
