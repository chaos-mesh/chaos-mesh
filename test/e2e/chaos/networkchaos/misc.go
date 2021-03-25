// Copyright 2020 Chaos Mesh Authors.
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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

const (
	networkConditionBlocked = "blocked"
	networkConditionSlow    = "slow"
	networkConditionGood    = "good"
)

func recvUDPPacket(c http.Client, port uint16) (string, error) {
	klog.Infof("sending request to http://localhost:%d/network/recv", port)
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/network/recv", port))
	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	result := string(out)
	return result, nil
}

func sendUDPPacket(c http.Client, port uint16, targetIP string) error {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	klog.Infof("sending request to http://localhost:%d/network/send with body: %s", port, string(body))

	resp, err := c.Post(fmt.Sprintf("http://localhost:%d/network/send", port), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	result := string(out)
	if result != "send successfully\n" {
		return fmt.Errorf("doesn't send successfully")
	}

	klog.Info("send request successfully")
	return nil
}

func testNetworkDelay(c http.Client, port uint16, targetIP string) (int64, error) {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	klog.Infof("sending request to localhost:%d with body: %s", port, string(body))

	resp, err := c.Post(fmt.Sprintf("http://localhost:%d/network/ping", port), "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	out, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}

	result := string(out)
	parts := strings.Split(result, " ")
	if len(parts) != 2 {
		return 0, fmt.Errorf("the length of parts is not 2 %v", parts)
	}

	if parts[0] != "OK" {
		return 0, fmt.Errorf("the first part of response is not OK")
	}

	return strconv.ParseInt(parts[1], 10, 64)
}

func makeNetworkPartitionChaos(
	namespace, name string, fromLabelSelectors, toLabelSelectors map[string]string,
	fromPodMode, toPodMode v1alpha1.PodMode,
	direction v1alpha1.Direction,
	duration *string,
) *v1alpha1.NetworkChaos {
	var target *v1alpha1.PodSelector
	if toLabelSelectors != nil {
		target = &v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				Namespaces:     []string{namespace},
				LabelSelectors: toLabelSelectors,
			},
			Mode: toPodMode,
		}
	}

	return &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action:    v1alpha1.PartitionAction,
			Direction: direction,
			Target:    target,
			Duration:  duration,
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					Namespaces:     []string{namespace},
					LabelSelectors: fromLabelSelectors,
				},
				Mode: fromPodMode,
			},
		},
	}
}

func makeNetworkDelayChaos(
	namespace, name string, fromLabelSelectors, toLabelSelectors map[string]string,
	fromPodMode, toPodMode v1alpha1.PodMode, direction v1alpha1.Direction, tcparam v1alpha1.TcParameter, duration *string,
) *v1alpha1.NetworkChaos {
	var target *v1alpha1.PodSelector
	if toLabelSelectors != nil {
		target = &v1alpha1.PodSelector{
			Selector: v1alpha1.PodSelectorSpec{
				Namespaces:     []string{namespace},
				LabelSelectors: toLabelSelectors,
			},
			Mode: toPodMode,
		}
	}

	return &v1alpha1.NetworkChaos{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1alpha1.NetworkChaosSpec{
			Action:      v1alpha1.DelayAction,
			TcParameter: tcparam,
			Duration:    duration,
			Target:      target,
			Direction:   direction,
			PodSelector: v1alpha1.PodSelector{
				Selector: v1alpha1.PodSelectorSpec{
					Namespaces:     []string{namespace},
					LabelSelectors: fromLabelSelectors,
				},
				Mode: fromPodMode,
			},
		},
	}
}

func probeNetworkCondition(c http.Client, peers []*corev1.Pod, ports []uint16, bidirection bool) map[string][][]int {
	result := make(map[string][][]int)

	testDelay := func(from int, to int) (int64, error) {
		return testNetworkDelay(c, ports[from], peers[to].Status.PodIP)
	}

	for source := 0; source < len(peers); source++ {
		initialTarget := source + 1
		if bidirection {
			initialTarget = 0
		}
		for target := initialTarget; target < len(peers); target++ {
			if target == source {
				continue
			}

			connectable := true

			var (
				wg           sync.WaitGroup
				link1, link2 bool
			)
			wg.Add(2)
			go func() {
				defer wg.Done()
				// case 1-1: source to target blocked?
				klog.Infof("testing connectivity from %s to %s", peers[source].Name, peers[target].Name)
				link1 = couldConnect(c, ports[source], peers[target].Status.PodIP, ports[target])

			}()

			go func() {
				defer wg.Done()
				// case 1-2: target to source blocked?
				klog.Infof("testing connectivity from %s to %s", peers[target].Name, peers[source].Name)
				link2 = couldConnect(c, ports[target], peers[source].Status.PodIP, ports[source])
			}()
			wg.Wait()

			if !link1 {
				klog.Infof("%s could not connect to %s", peers[source].Name, peers[target].Name)
				result[networkConditionBlocked] = append(result[networkConditionBlocked], []int{source, target})
				connectable = false
			}
			if !link2 {
				klog.Infof("%s could not connect to %s", peers[target].Name, peers[source].Name)
				result[networkConditionBlocked] = append(result[networkConditionBlocked], []int{target, source})
				connectable = false
			}

			if !connectable {
				continue
			}

			// case 2: slow network
			klog.Infof("testing delay from %s to %s", peers[source].Name, peers[target].Name)
			delay, err := testDelay(source, target)
			if err != nil {
				klog.Errorf("error from %d to %d: %v", source, target, err)
				continue
			}

			klog.Infof("delay from %d to %d: %d", source, target, delay)
			if delay > 100*1e6 {
				klog.Infof("detect slow network from %s to %s", peers[source].Name, peers[target].Name)
				result[networkConditionSlow] = append(result[networkConditionSlow], []int{source, target})
				continue
			}

			// case 3: otherwise, good network
			klog.Infof("good network from %d to %d", source, target)
			result[networkConditionGood] = append(result[networkConditionGood], []int{source, target})
		}
	}

	return result
}

func couldConnect(c http.Client, sourcePort uint16, targetPodIP string, targetPort uint16) bool {
	err := sendUDPPacket(c, sourcePort, targetPodIP)
	if err != nil {
		klog.Infof("Error: %v", err)
		return false
	}

	time.Sleep(time.Second)

	data, err := recvUDPPacket(c, targetPort)
	if err != nil {
		klog.Infof("Error: %v, Data: %s", err, data)
		return false
	}

	// FIXME: slow network may also make this happens
	if data != "ping\n" {
		klog.Infof("mismatch data return: %s, it may happens under bad network", data)
	}

	return true
}
