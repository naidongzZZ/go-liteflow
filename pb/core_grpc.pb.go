// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: pb/core.proto

package __

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

// CoreClient is the client API for Core service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CoreClient interface {
	// Event channels between task managers
	EventChannel(ctx context.Context, opts ...grpc.CallOption) (Core_EventChannelClient, error)
	// Deprecated: dont use
	DirectedEventChannel(ctx context.Context, opts ...grpc.CallOption) (Core_DirectedEventChannelClient, error)
	// TODO raft
	// Send heart beat to coordinator or Ask if the coordinator is alive
	SendHeartBeat(ctx context.Context, in *HeartBeatReq, opts ...grpc.CallOption) (*HeartBeatResp, error)
	// Submit tasks to the coordinator
	SubmitOpTask(ctx context.Context, in *SubmitOpTaskReq, opts ...grpc.CallOption) (*SubmitOpTaskResp, error)
	// Deploy tasks to task manager
	DeployOpTask(ctx context.Context, in *DeployOpTaskReq, opts ...grpc.CallOption) (*DeployOpTaskResp, error)
	// Manage tasks status
	ManageOpTask(ctx context.Context, in *ManageOpTaskReq, opts ...grpc.CallOption) (*ManageOpTaskResp, error)
	// Report task's infomation to coordinator
	ReportOpTask(ctx context.Context, in *ReportOpTaskReq, opts ...grpc.CallOption) (*ReportOpTaskResp, error)
	// Download executable file
	DownloadOpTaskEF(ctx context.Context, in *DownloadReq, opts ...grpc.CallOption) (Core_DownloadOpTaskEFClient, error)
	// Find task info
	FindOpTask(ctx context.Context, in *FindOpTaskReq, opts ...grpc.CallOption) (*FindOpTaskResp, error)
}

type coreClient struct {
	cc grpc.ClientConnInterface
}

func NewCoreClient(cc grpc.ClientConnInterface) CoreClient {
	return &coreClient{cc}
}

func (c *coreClient) EventChannel(ctx context.Context, opts ...grpc.CallOption) (Core_EventChannelClient, error) {
	stream, err := c.cc.NewStream(ctx, &Core_ServiceDesc.Streams[0], "/pb.core/EventChannel", opts...)
	if err != nil {
		return nil, err
	}
	x := &coreEventChannelClient{stream}
	return x, nil
}

type Core_EventChannelClient interface {
	Send(*Event) error
	Recv() (*Event, error)
	grpc.ClientStream
}

type coreEventChannelClient struct {
	grpc.ClientStream
}

func (x *coreEventChannelClient) Send(m *Event) error {
	return x.ClientStream.SendMsg(m)
}

func (x *coreEventChannelClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *coreClient) DirectedEventChannel(ctx context.Context, opts ...grpc.CallOption) (Core_DirectedEventChannelClient, error) {
	stream, err := c.cc.NewStream(ctx, &Core_ServiceDesc.Streams[1], "/pb.core/DirectedEventChannel", opts...)
	if err != nil {
		return nil, err
	}
	x := &coreDirectedEventChannelClient{stream}
	return x, nil
}

type Core_DirectedEventChannelClient interface {
	Send(*EventChannelReq) error
	CloseAndRecv() (*EventChannelResp, error)
	grpc.ClientStream
}

type coreDirectedEventChannelClient struct {
	grpc.ClientStream
}

func (x *coreDirectedEventChannelClient) Send(m *EventChannelReq) error {
	return x.ClientStream.SendMsg(m)
}

func (x *coreDirectedEventChannelClient) CloseAndRecv() (*EventChannelResp, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(EventChannelResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *coreClient) SendHeartBeat(ctx context.Context, in *HeartBeatReq, opts ...grpc.CallOption) (*HeartBeatResp, error) {
	out := new(HeartBeatResp)
	err := c.cc.Invoke(ctx, "/pb.core/SendHeartBeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) SubmitOpTask(ctx context.Context, in *SubmitOpTaskReq, opts ...grpc.CallOption) (*SubmitOpTaskResp, error) {
	out := new(SubmitOpTaskResp)
	err := c.cc.Invoke(ctx, "/pb.core/SubmitOpTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) DeployOpTask(ctx context.Context, in *DeployOpTaskReq, opts ...grpc.CallOption) (*DeployOpTaskResp, error) {
	out := new(DeployOpTaskResp)
	err := c.cc.Invoke(ctx, "/pb.core/DeployOpTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) ManageOpTask(ctx context.Context, in *ManageOpTaskReq, opts ...grpc.CallOption) (*ManageOpTaskResp, error) {
	out := new(ManageOpTaskResp)
	err := c.cc.Invoke(ctx, "/pb.core/ManageOpTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) ReportOpTask(ctx context.Context, in *ReportOpTaskReq, opts ...grpc.CallOption) (*ReportOpTaskResp, error) {
	out := new(ReportOpTaskResp)
	err := c.cc.Invoke(ctx, "/pb.core/ReportOpTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) DownloadOpTaskEF(ctx context.Context, in *DownloadReq, opts ...grpc.CallOption) (Core_DownloadOpTaskEFClient, error) {
	stream, err := c.cc.NewStream(ctx, &Core_ServiceDesc.Streams[2], "/pb.core/DownloadOpTaskEF", opts...)
	if err != nil {
		return nil, err
	}
	x := &coreDownloadOpTaskEFClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Core_DownloadOpTaskEFClient interface {
	Recv() (*DownloadResp, error)
	grpc.ClientStream
}

type coreDownloadOpTaskEFClient struct {
	grpc.ClientStream
}

func (x *coreDownloadOpTaskEFClient) Recv() (*DownloadResp, error) {
	m := new(DownloadResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *coreClient) FindOpTask(ctx context.Context, in *FindOpTaskReq, opts ...grpc.CallOption) (*FindOpTaskResp, error) {
	out := new(FindOpTaskResp)
	err := c.cc.Invoke(ctx, "/pb.core/FindOpTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CoreServer is the server API for Core service.
// All implementations must embed UnimplementedCoreServer
// for forward compatibility
type CoreServer interface {
	// Event channels between task managers
	EventChannel(Core_EventChannelServer) error
	// Deprecated: dont use
	DirectedEventChannel(Core_DirectedEventChannelServer) error
	// TODO raft
	// Send heart beat to coordinator or Ask if the coordinator is alive
	SendHeartBeat(context.Context, *HeartBeatReq) (*HeartBeatResp, error)
	// Submit tasks to the coordinator
	SubmitOpTask(context.Context, *SubmitOpTaskReq) (*SubmitOpTaskResp, error)
	// Deploy tasks to task manager
	DeployOpTask(context.Context, *DeployOpTaskReq) (*DeployOpTaskResp, error)
	// Manage tasks status
	ManageOpTask(context.Context, *ManageOpTaskReq) (*ManageOpTaskResp, error)
	// Report task's infomation to coordinator
	ReportOpTask(context.Context, *ReportOpTaskReq) (*ReportOpTaskResp, error)
	// Download executable file
	DownloadOpTaskEF(*DownloadReq, Core_DownloadOpTaskEFServer) error
	// Find task info
	FindOpTask(context.Context, *FindOpTaskReq) (*FindOpTaskResp, error)
	mustEmbedUnimplementedCoreServer()
}

// UnimplementedCoreServer must be embedded to have forward compatible implementations.
type UnimplementedCoreServer struct {
}

func (UnimplementedCoreServer) EventChannel(Core_EventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method EventChannel not implemented")
}
func (UnimplementedCoreServer) DirectedEventChannel(Core_DirectedEventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method DirectedEventChannel not implemented")
}
func (UnimplementedCoreServer) SendHeartBeat(context.Context, *HeartBeatReq) (*HeartBeatResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendHeartBeat not implemented")
}
func (UnimplementedCoreServer) SubmitOpTask(context.Context, *SubmitOpTaskReq) (*SubmitOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitOpTask not implemented")
}
func (UnimplementedCoreServer) DeployOpTask(context.Context, *DeployOpTaskReq) (*DeployOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeployOpTask not implemented")
}
func (UnimplementedCoreServer) ManageOpTask(context.Context, *ManageOpTaskReq) (*ManageOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ManageOpTask not implemented")
}
func (UnimplementedCoreServer) ReportOpTask(context.Context, *ReportOpTaskReq) (*ReportOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportOpTask not implemented")
}
func (UnimplementedCoreServer) DownloadOpTaskEF(*DownloadReq, Core_DownloadOpTaskEFServer) error {
	return status.Errorf(codes.Unimplemented, "method DownloadOpTaskEF not implemented")
}
func (UnimplementedCoreServer) FindOpTask(context.Context, *FindOpTaskReq) (*FindOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindOpTask not implemented")
}
func (UnimplementedCoreServer) mustEmbedUnimplementedCoreServer() {}

// UnsafeCoreServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CoreServer will
// result in compilation errors.
type UnsafeCoreServer interface {
	mustEmbedUnimplementedCoreServer()
}

func RegisterCoreServer(s grpc.ServiceRegistrar, srv CoreServer) {
	s.RegisterService(&Core_ServiceDesc, srv)
}

func _Core_EventChannel_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CoreServer).EventChannel(&coreEventChannelServer{stream})
}

type Core_EventChannelServer interface {
	Send(*Event) error
	Recv() (*Event, error)
	grpc.ServerStream
}

type coreEventChannelServer struct {
	grpc.ServerStream
}

func (x *coreEventChannelServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func (x *coreEventChannelServer) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Core_DirectedEventChannel_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CoreServer).DirectedEventChannel(&coreDirectedEventChannelServer{stream})
}

type Core_DirectedEventChannelServer interface {
	SendAndClose(*EventChannelResp) error
	Recv() (*EventChannelReq, error)
	grpc.ServerStream
}

type coreDirectedEventChannelServer struct {
	grpc.ServerStream
}

func (x *coreDirectedEventChannelServer) SendAndClose(m *EventChannelResp) error {
	return x.ServerStream.SendMsg(m)
}

func (x *coreDirectedEventChannelServer) Recv() (*EventChannelReq, error) {
	m := new(EventChannelReq)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Core_SendHeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartBeatReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).SendHeartBeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.core/SendHeartBeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).SendHeartBeat(ctx, req.(*HeartBeatReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_SubmitOpTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitOpTaskReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).SubmitOpTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.core/SubmitOpTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).SubmitOpTask(ctx, req.(*SubmitOpTaskReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_DeployOpTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeployOpTaskReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).DeployOpTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.core/DeployOpTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).DeployOpTask(ctx, req.(*DeployOpTaskReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_ManageOpTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ManageOpTaskReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).ManageOpTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.core/ManageOpTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).ManageOpTask(ctx, req.(*ManageOpTaskReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_ReportOpTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReportOpTaskReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).ReportOpTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.core/ReportOpTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).ReportOpTask(ctx, req.(*ReportOpTaskReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_DownloadOpTaskEF_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DownloadReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CoreServer).DownloadOpTaskEF(m, &coreDownloadOpTaskEFServer{stream})
}

type Core_DownloadOpTaskEFServer interface {
	Send(*DownloadResp) error
	grpc.ServerStream
}

type coreDownloadOpTaskEFServer struct {
	grpc.ServerStream
}

func (x *coreDownloadOpTaskEFServer) Send(m *DownloadResp) error {
	return x.ServerStream.SendMsg(m)
}

func _Core_FindOpTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindOpTaskReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).FindOpTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.core/FindOpTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).FindOpTask(ctx, req.(*FindOpTaskReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Core_ServiceDesc is the grpc.ServiceDesc for Core service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Core_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.core",
	HandlerType: (*CoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendHeartBeat",
			Handler:    _Core_SendHeartBeat_Handler,
		},
		{
			MethodName: "SubmitOpTask",
			Handler:    _Core_SubmitOpTask_Handler,
		},
		{
			MethodName: "DeployOpTask",
			Handler:    _Core_DeployOpTask_Handler,
		},
		{
			MethodName: "ManageOpTask",
			Handler:    _Core_ManageOpTask_Handler,
		},
		{
			MethodName: "ReportOpTask",
			Handler:    _Core_ReportOpTask_Handler,
		},
		{
			MethodName: "FindOpTask",
			Handler:    _Core_FindOpTask_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EventChannel",
			Handler:       _Core_EventChannel_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "DirectedEventChannel",
			Handler:       _Core_DirectedEventChannel_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "DownloadOpTaskEF",
			Handler:       _Core_DownloadOpTaskEF_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pb/core.proto",
}
