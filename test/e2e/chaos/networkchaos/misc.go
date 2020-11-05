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
	"k8s.io/klog"
	"net/http"
	"strconv"
	"strings"
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
