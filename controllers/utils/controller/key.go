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
	"errors"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

func ParseNamespacedName(namespacedName string) (types.NamespacedName, error) {
	parts := strings.Split(namespacedName, "/")
	if len(parts) > 1 {
		return types.NamespacedName{
			Namespace: parts[0],
			Name:      parts[1],
		}, nil
	}

	return types.NamespacedName{
		Namespace: "",
		Name:      "",
	}, errors.New(fmt.Sprintf("too few parts of namespacedname: %s", namespacedName))

}

func ParseNamespacedNameContainer(namespacedName string) (types.NamespacedName, string, error) {
	parts := strings.Split(namespacedName, "/")
	if len(parts) > 2 {
		return types.NamespacedName{
			Namespace: parts[0],
			Name:      parts[1],
		}, strings.Join(parts[2:], ""), nil
	}

	return types.NamespacedName{
		Namespace: "",
		Name:      "",
	}, "", errors.New(fmt.Sprintf("too few parts of namespacedname: %s", namespacedName))

}
