package errors

type NoNeedScheduleError struct {
	Op  string
	Err error

	WorkflowName string
}

func (e *NoNeedScheduleError) Error() string {
	return toJsonOrFallbackToError(e)
}

func (e *NoNeedScheduleError) Unwrap() error {
	return e.Err
}

func NewNoNeedScheduleError(op string, workflowName string) *NoNeedScheduleError {
	return &NoNeedScheduleError{
		Op:           op,
		Err:          ErrNoNeedSchedule,
		WorkflowName: workflowName,
	}
}
