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

package jvm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	BaseURL    = "http://%s:%d/sandbox/default/module/http/"
	ActiveURL  = BaseURL + "sandbox-module-mgr/active?ids=chaosblade"
	InjectURL  = BaseURL + "chaosblade/create"
	RecoverURL = BaseURL + "chaosblade/destroy"
)

// ActiveSandbox activates sandboxes
func ActiveSandbox(host string, port int) error {
	url := fmt.Sprintf(ActiveURL, host, port)

	_, err := http.Get(url)
	return err
}

// InjectChaos injects jvm chaos to a java process
func InjectChaos(host string, port int, body []byte) error {
	url := fmt.Sprintf(InjectURL, host, port)
	return httpPost(url, body)
}

// RecoverChaos recovers jvm chaos from a java process
func RecoverChaos(host string, port int, body []byte) error {
	url := fmt.Sprintf(RecoverURL, host, port)
	return httpPost(url, body)
}

func httpPost(url string, body []byte) error {
	client := &http.Client{}
	reqBody := bytes.NewBuffer([]byte(body))
	request, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return err
	}

	request.Header.Set("Content-type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("jvm sandbox error response:%s", data)
	}
	return nil
}
