package v1alpha1

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/validation/field"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ScheduleItem struct {
	EmbedChaos `json:",inline"`
	// +optional
	Workflow *WorkflowSpec `json:"workflow,omitempty"`
}

func (in EmbedChaos) Validate(chaosType string) field.ErrorList {
	allErrs := field.ErrorList{}
	spec := reflect.ValueOf(in).FieldByName(chaosType)

	if !spec.IsValid() || spec.IsNil() {
		allErrs = append(allErrs, field.Invalid(field.NewPath(chaosType),
			in,
			fmt.Sprintf("parse schedule field error: missing chaos spec")))
		return allErrs
	}
	addr, success := spec.Interface().(CommonSpec)
	if success == false {
		logf.Log.Info(fmt.Sprintf("%s does not seem to have a validator", chaosType))
		return allErrs
	}
	allErrs = append(allErrs, addr.Validate()...)
	return allErrs
}
