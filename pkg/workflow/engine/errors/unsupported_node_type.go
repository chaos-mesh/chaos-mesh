package errors

import "github.com/chaos-mesh/chaos-mesh/pkg/workflow/engine/model/template"

func NewUnsupportedNodeTypeError(op string, nodeName string, templateType template.TemplateType, workflowName string) error {
	panic("unimplemented")
}
