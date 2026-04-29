// Copyright 2024 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package steps contains re-usable probe helpers used by the networkchaos BDD
// step definitions. The logic here mirrors the unexported helpers in
// e2e-test/e2e/chaos/networkchaos/misc.go so the BDD layer stays self-contained.
package steps

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	networkConditionBlocked = "blocked"
	networkConditionSlow    = "slow"
	networkConditionGood    = "good"
)

func recvUDPPacket(c http.Client, port uint16) (string, error) {
	resp, err := c.Get(fmt.Sprintf("http://localhost:%d/network/recv", port))
	if err != nil {
		return "", err
	}
	out, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func sendUDPPacket(c http.Client, port uint16, targetIP string) error {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	resp, err := c.Post(fmt.Sprintf("http://localhost:%d/network/send", port), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	out, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if string(out) != "send successfully\n" {
		return fmt.Errorf("send failed: %s", string(out))
	}
	return nil
}

func testNetworkDelay(c http.Client, port uint16, targetIP string) (int64, error) {
	body := []byte(fmt.Sprintf("{\"targetIP\":\"%s\"}", targetIP))
	resp, err := c.Post(fmt.Sprintf("http://localhost:%d/network/ping", port), "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	out, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return 0, err
	}
	parts := strings.Split(string(out), " ")
	if len(parts) != 2 || parts[0] != "OK" {
		return 0, fmt.Errorf("unexpected response: %s", string(out))
	}
	return strconv.ParseInt(parts[1], 10, 64)
}

func couldConnect(c http.Client, sourcePort uint16, targetPodIP string, targetPort uint16) bool {
	if err := sendUDPPacket(c, sourcePort, targetPodIP); err != nil {
		return false
	}
	time.Sleep(time.Second)
	data, err := recvUDPPacket(c, targetPort)
	if err != nil {
		return false
	}
	if data != "ping\n" {
		klog.Infof("mismatch data return: %s", data)
	}
	return true
}

func probeNetworkCondition(c http.Client, peers []*corev1.Pod, ports []uint16, bidirection bool) map[string][][]int {
	result := make(map[string][][]int)

	testDelay := func(from, to int) (int64, error) {
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
				link1 = couldConnect(c, ports[source], peers[target].Status.PodIP, ports[target])
			}()
			go func() {
				defer wg.Done()
				link2 = couldConnect(c, ports[target], peers[source].Status.PodIP, ports[source])
			}()
			wg.Wait()

			if !link1 {
				result[networkConditionBlocked] = append(result[networkConditionBlocked], []int{source, target})
				connectable = false
			}
			if !link2 {
				result[networkConditionBlocked] = append(result[networkConditionBlocked], []int{target, source})
				connectable = false
			}
			if !connectable {
				continue
			}

			delay, err := testDelay(source, target)
			if err != nil {
				continue
			}
			if delay > 100*1e6 {
				result[networkConditionSlow] = append(result[networkConditionSlow], []int{source, target})
				continue
			}
			result[networkConditionGood] = append(result[networkConditionGood], []int{source, target})
		}
	}
	return result
}
