// Copyright Chaos Mesh Authors.
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

package http

import (
	"bytes"
	"errors"
	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type httpExecutor struct {
}

func NewExecutor() *httpExecutor {
	return &httpExecutor{}
}

type result struct {
	statusCode int
}

func (e *httpExecutor) Do(spec v1alpha1.StatusCheckSpec) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(spec.TimeoutSeconds) * time.Second,
	}
	if spec.EmbedStatusCheck == nil || spec.EmbedStatusCheck.HTTPStatusCheck == nil {
		return "", errors.New("")
	}

	httpStatusCheck := spec.HTTPStatusCheck
	result, err := DoHTTPRequest(client,
		httpStatusCheck.RequestUrl,
		string(httpStatusCheck.RequestMethod),
		httpStatusCheck.RequestHeaders,
		[]byte(httpStatusCheck.RequestBody))
	if err != nil {
		return "", err
	}
	if !validate(spec.HTTPStatusCheck.Criteria, *result) {

	}
	return "", nil
}

func DoHTTPRequest(client *http.Client, url, method string, headers http.Header, body []byte) (*result, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header = headers
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return &result{statusCode: response.StatusCode}, nil
}

func validate(criteria v1alpha1.HTTPCriteria, result result) bool {
	return validateStatusCode(criteria.StatusCode, result)
}

// validateStatusCode validate whether the result is as expected.
// A criteria(statusCode) string could be a single code (e.g. 200), or
// an inclusive range (e.g. 200-400, both `200` and `400` are included).
// The format of the criteria field will be validated in webhook.
func validateStatusCode(criteria string, result result) bool {
	if code, err := strconv.Atoi(criteria); err == nil {
		return code == result.statusCode
	}
	index := strings.Index(criteria, "-")
	if index == -1 {
		return false
	}
	start := criteria[:index]
	end := criteria[index+1:]
	startStatusCode, err := strconv.Atoi(start)
	if err != nil {
		return false
	}
	endStatusCode, err := strconv.Atoi(end)
	if err != nil {
		return false
	}
	return result.statusCode >= startStatusCode &&
		result.statusCode <= endStatusCode
}
