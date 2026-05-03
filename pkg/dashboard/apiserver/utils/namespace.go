// Copyright 2025 Chaos Mesh Authors.
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

package utils

import "strings"

// ParseNamespaceQuery parses a comma-separated namespace query parameter.
// It splits by comma, trims whitespace, and filters out empty strings.
// Returns an empty slice if the input is empty or only contains empty strings.
func ParseNamespaceQuery(namespaceParam string) []string {
	if namespaceParam == "" {
		return []string{}
	}

	namespaces := strings.Split(namespaceParam, ",")
	result := make([]string, 0, len(namespaces))

	for _, ns := range namespaces {
		ns = strings.TrimSpace(ns)
		if ns != "" {
			result = append(result, ns)
		}
	}

	return result
}
