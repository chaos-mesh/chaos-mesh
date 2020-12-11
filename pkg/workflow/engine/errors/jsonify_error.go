package errors

import (
	"encoding/json"
	"fmt"
	"golang.org/x/xerrors"
	"reflect"
)

func toJsonOrFallbackToError(origin error) string {
	out, err := json.Marshal(origin)
	if err != nil {
		if wrapper, ok := err.(xerrors.Wrapper); ok {
			return fmt.Sprintf(
				"failed to jsonify error on type %s, json error: %s; origin error message: %s",
				reflect.TypeOf(origin).Name(),
				err,
				wrapper.Unwrap().Error(),
			)
		} else {
			return fmt.Sprintf(
				"failed to jsonify erro on type %s, json error, %s; also failed to Unwrap() on it.",
				reflect.TypeOf(origin).Name(),
				err,
			)
		}
	} else {
		return string(out)
	}
}
