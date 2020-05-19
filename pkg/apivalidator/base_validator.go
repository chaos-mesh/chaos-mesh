// Copyright 2020 PingCAP, Inc.
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

package apivalidator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
)

// NameValid can be used to check whether the given name is valid.
func NameValid(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	if name == "" {
		return false
	}

	if len(name) > 63 {
		return false
	}

	if !checkName(name) {
		return false
	}

	return true
}

// PATTEN defines allowed characters for name, namespace.
const PATTEN = "^[-.\\w]*$"

// checkName can be used to check whether the given name meets PATTEN.
func checkName(name string) bool {
	patten, err := regexp.Compile(PATTEN)
	if err != nil {
		return false
	}

	return patten.MatchString(name)
}

// CronValid can be used to check whether the given cron valid.
func CronValid(fl validator.FieldLevel) bool {
	cr := fl.Field().String()
	if len(cr) == 0 {
		return true
	}

	if _, err := cron.ParseStandard(cr); err != nil {
		return false
	}

	return true
}
