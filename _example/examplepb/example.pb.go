// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        v4.24.4
// source: example.proto

package examplepb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Empty struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_example_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_example_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_example_proto_rawDescGZIP(), []int{0}
}

type HopRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hops int64 `protobuf:"varint,1,opt,name=hops,proto3" json:"hops,omitempty"`
}

func (x *HopRequest) Reset() {
	*x = HopRequest{}
	mi := &file_example_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HopRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HopRequest) ProtoMessage() {}

func (x *HopRequest) ProtoReflect() protoreflect.Message {
	mi := &file_example_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HopRequest.ProtoReflect.Descriptor instead.
func (*HopRequest) Descriptor() ([]byte, []int) {
	return file_example_proto_rawDescGZIP(), []int{1}
}

func (x *HopRequest) GetHops() int64 {
	if x != nil {
		return x.Hops
	}
	return 0
}

var File_example_proto protoreflect.FileDescriptor

var file_example_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x09, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x22, 0x20, 0x0a, 0x0a, 0x48, 0x6f, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x70, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x04, 0x68, 0x6f, 0x70, 0x73, 0x32, 0x3a, 0x0a, 0x06, 0x48, 0x6f, 0x70, 0x70, 0x65, 0x72, 0x12,
	0x30, 0x0a, 0x03, 0x48, 0x6f, 0x70, 0x12, 0x15, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65,
	0x70, 0x62, 0x2e, 0x48, 0x6f, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e,
	0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x42, 0x0e, 0x5a, 0x0c, 0x2e, 0x2e, 0x2f, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_example_proto_rawDescOnce sync.Once
	file_example_proto_rawDescData = file_example_proto_rawDesc
)

func file_example_proto_rawDescGZIP() []byte {
	file_example_proto_rawDescOnce.Do(func() {
		file_example_proto_rawDescData = protoimpl.X.CompressGZIP(file_example_proto_rawDescData)
	})
	return file_example_proto_rawDescData
}

var file_example_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_example_proto_goTypes = []any{
	(*Empty)(nil),      // 0: examplepb.Empty
	(*HopRequest)(nil), // 1: examplepb.HopRequest
}
var file_example_proto_depIdxs = []int32{
	1, // 0: examplepb.Hopper.Hop:input_type -> examplepb.HopRequest
	0, // 1: examplepb.Hopper.Hop:output_type -> examplepb.Empty
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_example_proto_init() }
func file_example_proto_init() {
	if File_example_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_example_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_example_proto_goTypes,
		DependencyIndexes: file_example_proto_depIdxs,
		MessageInfos:      file_example_proto_msgTypes,
	}.Build()
	File_example_proto = out.File
	file_example_proto_rawDesc = nil
	file_example_proto_goTypes = nil
	file_example_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// HopperClient is the client API for Hopper service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HopperClient interface {
	Hop(ctx context.Context, in *HopRequest, opts ...grpc.CallOption) (*Empty, error)
}

type hopperClient struct {
	cc grpc.ClientConnInterface
}

func NewHopperClient(cc grpc.ClientConnInterface) HopperClient {
	return &hopperClient{cc}
}

func (c *hopperClient) Hop(ctx context.Context, in *HopRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := c.cc.Invoke(ctx, "/examplepb.Hopper/Hop", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HopperServer is the server API for Hopper service.
type HopperServer interface {
	Hop(context.Context, *HopRequest) (*Empty, error)
}

// UnimplementedHopperServer can be embedded to have forward compatible implementations.
type UnimplementedHopperServer struct {
}

func (*UnimplementedHopperServer) Hop(context.Context, *HopRequest) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hop not implemented")
}

func RegisterHopperServer(s *grpc.Server, srv HopperServer) {
	s.RegisterService(&_Hopper_serviceDesc, srv)
}

func _Hopper_Hop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HopperServer).Hop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/examplepb.Hopper/Hop",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HopperServer).Hop(ctx, req.(*HopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Hopper_serviceDesc = grpc.ServiceDesc{
	ServiceName: "examplepb.Hopper",
	HandlerType: (*HopperServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hop",
			Handler:    _Hopper_Hop_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example.proto",
}
