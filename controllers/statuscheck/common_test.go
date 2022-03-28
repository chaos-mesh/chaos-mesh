package statuscheck

import (
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
)

// fakeHTTPExecutor
type fakeHTTPExecutor struct {
	logger logr.Logger
}

func (e *fakeHTTPExecutor) Do(spec v1alpha1.StatusCheckSpec) (bool, string, error) {
	if spec.EmbedStatusCheck == nil || spec.EmbedStatusCheck.HTTPStatusCheck == nil {
		// this should not happen, if the webhook works as expected
		return false, "illegal status check, http should not be empty", nil
	}
	return e.handle(spec)
}

func (e *fakeHTTPExecutor) Type() string {
	return "Fake-HTTP"
}

func newFakeExecutor(logger logr.Logger, statusCheck v1alpha1.StatusCheck) (Executor, error) {
	var executor Executor
	switch statusCheck.Spec.Type {
	case v1alpha1.TypeHTTP:
		executor = &fakeHTTPExecutor{logger: logger.WithName("fake-http-executor")}
	default:
		return nil, errors.New("unsupported type")
	}
	return executor, nil
}

func (e *fakeHTTPExecutor) handle(spec v1alpha1.StatusCheckSpec) (bool, string, error) {
	switch spec.HTTPStatusCheck.RequestBody {
	case "failure":
		return false, "failure", nil
	case "timeout":
		time.Sleep(time.Duration(spec.TimeoutSeconds) * time.Second)
		return false, "timeout", nil
	default:
		return true, "", nil
	}
}
