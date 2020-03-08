// Code generated by protoc-gen-go. DO NOT EDIT.
// source: chaosdaemon.proto

package chaosdaemon

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Rule_Action int32

const (
	Rule_ADD    Rule_Action = 0
	Rule_DELETE Rule_Action = 1
)

var Rule_Action_name = map[int32]string{
	0: "ADD",
	1: "DELETE",
}

var Rule_Action_value = map[string]int32{
	"ADD":    0,
	"DELETE": 1,
}

func (x Rule_Action) String() string {
	return proto.EnumName(Rule_Action_name, int32(x))
}

func (Rule_Action) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{7, 0}
}

type Rule_Direction int32

const (
	Rule_INPUT  Rule_Direction = 0
	Rule_OUTPUT Rule_Direction = 1
)

var Rule_Direction_name = map[int32]string{
	0: "INPUT",
	1: "OUTPUT",
}

var Rule_Direction_value = map[string]int32{
	"INPUT":  0,
	"OUTPUT": 1,
}

func (x Rule_Direction) String() string {
	return proto.EnumName(Rule_Direction_name, int32(x))
}

func (Rule_Direction) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{7, 1}
}

type ContainerAction_Action int32

const (
	ContainerAction_KILL   ContainerAction_Action = 0
	ContainerAction_GETPID ContainerAction_Action = 1
)

var ContainerAction_Action_name = map[int32]string{
	0: "KILL",
	1: "GETPID",
}

var ContainerAction_Action_value = map[string]int32{
	"KILL":   0,
	"GETPID": 1,
}

func (x ContainerAction_Action) String() string {
	return proto.EnumName(ContainerAction_Action_name, int32(x))
}

func (ContainerAction_Action) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{9, 0}
}

type ContainerRequest struct {
	Action               *ContainerAction `protobuf:"bytes,1,opt,name=action,proto3" json:"action,omitempty"`
	ContainerId          string           `protobuf:"bytes,2,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *ContainerRequest) Reset()         { *m = ContainerRequest{} }
func (m *ContainerRequest) String() string { return proto.CompactTextString(m) }
func (*ContainerRequest) ProtoMessage()    {}
func (*ContainerRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{0}
}

func (m *ContainerRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContainerRequest.Unmarshal(m, b)
}
func (m *ContainerRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContainerRequest.Marshal(b, m, deterministic)
}
func (m *ContainerRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContainerRequest.Merge(m, src)
}
func (m *ContainerRequest) XXX_Size() int {
	return xxx_messageInfo_ContainerRequest.Size(m)
}
func (m *ContainerRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ContainerRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ContainerRequest proto.InternalMessageInfo

func (m *ContainerRequest) GetAction() *ContainerAction {
	if m != nil {
		return m.Action
	}
	return nil
}

func (m *ContainerRequest) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

type ContainerResponse struct {
	Pid                  uint32   `protobuf:"varint,1,opt,name=pid,proto3" json:"pid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ContainerResponse) Reset()         { *m = ContainerResponse{} }
func (m *ContainerResponse) String() string { return proto.CompactTextString(m) }
func (*ContainerResponse) ProtoMessage()    {}
func (*ContainerResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{1}
}

func (m *ContainerResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContainerResponse.Unmarshal(m, b)
}
func (m *ContainerResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContainerResponse.Marshal(b, m, deterministic)
}
func (m *ContainerResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContainerResponse.Merge(m, src)
}
func (m *ContainerResponse) XXX_Size() int {
	return xxx_messageInfo_ContainerResponse.Size(m)
}
func (m *ContainerResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ContainerResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ContainerResponse proto.InternalMessageInfo

func (m *ContainerResponse) GetPid() uint32 {
	if m != nil {
		return m.Pid
	}
	return 0
}

type NetemRequest struct {
	Netem                *Netem   `protobuf:"bytes,1,opt,name=netem,proto3" json:"netem,omitempty"`
	ContainerId          string   `protobuf:"bytes,2,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NetemRequest) Reset()         { *m = NetemRequest{} }
func (m *NetemRequest) String() string { return proto.CompactTextString(m) }
func (*NetemRequest) ProtoMessage()    {}
func (*NetemRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{2}
}

func (m *NetemRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NetemRequest.Unmarshal(m, b)
}
func (m *NetemRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NetemRequest.Marshal(b, m, deterministic)
}
func (m *NetemRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NetemRequest.Merge(m, src)
}
func (m *NetemRequest) XXX_Size() int {
	return xxx_messageInfo_NetemRequest.Size(m)
}
func (m *NetemRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NetemRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NetemRequest proto.InternalMessageInfo

func (m *NetemRequest) GetNetem() *Netem {
	if m != nil {
		return m.Netem
	}
	return nil
}

func (m *NetemRequest) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

type Netem struct {
	Time                 uint32   `protobuf:"varint,1,opt,name=time,proto3" json:"time,omitempty"`
	Jitter               uint32   `protobuf:"varint,2,opt,name=jitter,proto3" json:"jitter,omitempty"`
	DelayCorr            float32  `protobuf:"fixed32,3,opt,name=delay_corr,json=delayCorr,proto3" json:"delay_corr,omitempty"`
	Limit                uint32   `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
	Loss                 float32  `protobuf:"fixed32,5,opt,name=loss,proto3" json:"loss,omitempty"`
	LossCorr             float32  `protobuf:"fixed32,6,opt,name=loss_corr,json=lossCorr,proto3" json:"loss_corr,omitempty"`
	Gap                  uint32   `protobuf:"varint,7,opt,name=gap,proto3" json:"gap,omitempty"`
	Duplicate            float32  `protobuf:"fixed32,8,opt,name=duplicate,proto3" json:"duplicate,omitempty"`
	DuplicateCorr        float32  `protobuf:"fixed32,9,opt,name=duplicate_corr,json=duplicateCorr,proto3" json:"duplicate_corr,omitempty"`
	Reorder              float32  `protobuf:"fixed32,10,opt,name=reorder,proto3" json:"reorder,omitempty"`
	ReorderCorr          float32  `protobuf:"fixed32,11,opt,name=reorder_corr,json=reorderCorr,proto3" json:"reorder_corr,omitempty"`
	Corrupt              float32  `protobuf:"fixed32,12,opt,name=corrupt,proto3" json:"corrupt,omitempty"`
	CorruptCorr          float32  `protobuf:"fixed32,13,opt,name=corrupt_corr,json=corruptCorr,proto3" json:"corrupt_corr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Netem) Reset()         { *m = Netem{} }
func (m *Netem) String() string { return proto.CompactTextString(m) }
func (*Netem) ProtoMessage()    {}
func (*Netem) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{3}
}

func (m *Netem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Netem.Unmarshal(m, b)
}
func (m *Netem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Netem.Marshal(b, m, deterministic)
}
func (m *Netem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Netem.Merge(m, src)
}
func (m *Netem) XXX_Size() int {
	return xxx_messageInfo_Netem.Size(m)
}
func (m *Netem) XXX_DiscardUnknown() {
	xxx_messageInfo_Netem.DiscardUnknown(m)
}

var xxx_messageInfo_Netem proto.InternalMessageInfo

func (m *Netem) GetTime() uint32 {
	if m != nil {
		return m.Time
	}
	return 0
}

func (m *Netem) GetJitter() uint32 {
	if m != nil {
		return m.Jitter
	}
	return 0
}

func (m *Netem) GetDelayCorr() float32 {
	if m != nil {
		return m.DelayCorr
	}
	return 0
}

func (m *Netem) GetLimit() uint32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *Netem) GetLoss() float32 {
	if m != nil {
		return m.Loss
	}
	return 0
}

func (m *Netem) GetLossCorr() float32 {
	if m != nil {
		return m.LossCorr
	}
	return 0
}

func (m *Netem) GetGap() uint32 {
	if m != nil {
		return m.Gap
	}
	return 0
}

func (m *Netem) GetDuplicate() float32 {
	if m != nil {
		return m.Duplicate
	}
	return 0
}

func (m *Netem) GetDuplicateCorr() float32 {
	if m != nil {
		return m.DuplicateCorr
	}
	return 0
}

func (m *Netem) GetReorder() float32 {
	if m != nil {
		return m.Reorder
	}
	return 0
}

func (m *Netem) GetReorderCorr() float32 {
	if m != nil {
		return m.ReorderCorr
	}
	return 0
}

func (m *Netem) GetCorrupt() float32 {
	if m != nil {
		return m.Corrupt
	}
	return 0
}

func (m *Netem) GetCorruptCorr() float32 {
	if m != nil {
		return m.CorruptCorr
	}
	return 0
}

type IpSetRequest struct {
	Ipset                *IpSet   `protobuf:"bytes,1,opt,name=ipset,proto3" json:"ipset,omitempty"`
	ContainerId          string   `protobuf:"bytes,2,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IpSetRequest) Reset()         { *m = IpSetRequest{} }
func (m *IpSetRequest) String() string { return proto.CompactTextString(m) }
func (*IpSetRequest) ProtoMessage()    {}
func (*IpSetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{4}
}

func (m *IpSetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IpSetRequest.Unmarshal(m, b)
}
func (m *IpSetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IpSetRequest.Marshal(b, m, deterministic)
}
func (m *IpSetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IpSetRequest.Merge(m, src)
}
func (m *IpSetRequest) XXX_Size() int {
	return xxx_messageInfo_IpSetRequest.Size(m)
}
func (m *IpSetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_IpSetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_IpSetRequest proto.InternalMessageInfo

func (m *IpSetRequest) GetIpset() *IpSet {
	if m != nil {
		return m.Ipset
	}
	return nil
}

func (m *IpSetRequest) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

type IpSet struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Ips                  []string `protobuf:"bytes,2,rep,name=ips,proto3" json:"ips,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IpSet) Reset()         { *m = IpSet{} }
func (m *IpSet) String() string { return proto.CompactTextString(m) }
func (*IpSet) ProtoMessage()    {}
func (*IpSet) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{5}
}

func (m *IpSet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IpSet.Unmarshal(m, b)
}
func (m *IpSet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IpSet.Marshal(b, m, deterministic)
}
func (m *IpSet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IpSet.Merge(m, src)
}
func (m *IpSet) XXX_Size() int {
	return xxx_messageInfo_IpSet.Size(m)
}
func (m *IpSet) XXX_DiscardUnknown() {
	xxx_messageInfo_IpSet.DiscardUnknown(m)
}

var xxx_messageInfo_IpSet proto.InternalMessageInfo

func (m *IpSet) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *IpSet) GetIps() []string {
	if m != nil {
		return m.Ips
	}
	return nil
}

type IpTablesRequest struct {
	Rule                 *Rule    `protobuf:"bytes,1,opt,name=rule,proto3" json:"rule,omitempty"`
	ContainerId          string   `protobuf:"bytes,2,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *IpTablesRequest) Reset()         { *m = IpTablesRequest{} }
func (m *IpTablesRequest) String() string { return proto.CompactTextString(m) }
func (*IpTablesRequest) ProtoMessage()    {}
func (*IpTablesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{6}
}

func (m *IpTablesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_IpTablesRequest.Unmarshal(m, b)
}
func (m *IpTablesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_IpTablesRequest.Marshal(b, m, deterministic)
}
func (m *IpTablesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_IpTablesRequest.Merge(m, src)
}
func (m *IpTablesRequest) XXX_Size() int {
	return xxx_messageInfo_IpTablesRequest.Size(m)
}
func (m *IpTablesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_IpTablesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_IpTablesRequest proto.InternalMessageInfo

func (m *IpTablesRequest) GetRule() *Rule {
	if m != nil {
		return m.Rule
	}
	return nil
}

func (m *IpTablesRequest) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

type Rule struct {
	Action               Rule_Action    `protobuf:"varint,1,opt,name=action,proto3,enum=chaosdaemon.Rule_Action" json:"action,omitempty"`
	Direction            Rule_Direction `protobuf:"varint,2,opt,name=direction,proto3,enum=chaosdaemon.Rule_Direction" json:"direction,omitempty"`
	Set                  string         `protobuf:"bytes,3,opt,name=set,proto3" json:"set,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Rule) Reset()         { *m = Rule{} }
func (m *Rule) String() string { return proto.CompactTextString(m) }
func (*Rule) ProtoMessage()    {}
func (*Rule) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{7}
}

func (m *Rule) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Rule.Unmarshal(m, b)
}
func (m *Rule) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Rule.Marshal(b, m, deterministic)
}
func (m *Rule) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Rule.Merge(m, src)
}
func (m *Rule) XXX_Size() int {
	return xxx_messageInfo_Rule.Size(m)
}
func (m *Rule) XXX_DiscardUnknown() {
	xxx_messageInfo_Rule.DiscardUnknown(m)
}

var xxx_messageInfo_Rule proto.InternalMessageInfo

func (m *Rule) GetAction() Rule_Action {
	if m != nil {
		return m.Action
	}
	return Rule_ADD
}

func (m *Rule) GetDirection() Rule_Direction {
	if m != nil {
		return m.Direction
	}
	return Rule_INPUT
}

func (m *Rule) GetSet() string {
	if m != nil {
		return m.Set
	}
	return ""
}

type TimeRequest struct {
	ContainerId          string   `protobuf:"bytes,1,opt,name=container_id,json=containerId,proto3" json:"container_id,omitempty"`
	Sec                  int64    `protobuf:"varint,2,opt,name=sec,proto3" json:"sec,omitempty"`
	Nsec                 int64    `protobuf:"varint,3,opt,name=nsec,proto3" json:"nsec,omitempty"`
	ClkIdsMask           uint64   `protobuf:"varint,4,opt,name=clk_ids_mask,json=clkIdsMask,proto3" json:"clk_ids_mask,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TimeRequest) Reset()         { *m = TimeRequest{} }
func (m *TimeRequest) String() string { return proto.CompactTextString(m) }
func (*TimeRequest) ProtoMessage()    {}
func (*TimeRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{8}
}

func (m *TimeRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TimeRequest.Unmarshal(m, b)
}
func (m *TimeRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TimeRequest.Marshal(b, m, deterministic)
}
func (m *TimeRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TimeRequest.Merge(m, src)
}
func (m *TimeRequest) XXX_Size() int {
	return xxx_messageInfo_TimeRequest.Size(m)
}
func (m *TimeRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TimeRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TimeRequest proto.InternalMessageInfo

func (m *TimeRequest) GetContainerId() string {
	if m != nil {
		return m.ContainerId
	}
	return ""
}

func (m *TimeRequest) GetSec() int64 {
	if m != nil {
		return m.Sec
	}
	return 0
}

func (m *TimeRequest) GetNsec() int64 {
	if m != nil {
		return m.Nsec
	}
	return 0
}

func (m *TimeRequest) GetClkIdsMask() uint64 {
	if m != nil {
		return m.ClkIdsMask
	}
	return 0
}

type ContainerAction struct {
	Action               ContainerAction_Action `protobuf:"varint,1,opt,name=action,proto3,enum=chaosdaemon.ContainerAction_Action" json:"action,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *ContainerAction) Reset()         { *m = ContainerAction{} }
func (m *ContainerAction) String() string { return proto.CompactTextString(m) }
func (*ContainerAction) ProtoMessage()    {}
func (*ContainerAction) Descriptor() ([]byte, []int) {
	return fileDescriptor_143136706133b591, []int{9}
}

func (m *ContainerAction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ContainerAction.Unmarshal(m, b)
}
func (m *ContainerAction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ContainerAction.Marshal(b, m, deterministic)
}
func (m *ContainerAction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContainerAction.Merge(m, src)
}
func (m *ContainerAction) XXX_Size() int {
	return xxx_messageInfo_ContainerAction.Size(m)
}
func (m *ContainerAction) XXX_DiscardUnknown() {
	xxx_messageInfo_ContainerAction.DiscardUnknown(m)
}

var xxx_messageInfo_ContainerAction proto.InternalMessageInfo

func (m *ContainerAction) GetAction() ContainerAction_Action {
	if m != nil {
		return m.Action
	}
	return ContainerAction_KILL
}

func init() {
	proto.RegisterEnum("chaosdaemon.Rule_Action", Rule_Action_name, Rule_Action_value)
	proto.RegisterEnum("chaosdaemon.Rule_Direction", Rule_Direction_name, Rule_Direction_value)
	proto.RegisterEnum("chaosdaemon.ContainerAction_Action", ContainerAction_Action_name, ContainerAction_Action_value)
	proto.RegisterType((*ContainerRequest)(nil), "chaosdaemon.ContainerRequest")
	proto.RegisterType((*ContainerResponse)(nil), "chaosdaemon.ContainerResponse")
	proto.RegisterType((*NetemRequest)(nil), "chaosdaemon.NetemRequest")
	proto.RegisterType((*Netem)(nil), "chaosdaemon.Netem")
	proto.RegisterType((*IpSetRequest)(nil), "chaosdaemon.IpSetRequest")
	proto.RegisterType((*IpSet)(nil), "chaosdaemon.IpSet")
	proto.RegisterType((*IpTablesRequest)(nil), "chaosdaemon.IpTablesRequest")
	proto.RegisterType((*Rule)(nil), "chaosdaemon.Rule")
	proto.RegisterType((*TimeRequest)(nil), "chaosdaemon.TimeRequest")
	proto.RegisterType((*ContainerAction)(nil), "chaosdaemon.ContainerAction")
}

func init() { proto.RegisterFile("chaosdaemon.proto", fileDescriptor_143136706133b591) }

var fileDescriptor_143136706133b591 = []byte{
	// 748 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x55, 0xdd, 0x4e, 0xdb, 0x4a,
	0x10, 0xc6, 0x71, 0x12, 0xe2, 0x31, 0x81, 0x64, 0x75, 0x84, 0xcc, 0xaf, 0x72, 0x7c, 0x84, 0x94,
	0x9b, 0x13, 0x2a, 0xda, 0x9b, 0xaa, 0x17, 0x15, 0x4d, 0x52, 0x1a, 0x81, 0x00, 0x99, 0xd0, 0x1b,
	0x2e, 0x90, 0xb1, 0x87, 0xb0, 0xad, 0x63, 0xbb, 0xf6, 0xa6, 0x12, 0x6f, 0x58, 0xf5, 0x51, 0xfa,
	0x14, 0xd5, 0xce, 0xda, 0x6e, 0x12, 0x52, 0x40, 0xe2, 0x2a, 0xb3, 0x33, 0xf3, 0x7d, 0xf3, 0xed,
	0xcc, 0x78, 0x03, 0x4d, 0xef, 0xce, 0x8d, 0x52, 0xdf, 0xc5, 0x71, 0x14, 0x76, 0xe2, 0x24, 0x12,
	0x11, 0x33, 0xa7, 0x5c, 0x9b, 0x5b, 0xa3, 0x28, 0x1a, 0x05, 0xb8, 0x4f, 0xa1, 0x9b, 0xc9, 0xed,
	0x3e, 0x8e, 0x63, 0x71, 0xaf, 0x32, 0xed, 0xaf, 0xd0, 0xe8, 0x46, 0xa1, 0x70, 0x79, 0x88, 0x89,
	0x83, 0xdf, 0x26, 0x98, 0x0a, 0xf6, 0x06, 0xaa, 0xae, 0x27, 0x78, 0x14, 0x5a, 0x5a, 0x4b, 0x6b,
	0x9b, 0x07, 0xdb, 0x9d, 0xe9, 0x0a, 0x45, 0xfa, 0x21, 0xe5, 0x38, 0x59, 0x2e, 0xfb, 0x17, 0x56,
	0xbc, 0x3c, 0x74, 0xcd, 0x7d, 0xab, 0xd4, 0xd2, 0xda, 0x86, 0x63, 0x16, 0xbe, 0x81, 0x6f, 0xef,
	0x41, 0x73, 0xaa, 0x58, 0x1a, 0x47, 0x61, 0x8a, 0xac, 0x01, 0x7a, 0xcc, 0x7d, 0x2a, 0x55, 0x77,
	0xa4, 0x69, 0x5f, 0xc1, 0xca, 0x29, 0x0a, 0x1c, 0xe7, 0x7a, 0xda, 0x50, 0x09, 0xe5, 0x39, 0x93,
	0xc3, 0x66, 0xe4, 0xa8, 0x4c, 0x95, 0xf0, 0x1c, 0x0d, 0xbf, 0x4a, 0x50, 0x21, 0x0c, 0x63, 0x50,
	0x16, 0x7c, 0x8c, 0x59, 0x65, 0xb2, 0xd9, 0x3a, 0x54, 0xbf, 0x70, 0x21, 0x30, 0x21, 0x68, 0xdd,
	0xc9, 0x4e, 0x6c, 0x07, 0xc0, 0xc7, 0xc0, 0xbd, 0xbf, 0xf6, 0xa2, 0x24, 0xb1, 0xf4, 0x96, 0xd6,
	0x2e, 0x39, 0x06, 0x79, 0xba, 0x51, 0x92, 0xb0, 0x7f, 0xa0, 0x12, 0xf0, 0x31, 0x17, 0x56, 0x99,
	0x50, 0xea, 0x20, 0x0b, 0x04, 0x51, 0x9a, 0x5a, 0x15, 0x4a, 0x27, 0x9b, 0x6d, 0x81, 0x21, 0x7f,
	0x15, 0x4f, 0x95, 0x02, 0x35, 0xe9, 0x20, 0x9a, 0x06, 0xe8, 0x23, 0x37, 0xb6, 0x96, 0x55, 0x2b,
	0x46, 0x6e, 0xcc, 0xb6, 0xc1, 0xf0, 0x27, 0x71, 0xc0, 0x3d, 0x57, 0xa0, 0x55, 0xcb, 0xca, 0xe6,
	0x0e, 0xb6, 0x07, 0xab, 0xc5, 0x41, 0x31, 0x1a, 0x94, 0x52, 0x2f, 0xbc, 0x44, 0x6b, 0xc1, 0x72,
	0x82, 0x51, 0xe2, 0x63, 0x62, 0x01, 0xc5, 0xf3, 0xa3, 0xec, 0x57, 0x66, 0x2a, 0xb8, 0x49, 0x61,
	0x33, 0xf3, 0xe5, 0x60, 0x19, 0x9a, 0xc4, 0xc2, 0x5a, 0x51, 0xe0, 0xec, 0xa8, 0x9a, 0x4d, 0xa6,
	0x02, 0xd7, 0x15, 0x38, 0xf3, 0x49, 0xb0, 0x9c, 0xe4, 0x20, 0xbe, 0x40, 0x31, 0x35, 0x49, 0x1e,
	0xa7, 0x28, 0x16, 0x4e, 0x52, 0x65, 0xaa, 0x84, 0xe7, 0x4c, 0xf2, 0x7f, 0xa8, 0x10, 0x44, 0xf6,
	0x39, 0x74, 0xb3, 0x41, 0x1a, 0x0e, 0xd9, 0xb2, 0x95, 0x3c, 0x4e, 0xad, 0x52, 0x4b, 0x6f, 0x1b,
	0x8e, 0x34, 0xed, 0x2b, 0x58, 0x1b, 0xc4, 0x43, 0xf7, 0x26, 0xc0, 0x34, 0x97, 0xb3, 0x07, 0xe5,
	0x64, 0x12, 0x60, 0xa6, 0xa6, 0x39, 0xa3, 0xc6, 0x99, 0x04, 0xe8, 0x50, 0xf8, 0x39, 0x5a, 0x7e,
	0x68, 0x50, 0x96, 0x08, 0xf6, 0x6a, 0xe6, 0xdb, 0x59, 0x3d, 0xb0, 0x1e, 0x90, 0x76, 0xe6, 0xbe,
	0x9b, 0xb7, 0x60, 0xf8, 0x3c, 0x41, 0x05, 0x2a, 0x11, 0x68, 0xeb, 0x21, 0xa8, 0x97, 0xa7, 0x38,
	0x7f, 0xb2, 0xe5, 0x25, 0x65, 0x33, 0x75, 0xd2, 0x23, 0x4d, 0x7b, 0x07, 0xaa, 0x8a, 0x9e, 0x2d,
	0x83, 0x7e, 0xd8, 0xeb, 0x35, 0x96, 0x18, 0x40, 0xb5, 0xd7, 0x3f, 0xe9, 0x0f, 0xfb, 0x0d, 0xcd,
	0xb6, 0xc1, 0x28, 0x88, 0x98, 0x01, 0x95, 0xc1, 0xe9, 0xf9, 0xe5, 0x50, 0xe5, 0x9c, 0x5d, 0x0e,
	0xa5, 0xad, 0xd9, 0x9f, 0xc1, 0x1c, 0xf2, 0x31, 0xe6, 0x3d, 0x9a, 0xbf, 0xbc, 0xf6, 0xe0, 0xf2,
	0x4a, 0x86, 0x47, 0xda, 0x75, 0x29, 0xc3, 0xa3, 0x89, 0x48, 0x97, 0x4e, 0x2e, 0xb2, 0xed, 0x10,
	0xd6, 0xe6, 0x9e, 0x0e, 0xf6, 0x6e, 0xae, 0x59, 0xff, 0x3d, 0xf6, 0xd0, 0xcc, 0xf5, 0xcd, 0xde,
	0x2d, 0xae, 0x5a, 0x83, 0xf2, 0xf1, 0xe0, 0xe4, 0x44, 0xdd, 0xe3, 0xa8, 0x3f, 0x3c, 0x1f, 0xf4,
	0x1a, 0xda, 0xc1, 0xcf, 0x32, 0x98, 0x5d, 0x49, 0xd7, 0x23, 0x3a, 0xf6, 0x1e, 0x6a, 0x17, 0x28,
	0xd4, 0xa7, 0xbf, 0xb1, 0xe0, 0x09, 0x51, 0xf7, 0xdd, 0x5c, 0xef, 0xa8, 0xe7, 0xb2, 0x93, 0x3f,
	0x97, 0x9d, 0xbe, 0x7c, 0x2e, 0xed, 0x25, 0xf6, 0x01, 0xcc, 0x1e, 0x06, 0x28, 0xf0, 0x05, 0x1c,
	0x87, 0x00, 0x1f, 0x83, 0x49, 0x7a, 0xa7, 0x16, 0x77, 0x63, 0xc1, 0xfe, 0x3f, 0x49, 0x71, 0x04,
	0xf5, 0x8c, 0x42, 0xd0, 0x32, 0xb3, 0xed, 0x39, 0x96, 0x99, 0x1d, 0x7f, 0x84, 0xa8, 0x0b, 0xf5,
	0x0b, 0x14, 0x72, 0xd6, 0x67, 0xb7, 0xb7, 0xf2, 0x9b, 0x9b, 0xdd, 0xd5, 0xa9, 0x25, 0x78, 0x54,
	0x4d, 0xd3, 0x41, 0x2f, 0xfa, 0x8e, 0xc9, 0x0b, 0x89, 0x3e, 0x41, 0xbd, 0x18, 0xf8, 0x31, 0x0f,
	0x02, 0xb6, 0xb3, 0x78, 0x19, 0x9e, 0x66, 0x72, 0xa6, 0x16, 0xed, 0x08, 0xc5, 0x39, 0xf7, 0x9f,
	0xe2, 0xda, 0xfd, 0x5b, 0x58, 0xfd, 0x45, 0xd9, 0x4b, 0x37, 0x55, 0xaa, 0xf2, 0xfa, 0x77, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xd2, 0xe9, 0x6e, 0x83, 0x6c, 0x07, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ChaosDaemonClient is the client API for ChaosDaemon service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChaosDaemonClient interface {
	SetNetem(ctx context.Context, in *NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	DeleteNetem(ctx context.Context, in *NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	FlushIpSet(ctx context.Context, in *IpSetRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	FlushIptables(ctx context.Context, in *IpTablesRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	SetTimeOffset(ctx context.Context, in *TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	RecoverTimeOffset(ctx context.Context, in *TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	ContainerKill(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	ContainerGetPid(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerResponse, error)
}

type chaosDaemonClient struct {
	cc *grpc.ClientConn
}

func NewChaosDaemonClient(cc *grpc.ClientConn) ChaosDaemonClient {
	return &chaosDaemonClient{cc}
}

func (c *chaosDaemonClient) SetNetem(ctx context.Context, in *NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/SetNetem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) DeleteNetem(ctx context.Context, in *NetemRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/DeleteNetem", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) FlushIpSet(ctx context.Context, in *IpSetRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/FlushIpSet", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) FlushIptables(ctx context.Context, in *IpTablesRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/FlushIptables", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) SetTimeOffset(ctx context.Context, in *TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/SetTimeOffset", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) RecoverTimeOffset(ctx context.Context, in *TimeRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/RecoverTimeOffset", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) ContainerKill(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/ContainerKill", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chaosDaemonClient) ContainerGetPid(ctx context.Context, in *ContainerRequest, opts ...grpc.CallOption) (*ContainerResponse, error) {
	out := new(ContainerResponse)
	err := c.cc.Invoke(ctx, "/chaosdaemon.ChaosDaemon/ContainerGetPid", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChaosDaemonServer is the server API for ChaosDaemon service.
type ChaosDaemonServer interface {
	SetNetem(context.Context, *NetemRequest) (*empty.Empty, error)
	DeleteNetem(context.Context, *NetemRequest) (*empty.Empty, error)
	FlushIpSet(context.Context, *IpSetRequest) (*empty.Empty, error)
	FlushIptables(context.Context, *IpTablesRequest) (*empty.Empty, error)
	SetTimeOffset(context.Context, *TimeRequest) (*empty.Empty, error)
	RecoverTimeOffset(context.Context, *TimeRequest) (*empty.Empty, error)
	ContainerKill(context.Context, *ContainerRequest) (*empty.Empty, error)
	ContainerGetPid(context.Context, *ContainerRequest) (*ContainerResponse, error)
}

// UnimplementedChaosDaemonServer can be embedded to have forward compatible implementations.
type UnimplementedChaosDaemonServer struct {
}

func (*UnimplementedChaosDaemonServer) SetNetem(ctx context.Context, req *NetemRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetNetem not implemented")
}
func (*UnimplementedChaosDaemonServer) DeleteNetem(ctx context.Context, req *NetemRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteNetem not implemented")
}
func (*UnimplementedChaosDaemonServer) FlushIpSet(ctx context.Context, req *IpSetRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FlushIpSet not implemented")
}
func (*UnimplementedChaosDaemonServer) FlushIptables(ctx context.Context, req *IpTablesRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FlushIptables not implemented")
}
func (*UnimplementedChaosDaemonServer) SetTimeOffset(ctx context.Context, req *TimeRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetTimeOffset not implemented")
}
func (*UnimplementedChaosDaemonServer) RecoverTimeOffset(ctx context.Context, req *TimeRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecoverTimeOffset not implemented")
}
func (*UnimplementedChaosDaemonServer) ContainerKill(ctx context.Context, req *ContainerRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ContainerKill not implemented")
}
func (*UnimplementedChaosDaemonServer) ContainerGetPid(ctx context.Context, req *ContainerRequest) (*ContainerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ContainerGetPid not implemented")
}

func RegisterChaosDaemonServer(s *grpc.Server, srv ChaosDaemonServer) {
	s.RegisterService(&_ChaosDaemon_serviceDesc, srv)
}

func _ChaosDaemon_SetNetem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).SetNetem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/SetNetem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).SetNetem(ctx, req.(*NetemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_DeleteNetem_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NetemRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).DeleteNetem(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/DeleteNetem",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).DeleteNetem(ctx, req.(*NetemRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_FlushIpSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).FlushIpSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/FlushIpSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).FlushIpSet(ctx, req.(*IpSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_FlushIptables_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IpTablesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).FlushIptables(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/FlushIptables",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).FlushIptables(ctx, req.(*IpTablesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_SetTimeOffset_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TimeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).SetTimeOffset(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/SetTimeOffset",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).SetTimeOffset(ctx, req.(*TimeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_RecoverTimeOffset_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TimeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).RecoverTimeOffset(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/RecoverTimeOffset",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).RecoverTimeOffset(ctx, req.(*TimeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_ContainerKill_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).ContainerKill(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/ContainerKill",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).ContainerKill(ctx, req.(*ContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChaosDaemon_ContainerGetPid_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ContainerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChaosDaemonServer).ContainerGetPid(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/chaosdaemon.ChaosDaemon/ContainerGetPid",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChaosDaemonServer).ContainerGetPid(ctx, req.(*ContainerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ChaosDaemon_serviceDesc = grpc.ServiceDesc{
	ServiceName: "chaosdaemon.ChaosDaemon",
	HandlerType: (*ChaosDaemonServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetNetem",
			Handler:    _ChaosDaemon_SetNetem_Handler,
		},
		{
			MethodName: "DeleteNetem",
			Handler:    _ChaosDaemon_DeleteNetem_Handler,
		},
		{
			MethodName: "FlushIpSet",
			Handler:    _ChaosDaemon_FlushIpSet_Handler,
		},
		{
			MethodName: "FlushIptables",
			Handler:    _ChaosDaemon_FlushIptables_Handler,
		},
		{
			MethodName: "SetTimeOffset",
			Handler:    _ChaosDaemon_SetTimeOffset_Handler,
		},
		{
			MethodName: "RecoverTimeOffset",
			Handler:    _ChaosDaemon_RecoverTimeOffset_Handler,
		},
		{
			MethodName: "ContainerKill",
			Handler:    _ChaosDaemon_ContainerKill_Handler,
		},
		{
			MethodName: "ContainerGetPid",
			Handler:    _ChaosDaemon_ContainerGetPid_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chaosdaemon.proto",
}
