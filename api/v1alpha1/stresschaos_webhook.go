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

package v1alpha1

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/docker/go-units"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Validate validates the scheduler and duration
func (in *StressChaosSpec) Validate(root interface{}, path *field.Path) field.ErrorList {
	if len(in.StressngStressors) == 0 && in.Stressors == nil {
		return field.ErrorList{
			field.Invalid(path, in, "missing stressors"),
		}
	}
	return nil
}

// Validate validates whether the Stressors are all well defined
func (in *Stressors) Validate(root interface{}, path *field.Path) field.ErrorList {
	if in == nil {
		return nil
	}

	if in.MemoryStressor == nil && in.CPUStressor == nil {
		return field.ErrorList{
			field.Invalid(path, in, "missing stressors"),
		}
	}
	return nil
}

func (in Bytes) Validate(root interface{}, path *field.Path) field.ErrorList {
	size := in
	length := len(size)
	if length == 0 {
		return nil
	}

	var err error
	if size[length-1] == '%' {
		var percent int
		percent, err = strconv.Atoi(string(size)[:length-1])
		if err != nil {
			goto handleErr
		}
		if percent > 100 || percent < 0 {
			err = errors.New("illegal proportion")
			goto handleErr
		}
	} else {
		_, err = units.FromHumanSize(string(size))
		if err != nil {
			goto handleErr
		}
	}

	return nil

handleErr:
	return field.ErrorList{
		field.Invalid(path, in, fmt.Sprintf("incorrect bytes format: %s", err.Error())),
	}
}

// Validate validates whether the Stressor is well defined
func (in *Stressor) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if in.Workers <= 0 {
		errs = append(errs, field.Invalid(parent, in, "workers should always be positive"))
	}
	return errs
}
