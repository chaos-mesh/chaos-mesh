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

package statuscheck

import (
	"crypto/x509"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// fakeHTTPExecutor
type fakeHTTPExecutor struct {
	logger logr.Logger

	httpStatusCheck v1alpha1.HTTPStatusCheck
	timeoutSeconds  int
}

func (e *fakeHTTPExecutor) Do() (bool, string, error) {
	return e.handle()
}

func (e *fakeHTTPExecutor) Type() string {
	return "Fake-HTTP"
}

func newFakeExecutor(logger logr.Logger, _ *x509.CertPool, statusCheck v1alpha1.StatusCheck) (Executor, error) {
	var executor Executor
	switch statusCheck.Spec.Type {
	case v1alpha1.TypeHTTP:
		if statusCheck.Spec.EmbedStatusCheck == nil || statusCheck.Spec.HTTPStatusCheck == nil {
			// this should not happen, if the webhook works as expected
			return nil, errors.New("illegal status check, http should not be empty")
		}
		executor = &fakeHTTPExecutor{
			logger:          logger.WithName("fake-http-executor"),
			httpStatusCheck: *statusCheck.Spec.HTTPStatusCheck,
			timeoutSeconds:  statusCheck.Spec.TimeoutSeconds,
		}
	default:
		return nil, errors.New("unsupported type")
	}
	return executor, nil
}

func (e *fakeHTTPExecutor) handle() (bool, string, error) {
	switch e.httpStatusCheck.RequestBody {
	case "failure":
		return false, "failure", nil
	case "timeout":
		time.Sleep(time.Duration(e.timeoutSeconds) * time.Second)
		return false, "timeout", nil
	default:
		return true, "", nil
	}
}
