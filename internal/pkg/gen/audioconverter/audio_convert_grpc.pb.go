// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: protos/audio_convert.proto

package audioconverter

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AudioConverter_Convert_FullMethodName       = "/audio_convert.v1.AudioConverter/Convert"
	AudioConverter_ConvertStream_FullMethodName = "/audio_convert.v1.AudioConverter/ConvertStream"
)

// AudioConverterClient is the client API for AudioConverter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AudioConverterClient interface {
	Convert(ctx context.Context, in *ConvertInput, opts ...grpc.CallOption) (*ConvertReply, error)
	ConvertStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[StreamConvertInput, StreamFileReply], error)
}

type audioConverterClient struct {
	cc grpc.ClientConnInterface
}

func NewAudioConverterClient(cc grpc.ClientConnInterface) AudioConverterClient {
	return &audioConverterClient{cc}
}

func (c *audioConverterClient) Convert(ctx context.Context, in *ConvertInput, opts ...grpc.CallOption) (*ConvertReply, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ConvertReply)
	err := c.cc.Invoke(ctx, AudioConverter_Convert_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *audioConverterClient) ConvertStream(ctx context.Context, opts ...grpc.CallOption) (grpc.BidiStreamingClient[StreamConvertInput, StreamFileReply], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &AudioConverter_ServiceDesc.Streams[0], AudioConverter_ConvertStream_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[StreamConvertInput, StreamFileReply]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type AudioConverter_ConvertStreamClient = grpc.BidiStreamingClient[StreamConvertInput, StreamFileReply]

// AudioConverterServer is the server API for AudioConverter service.
// All implementations should embed UnimplementedAudioConverterServer
// for forward compatibility.
type AudioConverterServer interface {
	Convert(context.Context, *ConvertInput) (*ConvertReply, error)
	ConvertStream(grpc.BidiStreamingServer[StreamConvertInput, StreamFileReply]) error
}

// UnimplementedAudioConverterServer should be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAudioConverterServer struct{}

func (UnimplementedAudioConverterServer) Convert(context.Context, *ConvertInput) (*ConvertReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Convert not implemented")
}
func (UnimplementedAudioConverterServer) ConvertStream(grpc.BidiStreamingServer[StreamConvertInput, StreamFileReply]) error {
	return status.Errorf(codes.Unimplemented, "method ConvertStream not implemented")
}
func (UnimplementedAudioConverterServer) testEmbeddedByValue() {}

// UnsafeAudioConverterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AudioConverterServer will
// result in compilation errors.
type UnsafeAudioConverterServer interface {
	mustEmbedUnimplementedAudioConverterServer()
}

func RegisterAudioConverterServer(s grpc.ServiceRegistrar, srv AudioConverterServer) {
	// If the following call pancis, it indicates UnimplementedAudioConverterServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AudioConverter_ServiceDesc, srv)
}

func _AudioConverter_Convert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConvertInput)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AudioConverterServer).Convert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AudioConverter_Convert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AudioConverterServer).Convert(ctx, req.(*ConvertInput))
	}
	return interceptor(ctx, in, info, handler)
}

func _AudioConverter_ConvertStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AudioConverterServer).ConvertStream(&grpc.GenericServerStream[StreamConvertInput, StreamFileReply]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type AudioConverter_ConvertStreamServer = grpc.BidiStreamingServer[StreamConvertInput, StreamFileReply]

// AudioConverter_ServiceDesc is the grpc.ServiceDesc for AudioConverter service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AudioConverter_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "audio_convert.v1.AudioConverter",
	HandlerType: (*AudioConverterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Convert",
			Handler:    _AudioConverter_Convert_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ConvertStream",
			Handler:       _AudioConverter_ConvertStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "protos/audio_convert.proto",
}
