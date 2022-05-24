// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: fetchai/blsgroup/v1/tx.proto

package blsgroup

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	group "github.com/cosmos/cosmos-sdk/x/group"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type MsgVoteResponse struct {
	// proposal is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	// voter is the voter account address.
	Voter string `protobuf:"bytes,2,opt,name=voter,proto3" json:"voter,omitempty"`
	// option is the voter's choice on the proposal.
	Option group.VoteOption `protobuf:"varint,3,opt,name=option,proto3,enum=cosmos.group.v1.VoteOption" json:"option,omitempty"`
	// pub_key is the voter's public key
	PubKey *types.Any `protobuf:"bytes,4,opt,name=pub_key,json=pubKey,proto3" json:"public_key,omitempty"`
	// sig is the individual signature which will be aggregated with other signatures
	Sig []byte `protobuf:"bytes,5,opt,name=sig,proto3" json:"sig,omitempty"`
}

func (m *MsgVoteResponse) Reset()         { *m = MsgVoteResponse{} }
func (m *MsgVoteResponse) String() string { return proto.CompactTextString(m) }
func (*MsgVoteResponse) ProtoMessage()    {}
func (*MsgVoteResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e5efa1b539f0d6e, []int{0}
}
func (m *MsgVoteResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgVoteResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgVoteResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgVoteResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgVoteResponse.Merge(m, src)
}
func (m *MsgVoteResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgVoteResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgVoteResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgVoteResponse proto.InternalMessageInfo

func (m *MsgVoteResponse) GetProposalId() uint64 {
	if m != nil {
		return m.ProposalId
	}
	return 0
}

func (m *MsgVoteResponse) GetVoter() string {
	if m != nil {
		return m.Voter
	}
	return ""
}

func (m *MsgVoteResponse) GetOption() group.VoteOption {
	if m != nil {
		return m.Option
	}
	return group.VOTE_OPTION_UNSPECIFIED
}

func (m *MsgVoteResponse) GetPubKey() *types.Any {
	if m != nil {
		return m.PubKey
	}
	return nil
}

func (m *MsgVoteResponse) GetSig() []byte {
	if m != nil {
		return m.Sig
	}
	return nil
}

// MsgVoteAgg is the Msg/VoteAgg request type.
type MsgVoteAgg struct {
	// sender is the account address who submits the votes
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// proposal is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,2,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	// votes are the list of voters' choices on the proposal.
	Votes []group.VoteOption `protobuf:"varint,3,rep,packed,name=votes,proto3,enum=cosmos.group.v1.VoteOption" json:"votes,omitempty"`
	// agg_sig is the bls aggregated signature for all the votes
	AggSig []byte `protobuf:"bytes,5,opt,name=agg_sig,json=aggSig,proto3" json:"agg_sig,omitempty"`
	// exec defines whether the proposal should be executed
	// immediately after voting or not.
	Exec group.Exec `protobuf:"varint,6,opt,name=exec,proto3,enum=cosmos.group.v1.Exec" json:"exec,omitempty"`
}

func (m *MsgVoteAgg) Reset()         { *m = MsgVoteAgg{} }
func (m *MsgVoteAgg) String() string { return proto.CompactTextString(m) }
func (*MsgVoteAgg) ProtoMessage()    {}
func (*MsgVoteAgg) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e5efa1b539f0d6e, []int{1}
}
func (m *MsgVoteAgg) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgVoteAgg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgVoteAgg.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgVoteAgg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgVoteAgg.Merge(m, src)
}
func (m *MsgVoteAgg) XXX_Size() int {
	return m.Size()
}
func (m *MsgVoteAgg) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgVoteAgg.DiscardUnknown(m)
}

var xxx_messageInfo_MsgVoteAgg proto.InternalMessageInfo

func (m *MsgVoteAgg) GetSender() string {
	if m != nil {
		return m.Sender
	}
	return ""
}

func (m *MsgVoteAgg) GetProposalId() uint64 {
	if m != nil {
		return m.ProposalId
	}
	return 0
}

func (m *MsgVoteAgg) GetVotes() []group.VoteOption {
	if m != nil {
		return m.Votes
	}
	return nil
}

func (m *MsgVoteAgg) GetAggSig() []byte {
	if m != nil {
		return m.AggSig
	}
	return nil
}

func (m *MsgVoteAgg) GetExec() group.Exec {
	if m != nil {
		return m.Exec
	}
	return group.Exec_EXEC_UNSPECIFIED
}

// MsgVoteResponse is the Msg/Vote response type.
type MsgVoteAggResponse struct {
}

func (m *MsgVoteAggResponse) Reset()         { *m = MsgVoteAggResponse{} }
func (m *MsgVoteAggResponse) String() string { return proto.CompactTextString(m) }
func (*MsgVoteAggResponse) ProtoMessage()    {}
func (*MsgVoteAggResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e5efa1b539f0d6e, []int{2}
}
func (m *MsgVoteAggResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgVoteAggResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgVoteAggResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgVoteAggResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgVoteAggResponse.Merge(m, src)
}
func (m *MsgVoteAggResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgVoteAggResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgVoteAggResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgVoteAggResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgVoteResponse)(nil), "fetchai.blsgroup.v1.MsgVoteResponse")
	proto.RegisterType((*MsgVoteAgg)(nil), "fetchai.blsgroup.v1.MsgVoteAgg")
	proto.RegisterType((*MsgVoteAggResponse)(nil), "fetchai.blsgroup.v1.MsgVoteAggResponse")
}

func init() { proto.RegisterFile("fetchai/blsgroup/v1/tx.proto", fileDescriptor_4e5efa1b539f0d6e) }

var fileDescriptor_4e5efa1b539f0d6e = []byte{
	// 518 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x93, 0x4f, 0x8b, 0xd3, 0x40,
	0x18, 0xc6, 0x77, 0xb6, 0xdd, 0x94, 0x9d, 0x15, 0x95, 0xb1, 0xba, 0xb1, 0x2b, 0x69, 0x28, 0x82,
	0x51, 0x74, 0x42, 0xbb, 0x37, 0x0f, 0x42, 0x0b, 0x22, 0x22, 0x8b, 0x90, 0x82, 0x87, 0xbd, 0x94,
	0xfc, 0x99, 0x9d, 0x0d, 0x36, 0x99, 0x21, 0x33, 0x29, 0xcd, 0xd5, 0x4f, 0xe0, 0x47, 0xf1, 0xe0,
	0xd1, 0x0f, 0xe0, 0x71, 0xf1, 0xe4, 0x49, 0xa4, 0x3d, 0x08, 0x7e, 0x08, 0x91, 0x64, 0x66, 0x2c,
	0xb4, 0xc2, 0x9e, 0x32, 0xef, 0x3c, 0xbf, 0xf7, 0xcd, 0xfb, 0x3e, 0x33, 0x03, 0x1f, 0x5c, 0x10,
	0x19, 0x5f, 0x86, 0xa9, 0x1f, 0xcd, 0x05, 0x2d, 0x58, 0xc9, 0xfd, 0xc5, 0xd0, 0x97, 0x4b, 0xcc,
	0x0b, 0x26, 0x19, 0xba, 0xa3, 0x55, 0x6c, 0x54, 0xbc, 0x18, 0xf6, 0xba, 0x94, 0x51, 0xd6, 0xe8,
	0x7e, 0xbd, 0x52, 0x68, 0xcf, 0xa1, 0x8c, 0xd1, 0x39, 0xf1, 0x9b, 0x28, 0x2a, 0x2f, 0xfc, 0xa4,
	0x2c, 0x42, 0x99, 0xb2, 0x5c, 0xeb, 0xfd, 0x6d, 0x5d, 0xa6, 0x19, 0x11, 0x32, 0xcc, 0xb8, 0x06,
	0xee, 0xc7, 0x4c, 0x64, 0x4c, 0xcc, 0x54, 0x65, 0x15, 0x18, 0x69, 0x3b, 0x37, 0xcc, 0x2b, 0x2d,
	0x9d, 0x28, 0xd0, 0xdf, 0xf4, 0x5e, 0x71, 0x62, 0xf2, 0xec, 0x1d, 0x51, 0x0f, 0xd6, 0x3b, 0xd6,
	0x4a, 0x26, 0x68, 0xbd, 0x9f, 0x09, 0xaa, 0x84, 0xc1, 0x1f, 0x00, 0x6f, 0x9d, 0x09, 0xfa, 0x8e,
	0x49, 0x12, 0x10, 0xc1, 0x59, 0x2e, 0x08, 0xea, 0xc3, 0x23, 0x5e, 0x30, 0xce, 0x44, 0x38, 0x9f,
	0xa5, 0x89, 0x0d, 0x5c, 0xe0, 0xb5, 0x03, 0x68, 0xb6, 0x5e, 0x27, 0x08, 0xc3, 0x83, 0x05, 0x93,
	0xa4, 0xb0, 0xf7, 0x5d, 0xe0, 0x1d, 0x4e, 0xec, 0x6f, 0x9f, 0x9f, 0x75, 0xf5, 0x00, 0xe3, 0x24,
	0x29, 0x88, 0x10, 0x53, 0x59, 0xa4, 0x39, 0x0d, 0x14, 0x86, 0x4e, 0xa1, 0xc5, 0x78, 0xed, 0x8d,
	0xdd, 0x72, 0x81, 0x77, 0x73, 0x74, 0x82, 0x35, 0x6d, 0x3c, 0xc6, 0xf5, 0xff, 0xdf, 0x36, 0x48,
	0xa0, 0x51, 0xf4, 0x0a, 0x76, 0x78, 0x19, 0xcd, 0xde, 0x93, 0xca, 0x6e, 0xbb, 0xc0, 0x3b, 0x1a,
	0x75, 0xb1, 0xb2, 0x05, 0x1b, 0x5b, 0xf0, 0x38, 0xaf, 0x26, 0xf6, 0xef, 0x1f, 0xfd, 0x2e, 0x2f,
	0xa3, 0x79, 0x1a, 0xd7, 0xec, 0x53, 0x96, 0xa5, 0x92, 0x64, 0x5c, 0x56, 0x81, 0xc5, 0xcb, 0xe8,
	0x0d, 0xa9, 0xd0, 0x6d, 0xd8, 0x12, 0x29, 0xb5, 0x0f, 0x5c, 0xe0, 0xdd, 0x08, 0xea, 0xe5, 0x73,
	0xf8, 0xe1, 0xd7, 0xa7, 0x27, 0xaa, 0xb7, 0xc1, 0x17, 0x00, 0xa1, 0x36, 0x60, 0x4c, 0x29, 0xba,
	0x07, 0x2d, 0x41, 0xf2, 0x84, 0x14, 0xcd, 0xd8, 0x87, 0x81, 0x8e, 0xb6, 0x3d, 0xd9, 0xdf, 0xf1,
	0x64, 0xa8, 0x3c, 0x11, 0x76, 0xcb, 0x6d, 0x5d, 0x37, 0xa2, 0x22, 0xd1, 0x31, 0xec, 0x84, 0x94,
	0xce, 0x36, 0xcd, 0x59, 0x21, 0xa5, 0xd3, 0x94, 0xa2, 0xc7, 0xb0, 0x4d, 0x96, 0x24, 0xb6, 0xad,
	0xc6, 0xad, 0xbb, 0x3b, 0xa5, 0x5e, 0x2e, 0x49, 0x1c, 0x34, 0xc8, 0xa0, 0x0b, 0xd1, 0xa6, 0x7b,
	0x73, 0x82, 0xa3, 0x73, 0xd8, 0x3a, 0x13, 0x14, 0x4d, 0x61, 0xc7, 0xcc, 0xd5, 0xc7, 0xff, 0xb9,
	0xda, 0x78, 0x93, 0xda, 0x7b, 0x74, 0x0d, 0x60, 0x6a, 0x4f, 0x5e, 0x7c, 0x5d, 0x39, 0xe0, 0x6a,
	0xe5, 0x80, 0x9f, 0x2b, 0x07, 0x7c, 0x5c, 0x3b, 0x7b, 0x57, 0x6b, 0x67, 0xef, 0xfb, 0xda, 0xd9,
	0x3b, 0x7f, 0x48, 0x53, 0x79, 0x59, 0x46, 0x38, 0x66, 0x99, 0x6f, 0x9e, 0x59, 0xf3, 0x4d, 0xfc,
	0xe5, 0xbf, 0xf7, 0x16, 0x59, 0xcd, 0xf1, 0x9d, 0xfe, 0x0d, 0x00, 0x00, 0xff, 0xff, 0xeb, 0x88,
	0x8d, 0x14, 0x8a, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// VoteAgg allows a sender to submit a set of votes
	VoteAgg(ctx context.Context, in *MsgVoteAgg, opts ...grpc.CallOption) (*MsgVoteAggResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) VoteAgg(ctx context.Context, in *MsgVoteAgg, opts ...grpc.CallOption) (*MsgVoteAggResponse, error) {
	out := new(MsgVoteAggResponse)
	err := c.cc.Invoke(ctx, "/fetchai.blsgroup.v1.Msg/VoteAgg", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// VoteAgg allows a sender to submit a set of votes
	VoteAgg(context.Context, *MsgVoteAgg) (*MsgVoteAggResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) VoteAgg(ctx context.Context, req *MsgVoteAgg) (*MsgVoteAggResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VoteAgg not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_VoteAgg_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgVoteAgg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).VoteAgg(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fetchai.blsgroup.v1.Msg/VoteAgg",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).VoteAgg(ctx, req.(*MsgVoteAgg))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fetchai.blsgroup.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VoteAgg",
			Handler:    _Msg_VoteAgg_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fetchai/blsgroup/v1/tx.proto",
}

func (m *MsgVoteResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgVoteResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgVoteResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Sig) > 0 {
		i -= len(m.Sig)
		copy(dAtA[i:], m.Sig)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sig)))
		i--
		dAtA[i] = 0x2a
	}
	if m.PubKey != nil {
		{
			size, err := m.PubKey.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintTx(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x22
	}
	if m.Option != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Option))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Voter) > 0 {
		i -= len(m.Voter)
		copy(dAtA[i:], m.Voter)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Voter)))
		i--
		dAtA[i] = 0x12
	}
	if m.ProposalId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.ProposalId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *MsgVoteAgg) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgVoteAgg) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgVoteAgg) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Exec != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.Exec))
		i--
		dAtA[i] = 0x30
	}
	if len(m.AggSig) > 0 {
		i -= len(m.AggSig)
		copy(dAtA[i:], m.AggSig)
		i = encodeVarintTx(dAtA, i, uint64(len(m.AggSig)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.Votes) > 0 {
		dAtA3 := make([]byte, len(m.Votes)*10)
		var j2 int
		for _, num := range m.Votes {
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		i -= j2
		copy(dAtA[i:], dAtA3[:j2])
		i = encodeVarintTx(dAtA, i, uint64(j2))
		i--
		dAtA[i] = 0x1a
	}
	if m.ProposalId != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.ProposalId))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgVoteAggResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgVoteAggResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgVoteAggResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgVoteResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ProposalId != 0 {
		n += 1 + sovTx(uint64(m.ProposalId))
	}
	l = len(m.Voter)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Option != 0 {
		n += 1 + sovTx(uint64(m.Option))
	}
	if m.PubKey != nil {
		l = m.PubKey.Size()
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.Sig)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *MsgVoteAgg) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.ProposalId != 0 {
		n += 1 + sovTx(uint64(m.ProposalId))
	}
	if len(m.Votes) > 0 {
		l = 0
		for _, e := range m.Votes {
			l += sovTx(uint64(e))
		}
		n += 1 + sovTx(uint64(l)) + l
	}
	l = len(m.AggSig)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.Exec != 0 {
		n += 1 + sovTx(uint64(m.Exec))
	}
	return n
}

func (m *MsgVoteAggResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgVoteResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgVoteResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgVoteResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
			}
			m.ProposalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProposalId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Voter", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Voter = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Option", wireType)
			}
			m.Option = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Option |= group.VoteOption(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PubKey", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.PubKey == nil {
				m.PubKey = &types.Any{}
			}
			if err := m.PubKey.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sig", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sig = append(m.Sig[:0], dAtA[iNdEx:postIndex]...)
			if m.Sig == nil {
				m.Sig = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgVoteAgg) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgVoteAgg: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgVoteAgg: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
			}
			m.ProposalId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ProposalId |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType == 0 {
				var v group.VoteOption
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTx
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= group.VoteOption(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Votes = append(m.Votes, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTx
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthTx
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthTx
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Votes) == 0 {
					m.Votes = make([]group.VoteOption, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v group.VoteOption
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowTx
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= group.VoteOption(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Votes = append(m.Votes, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Votes", wireType)
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AggSig", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AggSig = append(m.AggSig[:0], dAtA[iNdEx:postIndex]...)
			if m.AggSig == nil {
				m.AggSig = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Exec", wireType)
			}
			m.Exec = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Exec |= group.Exec(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgVoteAggResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgVoteAggResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgVoteAggResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
