package errors

type NoSuchNodeError struct {
	Op  string
	Err error

	NodeName     string
	WorkflowName string
}

func (e *NoSuchNodeError) Error() string {
	return toJsonOrFallbackToError(e)
}

func (e *NoSuchNodeError) Unwrap() error {
	return e.Err
}

func NewNoSuchNodeError(op string, nodeName string, workflowName string) *NoSuchNodeError {
	return &NoSuchNodeError{Op: op, NodeName: nodeName, WorkflowName: workflowName, Err: ErrNoSuchNode}
}
