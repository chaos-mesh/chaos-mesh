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

package label

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestLabelString(t *testing.T) {
	g := NewGomegaWithT(t)

	la := Label(make(map[string]string))
	la["test-label-1"] = "t1"
	la["test-label-2"] = "t2"

	g.Expect(strings.Contains(la.String(), "test-label-1=t1")).To(Equal(true))
	g.Expect(strings.Contains(la.String(), "test-label-2=t2")).To(Equal(true))
	g.Expect(strings.Contains(la.String(), ",")).To(Equal(true))

	g.Expect(len(la.String())).To(Equal(len("test-label-1=t1,test-label-2=t2")))

	la[""] = "t3"
	g.Expect(len(la.String())).To(Equal(len("test-label-1=t1,test-label-2=t2")))
	g.Expect(strings.Contains(la.String(), "t3")).To(Equal(false))
}

func TestParseLabel(t *testing.T) {
	g := NewGomegaWithT(t)

	label1 := "k1=v1"
	labelMap1 := Label{"k1": "v1"}
	result, err := ParseLabel(label1)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(reflect.DeepEqual(labelMap1, result)).To(Equal(true))

	label2 := "k1=v1,k2=v2"
	labelMap2 := Label{"k1": "v1", "k2": "v2"}
	result, err = ParseLabel(label2)

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(reflect.DeepEqual(labelMap2, result)).To(Equal(true))

	label3 := ""
	_, err = ParseLabel(label3)
	g.Expect(err).ToNot(HaveOccurred())

	label4 := "k1=v2,,"
	_, err = ParseLabel(label4)
	g.Expect(err).To(HaveOccurred())
	g.Expect(strings.Contains(err.Error(), "invalid labels")).To(BeTrue())
}
