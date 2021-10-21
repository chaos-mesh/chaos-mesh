// Copyright 2021 Chaos Mesh Authors.
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

package genericwebhook

import (
	"reflect"
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestBfs(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	{
		type Test struct {
			A, B, C int
		}

		testStruct := &Test{1, 2, 3}
		walker := NewFieldWalker(testStruct, func(path *field.Path, obj interface{}, field *reflect.StructField) bool {
			val := obj.(*int)
			*val = 2
			return true
		})
		walker.Walk()
		g.Expect(testStruct).To(gomega.Equal(&Test{2, 2, 2}))
	}

	{
		type Test struct {
			A, B int
			C    *int
		}

		testC := 3
		two := 2
		testStruct := &Test{1, 2, &testC}
		walker := NewFieldWalker(testStruct, func(path *field.Path, obj interface{}, field *reflect.StructField) bool {
			switch obj := obj.(type) {
			case *int:
				*obj = 2
			case **int:
				*obj = &two
			default:
				panic("unexpected type")
			}
			return true
		})
		walker.Walk()
		g.Expect(testStruct).To(gomega.Equal(&Test{2, 2, &two}))
	}

	{
		type Inside struct {
			A, B int
		}
		type DeepTest struct {
			A, B int
			C    Inside
		}

		testStruct := &DeepTest{1, 2, Inside{3, 4}}
		walker := NewFieldWalker(testStruct, func(path *field.Path, obj interface{}, field *reflect.StructField) bool {
			switch obj := obj.(type) {
			case *int:
				*obj = 2
			case *Inside:
				*obj = Inside{2, 2}
			default:
				panic("unexpected type")
			}

			return false
		})
		walker.Walk()
		g.Expect(testStruct).To(gomega.Equal(&DeepTest{2, 2, Inside{2, 2}}))
	}

	{
		type Inside struct {
			A, B int
		}
		type DeepTest struct {
			A, B int
			C    Inside
		}

		testStruct := &DeepTest{1, 2, Inside{3, 4}}
		walker := NewFieldWalker(testStruct, func(path *field.Path, obj interface{}, field *reflect.StructField) bool {
			switch obj := obj.(type) {
			case *int:
				*obj = 2
			case *Inside:
				return true
			default:
				panic("unexpected type")
			}

			return false
		})
		walker.Walk()
		g.Expect(testStruct).To(gomega.Equal(&DeepTest{2, 2, Inside{2, 2}}))
	}
}
