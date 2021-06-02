// Copyright 2021 Chaos Mesh Authors.
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

package controller

import (
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

func ParseNamespacedName(namespacedName string) types.NamespacedName {
	parts := strings.Split(namespacedName, "/")
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}
}

func ParseNamespacedNameContainer(namespacedName string) (types.NamespacedName, string) {
	parts := strings.Split(namespacedName, "/")
	return types.NamespacedName{
		Namespace: parts[0],
		Name:      parts[1],
	}, strings.Join(parts[2:], "")
}
