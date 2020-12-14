package errors

type NoTemplatesError struct {
	Op  string
	Err error

	WorkflowName string
}

func (e *NoTemplatesError) Error() string {
	return toJsonOrFallbackToError(e)
}

func (e *NoTemplatesError) Unwrap() error {
	return e.Err
}

func NewNoTemplatesError(op, workflowName string) *NoTemplatesError {
	return &NoTemplatesError{
		Op:           op,
		Err:          ErrNoTemplates,
		WorkflowName: workflowName,
	}
}
