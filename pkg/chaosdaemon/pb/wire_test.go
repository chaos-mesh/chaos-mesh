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

package pb

import (
	"testing"

	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"

	"github.com/chaos-mesh/chaos-mesh/test/pkg/protocontract"
)

func TestDaemonWireContract(t *testing.T) {
	protocontract.CheckWireContract(t, File_chaosdaemon_proto)
}

func TestStressRequestWithLegacyFieldNumbering(t *testing.T) {
	g := NewWithT(t)
	// scope=Pod, target="pod", obsolete field 6, oomScoreAdj=250 at field 7.
	wire := []byte{0x08, 0x01, 0x12, 0x03, 'p', 'o', 'd', 0x30, 0x09, 0x38, 0xfa, 0x01}
	request := &ExecStressRequest{}
	g.Expect(proto.Unmarshal(wire, request)).To(Succeed())
	g.Expect(request.GetScope()).To(Equal(ExecStressRequest_POD))
	g.Expect(request.GetTarget()).To(Equal("pod"))
	g.Expect(request.GetOomScoreAdj()).To(Equal(int32(250)))
	g.Expect(request.GetEnterNS()).To(BeFalse())
	g.Expect([]byte(request.ProtoReflect().GetUnknown())).To(Equal([]byte{0x30, 0x09}))
}
