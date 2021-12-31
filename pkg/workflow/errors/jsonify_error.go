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

import (
	"encoding/json"
	"fmt"
	"reflect"

	"golang.org/x/xerrors"
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
		}
		return fmt.Sprintf(
			"failed to jsonify error on type %s, json error, %s; also failed to Unwrap() on it.",
			reflect.TypeOf(origin).Name(),
			err,
		)

	}
	return string(out)
}
