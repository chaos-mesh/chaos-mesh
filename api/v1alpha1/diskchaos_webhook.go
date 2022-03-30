package v1alpha1

import (
	"github.com/chaos-mesh/chaos-mesh/api/genericwebhook"
	"reflect"
)

type ProcessNum uint8

func (p *ProcessNum) Default(root interface{}, field *reflect.StructField) {
	if *p == 0 {
		*p = 1
	}
}

func init() {
	genericwebhook.Register("ProcessNum", reflect.PtrTo(reflect.TypeOf(ClockIds{})))
}
