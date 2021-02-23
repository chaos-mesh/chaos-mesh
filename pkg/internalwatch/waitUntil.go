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

package internalwatch

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

var notifyChannel = make(chan runtime.Object)

func Notify(obj runtime.Object) {
	notifyChannel <- obj
}

func WaitUntil(ctx context.Context, obj runtime.Object, until func(object runtime.Object) (bool, error)) error {
	key, err := getKeyFromObj(obj)

	if err != nil {
		return err
	}
	for {
		select {
		case newObj := <-notifyChannel:
			newKey, err := getKeyFromObj(newObj)
			if err != nil {
				return err
			}

			if *newKey == *key {
				finished, err := until(newObj)
				if err != nil {
					return err
				}

				if finished {
					return nil
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func getKeyFromObj(obj runtime.Object) (*types.NamespacedName, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	return &types.NamespacedName{
		Name:      accessor.GetName(),
		Namespace: accessor.GetNamespace(),
	}, nil
}
