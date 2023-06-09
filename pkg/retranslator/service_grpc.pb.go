// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: service.proto

package retranslator

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CommandStreamClient is the client API for CommandStream service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CommandStreamClient interface {
	ListenCommands(ctx context.Context, opts ...grpc.CallOption) (CommandStream_ListenCommandsClient, error)
}

type commandStreamClient struct {
	cc grpc.ClientConnInterface
}

func NewCommandStreamClient(cc grpc.ClientConnInterface) CommandStreamClient {
	return &commandStreamClient{cc}
}

func (c *commandStreamClient) ListenCommands(ctx context.Context, opts ...grpc.CallOption) (CommandStream_ListenCommandsClient, error) {
	stream, err := c.cc.NewStream(ctx, &CommandStream_ServiceDesc.Streams[0], "/retranslator.CommandStream/ListenCommands", opts...)
	if err != nil {
		return nil, err
	}
	x := &commandStreamListenCommandsClient{stream}
	return x, nil
}

type CommandStream_ListenCommandsClient interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ClientStream
}

type commandStreamListenCommandsClient struct {
	grpc.ClientStream
}

func (x *commandStreamListenCommandsClient) Send(m *Response) error {
	return x.ClientStream.SendMsg(m)
}

func (x *commandStreamListenCommandsClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CommandStreamServer is the server API for CommandStream service.
// All implementations must embed UnimplementedCommandStreamServer
// for forward compatibility
type CommandStreamServer interface {
	ListenCommands(CommandStream_ListenCommandsServer) error
	mustEmbedUnimplementedCommandStreamServer()
}

// UnimplementedCommandStreamServer must be embedded to have forward compatible implementations.
type UnimplementedCommandStreamServer struct {
}

func (UnimplementedCommandStreamServer) ListenCommands(CommandStream_ListenCommandsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListenCommands not implemented")
}
func (UnimplementedCommandStreamServer) mustEmbedUnimplementedCommandStreamServer() {}

// UnsafeCommandStreamServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CommandStreamServer will
// result in compilation errors.
type UnsafeCommandStreamServer interface {
	mustEmbedUnimplementedCommandStreamServer()
}

func RegisterCommandStreamServer(s grpc.ServiceRegistrar, srv CommandStreamServer) {
	s.RegisterService(&CommandStream_ServiceDesc, srv)
}

func _CommandStream_ListenCommands_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CommandStreamServer).ListenCommands(&commandStreamListenCommandsServer{stream})
}

type CommandStream_ListenCommandsServer interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ServerStream
}

type commandStreamListenCommandsServer struct {
	grpc.ServerStream
}

func (x *commandStreamListenCommandsServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

func (x *commandStreamListenCommandsServer) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CommandStream_ServiceDesc is the grpc.ServiceDesc for CommandStream service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CommandStream_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "retranslator.CommandStream",
	HandlerType: (*CommandStreamServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListenCommands",
			Handler:       _CommandStream_ListenCommands_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "service.proto",
}
