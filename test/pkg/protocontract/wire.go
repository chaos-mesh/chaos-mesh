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

package protocontract

import (
	"fmt"
	"testing"

	"github.com/onsi/gomega"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// CheckWireContract exercises the generated RPC messages against a descriptor-driven
// peer, including messages that carry fields from a newer peer.
func CheckWireContract(t *testing.T, file protoreflect.FileDescriptor) {
	t.Helper()
	checkMessages(t, file.Messages())
}

func checkMessages(t *testing.T, messages protoreflect.MessageDescriptors) {
	for i := range messages.Len() {
		descriptor := messages.Get(i)
		t.Run(string(descriptor.Name()), func(t *testing.T) {
			g := gomega.NewWithT(t)
			messageType, err := protoregistry.GlobalTypes.FindMessageByName(descriptor.FullName())
			g.Expect(err).NotTo(gomega.HaveOccurred())

			peer := dynamicpb.NewMessage(descriptor)
			fillMessage(t, peer, 1)
			// Pick a field number absent from the current schema to simulate a newer peer.
			unknown := protowire.AppendVarint(protowire.AppendTag(nil, 50000, protowire.VarintType), 321)
			g.Expect(descriptor.Fields().ByNumber(50000)).To(gomega.BeNil())
			peer.SetUnknown(unknown)
			wire, err := proto.Marshal(peer)
			g.Expect(err).NotTo(gomega.HaveOccurred())

			generated := messageType.New().Interface()
			g.Expect(proto.Unmarshal(wire, generated)).To(gomega.Succeed())
			g.Expect(proto.Equal(peer, generated)).To(gomega.BeTrue(), "generated message lost fields: %v", generated)

			reply, err := proto.Marshal(generated)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			response := dynamicpb.NewMessage(descriptor)
			g.Expect(proto.Unmarshal(reply, response)).To(gomega.Succeed())
			g.Expect(proto.Equal(peer, response)).To(gomega.BeTrue(), "peer received a changed message: %v", response)

			checkMessages(t, descriptor.Messages())
		})
	}
}

func fillMessage(t *testing.T, message *dynamicpb.Message, seed int) {
	t.Helper()
	fields := message.Descriptor().Fields()
	for i := range fields.Len() {
		field := fields.Get(i)
		switch {
		case field.IsMap():
			t.Fatalf("add fixture values for map field %s", field.FullName())
		case field.IsList():
			list := message.Mutable(field).List()
			list.Append(fieldValue(t, field, seed+i))
			list.Append(fieldValue(t, field, seed+i+1))
		default:
			message.Set(field, fieldValue(t, field, seed+i))
		}
	}
}

func fieldValue(t *testing.T, field protoreflect.FieldDescriptor, seed int) protoreflect.Value {
	t.Helper()
	switch field.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(true)
	case protoreflect.EnumKind:
		// Proto3 peers must retain enum numbers that their schema does not yet name.
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(100 + seed))
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(-int32(seed))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(-int64(seed))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(uint32(seed))
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(1<<40 + uint64(seed))
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(float32(seed) + 0.25)
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(float64(seed) + 0.25)
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(fmt.Sprintf("%s-%d", field.Name(), seed))
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes([]byte{0, byte(seed), 255})
	case protoreflect.MessageKind:
		nested := dynamicpb.NewMessage(field.Message())
		fillMessage(t, nested, seed)
		return protoreflect.ValueOfMessage(nested)
	default:
		t.Fatalf("add fixture values for field %s of kind %s", field.FullName(), field.Kind())
		return protoreflect.Value{}
	}
}
