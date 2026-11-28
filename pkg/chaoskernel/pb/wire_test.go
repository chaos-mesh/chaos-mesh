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

package bpfki

import (
	"testing"

	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/chaos-mesh/chaos-mesh/test/pkg/protocontract"
)

func TestKernelWireContract(t *testing.T) {
	protocontract.CheckWireContract(t, File_bpfki_proto)
}

func TestKernelCallChainWithLegacyFieldNumbering(t *testing.T) {
	g := NewWithT(t)
	// pid=42, fail type BIO, callchain=[{funcname: "f"}], probability=0.5, times=3.
	wire := []byte{0x08, 42, 0x18, 2, 0x2a, 3, 0x0a, 1, 'f', 0x35, 0, 0, 0, 0x3f, 0x38, 3}
	request := &FailKernRequest{}
	g.Expect(proto.Unmarshal(wire, request)).To(Succeed())
	g.Expect(request.GetPid()).To(Equal(uint32(42)))
	g.Expect(request.GetFtype()).To(Equal(FailKernRequest_BIO))
	g.Expect(request.GetCallchain()).To(HaveLen(1))
	g.Expect(request.GetCallchain()[0].GetFuncname()).To(Equal("f"))
	g.Expect(request.GetProbability()).To(Equal(float32(0.5)))
	g.Expect(request.GetTimes()).To(Equal(uint32(3)))
}
