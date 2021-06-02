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

package utils

import "fmt"

type failToFindContainer struct {
	namespace     string
	name          string
	containerName string

	err error
}

func NewFailToFindContainer(namespace string, name string, containerName string, err error) error {
	return &failToFindContainer{
		namespace,
		name,
		containerName,
		err,
	}
}

func (e *failToFindContainer) Error() string {
	if e.err == nil {
		return fmt.Sprintf("fail to find container %s on pod %s/%s", e.containerName, e.namespace, e.name)
	}

	return e.err.Error()
}

func IsFailToGet(e error) bool {
	switch e.(type) {
	case *failToFindContainer:
		return true
	default:
		return false
	}
}
