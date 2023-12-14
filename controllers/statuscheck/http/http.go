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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

type httpExecutor struct {
	logger   logr.Logger
	certPool *x509.CertPool

	timeoutSeconds  int
	httpStatusCheck v1alpha1.HTTPStatusCheck
}

func NewExecutor(logger logr.Logger, certPool *x509.CertPool, timeoutSeconds int, httpStatusCheck v1alpha1.HTTPStatusCheck) *httpExecutor {
	return &httpExecutor{logger: logger, certPool: certPool, timeoutSeconds: timeoutSeconds, httpStatusCheck: httpStatusCheck}
}

type response struct {
	statusCode int
	body       string
}

func (e *httpExecutor) Type() string {
	return "HTTP"
}

func (e *httpExecutor) Do() (bool, string, error) {
	client := &http.Client{
		Timeout: time.Duration(e.timeoutSeconds) * time.Second,
	}

	if e.certPool != nil {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: e.certPool,
			},
		}
	}

	httpStatusCheck := e.httpStatusCheck
	return e.DoHTTPRequest(client,
		httpStatusCheck.RequestUrl,
		string(httpStatusCheck.RequestMethod),
		httpStatusCheck.RequestHeaders,
		[]byte(httpStatusCheck.RequestBody),
		httpStatusCheck.Criteria)
}

func (e *httpExecutor) DoHTTPRequest(client *http.Client, url, method string,
	headers http.Header, body []byte, criteria v1alpha1.HTTPCriteria) (bool, string, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return false, errors.Wrap(err, "new http request").Error(), nil
	}
	req.Header = headers
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "do http request").Error(), nil
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", errors.Wrap(err, "read response body")
	}

	return validate(e.logger.WithValues("url", url),
		criteria, response{statusCode: resp.StatusCode, body: string(responseBody)})
}

func validate(logger logr.Logger, criteria v1alpha1.HTTPCriteria, resp response) (bool, string, error) {
	ok := validateStatusCode(criteria.StatusCode, resp)
	if !ok {
		logger.Info("validate status code failed",
			"criteria", criteria.StatusCode,
			"statusCode", resp.statusCode)
		return false, fmt.Sprintf("unexpected status code: %d", resp.statusCode), nil
	}
	return ok, "", nil
}

// validateStatusCode validate whether the result is as expected.
// A criteria(statusCode) string could be a single code (e.g. 200), or
// an inclusive range (e.g. 200-400, both `200` and `400` are included).
// The format of the criteria field will be validated in webhook.
func validateStatusCode(criteria string, resp response) bool {
	if code, err := strconv.Atoi(criteria); err == nil {
		return code == resp.statusCode
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
	return resp.statusCode >= startStatusCode &&
		resp.statusCode <= endStatusCode
}
