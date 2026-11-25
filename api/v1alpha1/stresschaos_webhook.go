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

package v1alpha1

import (
	"flag"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/dustin/go-humanize"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/chaos-mesh/chaos-mesh/api/genericwebhook"
)

// Validate checks the stressors that this experiment will run.
func (in *StressChaosSpec) Validate(_ any, path *field.Path) field.ErrorList {
	if len(in.StressngStressors) == 0 && in.Stressors == nil {
		return field.ErrorList{
			field.Invalid(path, in, "missing stressors"),
		}
	}
	if in.Stressors != nil && in.Stressors.MemoryStressor != nil {
		if err := in.Stressors.MemoryStressor.validateSize(path.Child("stressors", "memory"), in.StressngStressors == ""); err != nil {
			return field.ErrorList{err}
		}
	}
	return nil
}

// Validate validates whether the Stressors are all well defined
func (in *Stressors) Validate(_ any, path *field.Path) field.ErrorList {
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

type Bytes string

func (in *Bytes) Validate(_ any, path *field.Path) field.ErrorList {
	if _, err := positiveMemorySize(string(*in)); err != nil {
		return field.ErrorList{field.Invalid(path, *in, err.Error())}
	}
	return nil
}

func positiveMemorySize(size string) (bool, error) {
	if size == "" {
		return false, nil
	}
	if strings.IndexFunc(size, unicode.IsSpace) >= 0 {
		return false, fmt.Errorf("write memory size without spaces, for example 100MB")
	}
	if number, ok := strings.CutSuffix(size, "%"); ok {
		percent, err := strconv.Atoi(number)
		if err != nil || percent < 0 || percent > 100 {
			return false, fmt.Errorf("memory size percentage must be an integer between 0 and 100")
		}
		return percent > 0, nil
	}
	n, err := humanize.ParseBytes(size)
	return n > 0, err
}

// The daemon splits normalized options into words and memStress takes the last size flag.
func memorySizeOverride(options []string) (*string, error) {
	flags := flag.NewFlagSet("memStress", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var size *string
	flags.Func("size", "", func(value string) error {
		size = &value
		return nil
	})
	flags.String("time", "", "")
	flags.Int("workers", 1, "")
	flags.Bool("client", false, "")
	if err := flags.Parse(strings.Fields(strings.Join(options, " "))); err != nil {
		return nil, err
	}
	if flags.NArg() != 0 {
		return nil, fmt.Errorf("unexpected memStress argument %q, write memory sizes without spaces", flags.Arg(0))
	}
	return size, nil
}

func (in *MemoryStressor) validateSize(path *field.Path, active bool) *field.Error {
	sizePath := path.Child("size")
	size := in.Size
	positive, err := positiveMemorySize(size)
	if err != nil {
		return field.Invalid(sizePath, size, err.Error())
	}
	if !active || in.Workers == 0 {
		return nil
	}
	if len(in.Options) > 0 {
		override, err := memorySizeOverride(in.Options)
		if err != nil {
			return field.Invalid(path.Child("options"), in.Options, err.Error())
		}
		if override != nil {
			sizePath, size = path.Child("options"), *override
			positive, err = positiveMemorySize(size)
			if err != nil {
				return field.Invalid(sizePath, size, err.Error())
			}
		}
	}
	if size == "" {
		return field.Required(sizePath, "specify a memory size in size or with --size in options")
	}
	if !positive {
		return field.Invalid(sizePath, size, "memory size must be greater than zero")
	}
	return nil
}

// Validate validates whether the Stressor is well defined
func (in *Stressor) Validate(parent *field.Path) field.ErrorList {
	errs := field.ErrorList{}
	if in.Workers <= 0 {
		errs = append(errs, field.Invalid(parent, in, "workers should always be positive"))
	}
	return errs
}

func init() {
	genericwebhook.Register("Bytes", reflect.PointerTo(reflect.TypeFor[Bytes]()))
}
