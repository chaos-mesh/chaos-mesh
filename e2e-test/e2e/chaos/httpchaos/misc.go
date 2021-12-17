// Copyright 2021 Chaos Mesh Authors.
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
//

package httpchaos

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const SECRET = "Secret"

type HTTPE2EClient struct {
	IP string
	C  *http.Client
}

func getPodHttp(c HTTPE2EClient, port uint16, secret, body string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/http", c.IP, port), strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set(SECRET, secret)
	client := &http.Client{
		Transport: &http.Transport{},
	}
	resp, err := client.Do(request)
	return resp, err
}

func getPodHttpNoSecret(c HTTPE2EClient, port uint16) (*http.Response, error) {
	return getPodHttp(c, port, "", "")
}

func getPodHttpDefaultSecret(c HTTPE2EClient, port uint16, body string) (*http.Response, error) {
	return getPodHttp(c, port, "foo", body)
}

func getPodHttpNoBody(c HTTPE2EClient, port uint16) (*http.Response, error) {
	return getPodHttpDefaultSecret(c, port, "")
}

// get pod http delay
func getPodHttpDelay(c HTTPE2EClient, port uint16) (*http.Response, time.Duration, error) {
	start := time.Now()
	resp, err := getPodHttpNoBody(c, port)
	if err != nil {
		return nil, 0, err
	}

	return resp, time.Now().Sub(start), nil
}
