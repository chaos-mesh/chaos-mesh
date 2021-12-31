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

package errors

type NoMoreTemplateInSerialTemplateError struct {
	Op  string
	Err error

	WorkflowName string
	TemplateName string
	NodeName     string
}

func (e *NoMoreTemplateInSerialTemplateError) Error() string {
	return toJsonOrFallbackToError(e)
}

func (e *NoMoreTemplateInSerialTemplateError) Unwrap() error {
	return e.Err
}

func NewNoMoreTemplateInSerialTemplateError(op, workflowName, templateName, nodeName string) *NoMoreTemplateInSerialTemplateError {
	return &NoMoreTemplateInSerialTemplateError{
		Op:           op,
		Err:          ErrNoMoreTemplateInSerialTemplate,
		WorkflowName: workflowName,
		TemplateName: templateName,
		NodeName:     nodeName,
	}
}
