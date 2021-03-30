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

package mock

import (
	"path"
	"reflect"
	"sync"

	"github.com/pingcap/failpoint"
)

// Finalizer represent the function that clean a mock point
type Finalizer func() error

type mockPoints struct {
	m map[string]interface{}
	l sync.Mutex
}

func (p *mockPoints) set(fpname string, value interface{}) {
	p.l.Lock()
	defer p.l.Unlock()

	p.m[fpname] = value
}

func (p *mockPoints) get(fpname string) interface{} {
	p.l.Lock()
	defer p.l.Unlock()

	return p.m[fpname]
}

func (p *mockPoints) clr(fpname string) {
	p.l.Lock()
	defer p.l.Unlock()

	delete(p.m, fpname)
}

var points = mockPoints{m: make(map[string]interface{})}

// On inject a failpoint
func On(fpname string) interface{} {
	var ret interface{}
	failpoint.Inject(fpname, func() {
		ret = points.get(fpname)
	})
	return ret
}

// With enable failpoint and provide a value
func With(fpname string, value interface{}) Finalizer {
	if err := failpoint.Enable(failpath(fpname), "return(true)"); err != nil {
		panic(err)
	}
	points.set(fpname, value)
	return func() error { return Reset(fpname) }
}

// Reset disable failpoint and remove mock value
func Reset(fpname string) error {
	if err := failpoint.Disable(failpath(fpname)); err != nil {
		return err
	}
	points.clr(fpname)
	return nil
}

func failpath(fpname string) string {
	type em struct{}
	return path.Join(reflect.TypeOf(em{}).PkgPath(), fpname)
}
