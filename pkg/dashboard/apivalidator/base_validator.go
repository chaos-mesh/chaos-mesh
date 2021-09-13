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

package apivalidator

import (
	"regexp"
	"time"

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

var namePattern = regexp.MustCompile(`^[-.\w]*$`)

// checkName can be used to check resource names.
func checkName(name string) bool {
	return namePattern.MatchString(name)
}

// CronValid can be used to check whether the given cron valid.
func CronValid(fl validator.FieldLevel) bool {
	cr := fl.Field().String()
	if cr == "" {
		return true
	}

	if _, err := cron.ParseStandard(cr); err != nil {
		return false
	}

	return true
}

// DurationValid can be used to check whether the given duration valid.
func DurationValid(fl validator.FieldLevel) bool {
	dur := fl.Field().String()
	if dur == "" {
		return true
	}

	_, err := time.ParseDuration(dur)
	return err == nil
}
