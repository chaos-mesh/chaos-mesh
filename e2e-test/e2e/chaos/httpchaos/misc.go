// Copyright 2021 Chaos Mesh Authors.
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

package httpchaos

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func getPodHttp(c http.Client, port uint16, secret, body string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/http", port), strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Secret", time.Now().Format(time.RFC3339))
	client := &http.Client{
		Transport: &http.Transport{},
	}
	resp, err := client.Do(request)
	return resp, err
}

func getPodHttpNoSecret(c http.Client, port uint16) (*http.Response, error) {
	return getPodHttp(c, port, "", "")
}

func getPodHttpDefaultSecret(c http.Client, port uint16, body string) (*http.Response, error) {
	return getPodHttp(c, port, "foo", body)
}

func getPodHttpNoBody(c http.Client, port uint16) (*http.Response, error) {
	return getPodHttpDefaultSecret(c, port, "")
}

// get pod http delay
func getPodHttpDelay(c http.Client, port uint16) (*http.Response, time.Duration, error) {
	start := time.Now()
	resp, err := getPodHttpNoBody(c, port)
	if err != nil {
		return nil, 0, err
	}

	return resp, time.Now().Sub(start), nil
}
