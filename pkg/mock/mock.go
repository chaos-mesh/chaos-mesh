package mock

import (
	"fmt"
	"path"
	"reflect"
	"unsafe"

	"github.com/pingcap/failpoint"
)

type finalizer func() error

func On(fpname string) interface{} {
	var ret interface{}
	if val, ok := failpoint.Eval(_curpkg_(fpname)); ok {
		ret = *(*interface{})(unsafe.Pointer(uintptr(val.(int))))
	}
	return ret
}

func With(fpname string, value interface{}) finalizer {
	type em struct{}
	k := path.Join(reflect.TypeOf(em{}).PkgPath(), fpname)
	v := fmt.Sprintf("return(%d)", int(uintptr(unsafe.Pointer(&value))))
	if err := failpoint.Enable(k, v); err != nil {
		panic(err)
	}
	return func() error { return failpoint.Disable(k) }
}

func Reset(fpname string) error {
	type em struct{}
	k := path.Join(reflect.TypeOf(em{}).PkgPath(), fpname)
	return failpoint.Disable(k)
}
