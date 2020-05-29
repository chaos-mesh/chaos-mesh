// Copyright 2019 PingCAP, Inc.
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
	"os"
	"strings"
)

var (
	// AllowedNamespaces the namespace list allow the execution of a chaos task
	AllowedNamespaces []string
	// IgnoredNamespaces the namespace list ignore the chaos task
	IgnoredNamespaces []string
)

func init() {

	ignoredNamespacesText, ok := os.LookupEnv("IGNORED_NAMESPACES")
	if ok {
		IgnoredNamespaces = strings.Split(ignoredNamespacesText, ",")
	}

	allowedNamespacesText, ok := os.LookupEnv("ALLOWED_NAMESPACES")
	if ok {
		AllowedNamespaces = strings.Split(allowedNamespacesText, ",")
	}
}
