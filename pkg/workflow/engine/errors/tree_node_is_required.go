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

package errors

type TreeNodeIsRequiredError struct {
	Op  string
	Err error

	WorkflowName string
}

func (e *TreeNodeIsRequiredError) Error() string {
	return toJsonOrFallbackToError(e)
}

func (e *TreeNodeIsRequiredError) Unwrap() error {
	return e.Err
}

func NewTreeNodeIsRequiredError(op string, workflowName string) *TreeNodeIsRequiredError {
	return &TreeNodeIsRequiredError{
		Op:           op,
		Err:          ErrTreeNodeIsRequired,
		WorkflowName: workflowName,
	}
}
