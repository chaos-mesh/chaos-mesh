// Copyright Chaos Mesh Authors.
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

package v1alpha1

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/chaos-mesh/chaos-mesh/api/genericwebhook"
)

func (in *StatusCheckSpec) Default(root interface{}, field *reflect.StructField) {
	if in.Mode == "" {
		in.Mode = StatusCheckSynchronous
	}
}

func (in *StatusCheckSpec) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if in.Type == TypeHTTP {
		if in.EmbedStatusCheck == nil || in.EmbedStatusCheck.HTTPStatusCheck == nil {
			allErrs = append(allErrs, field.Invalid(path.Child("http"), nil, "the detail of http status check is required"))
		}
	} else {
		allErrs = append(allErrs, field.Invalid(path.Child("type"), in.Type, fmt.Sprintf("unrecognized type: %s", in.Type)))
	}

	return allErrs
}

func (in *HTTPStatusCheck) Validate(root interface{}, path *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}
	if in.RequestUrl == "" {
		allErrs = append(allErrs, field.Invalid(path.Child("url"), in.RequestUrl, "request url is required"))
		return allErrs
	}

	if _, err := url.ParseRequestURI(in.RequestUrl); err != nil {
		allErrs = append(allErrs, field.Invalid(path.Child("url"), in.RequestUrl, "invalid http request url"))
	}
	return allErrs
}

type StatusCode string

func (in *StatusCode) Validate(root interface{}, path *field.Path) field.ErrorList {
	packError := func(err error) field.ErrorList {
		return field.ErrorList{
			field.Invalid(path, in, fmt.Sprintf("incorrect status code format: %s", err.Error())),
		}
	}

	codeStr := string(*in)
	if codeStr == "" {
		return field.ErrorList{
			field.Invalid(path, in, "status code is required"),
		}
	}

	if code, err := strconv.Atoi(codeStr); err == nil {
		if !validateHTTPStatusCode(code) {
			return packError(errors.New("invalid status code"))
		}
	} else {
		index := strings.Index(codeStr, "-")
		if index == -1 {
			return packError(errors.New("not a single number or a range"))
		}

		validateRange := func(codeStr string) error {
			code, err := strconv.Atoi(codeStr)
			if err != nil {
				return err
			}
			if !validateHTTPStatusCode(code) {
				return errors.Errorf("invalid status code range, code: %d", code)
			}
			return nil
		}
		start := codeStr[:index]
		end := codeStr[index+1:]
		if err := validateRange(start); err != nil {
			return packError(err)
		}
		if err := validateRange(end); err != nil {
			return packError(err)
		}
	}
	return nil
}

func validateHTTPStatusCode(code int) bool {
	return code > 0 && code < 1000
}

func init() {
	genericwebhook.Register("StatusCode", reflect.PtrTo(reflect.TypeOf(StatusCode(""))))
}
