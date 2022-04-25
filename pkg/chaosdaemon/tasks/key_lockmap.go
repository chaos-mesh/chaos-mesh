// Copyright 2022 Chaos Mesh Authors.
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

package tasks

import "sync"

type LockMap[K comparable] struct {
	sync.Map
}

func NewLockMap[K comparable]() LockMap[K] {
	return LockMap[K]{
		sync.Map{},
	}
}

func (l *LockMap[K]) Lock(key K) func() {
	value, _ := l.LoadOrStore(key, &sync.Mutex{})
	mtx := value.(*sync.Mutex)
	mtx.Lock()

	return func() {
		if mtx != nil {
			mtx.Unlock()
		}
	}
}

// Del :TODO: Fix bug on deleting a using value
func (l *LockMap[K]) Del(key K) {
	l.Delete(key)
}
