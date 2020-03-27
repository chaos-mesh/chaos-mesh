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

package utils

import (
	"strings"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func IgnoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func IsCaredNetError(err error) bool {
	if err == nil {
		return false
	}

	errString := strings.ToLower(err.Error())

	if strings.Contains(errString, "i/o timeout") {
		return true
	}

	if strings.Contains(errString, "connection refused") {
		return true
	}

	return false
}
