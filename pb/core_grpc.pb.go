// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.27.0--rc3
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

const (
	Core_EmitEvent_FullMethodName    = "/pb.core/EmitEvent"
	Core_HeartBeat_FullMethodName    = "/pb.core/HeartBeat"
	Core_SubmitOpTask_FullMethodName = "/pb.core/SubmitOpTask"
	Core_DeployOpTask_FullMethodName = "/pb.core/DeployOpTask"
	Core_ReportOpTask_FullMethodName = "/pb.core/ReportOpTask"
	Core_FindOpTask_FullMethodName   = "/pb.core/FindOpTask"
)

// CoreClient is the client API for Core service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CoreClient interface {
	// Event channels between task managers
	EmitEvent(ctx context.Context, in *Event, opts ...grpc.CallOption) (*EmitEventResp, error)
	// TODO raft
	// Send heart beat to coordinator or Ask if the coordinator is alive
	HeartBeat(ctx context.Context, in *HeartBeatReq, opts ...grpc.CallOption) (*HeartBeatResp, error)
	// Submit tasks to the coordinator
	SubmitOpTask(ctx context.Context, in *SubmitOpTaskReq, opts ...grpc.CallOption) (*SubmitOpTaskResp, error)
	// Deploy tasks to task manager
	DeployOpTask(ctx context.Context, in *DeployOpTaskReq, opts ...grpc.CallOption) (*DeployOpTaskResp, error)
	// Manage tasks status
	// rpc ManageOpTask(ManageOpTaskReq) returns (ManageOpTaskResp) {}
	// Report task's infomation to coordinator
	ReportOpTask(ctx context.Context, in *ReportOpTaskReq, opts ...grpc.CallOption) (*ReportOpTaskResp, error)
	// Download executable file
	// rpc DownloadOpTaskEF(DownloadReq) returns (stream DownloadResp) {}
	// Find task info
	FindOpTask(ctx context.Context, in *FindOpTaskReq, opts ...grpc.CallOption) (*FindOpTaskResp, error)
}

type coreClient struct {
	cc grpc.ClientConnInterface
}

func NewCoreClient(cc grpc.ClientConnInterface) CoreClient {
	return &coreClient{cc}
}

func (c *coreClient) EmitEvent(ctx context.Context, in *Event, opts ...grpc.CallOption) (*EmitEventResp, error) {
	out := new(EmitEventResp)
	err := c.cc.Invoke(ctx, Core_EmitEvent_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) HeartBeat(ctx context.Context, in *HeartBeatReq, opts ...grpc.CallOption) (*HeartBeatResp, error) {
	out := new(HeartBeatResp)
	err := c.cc.Invoke(ctx, Core_HeartBeat_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) SubmitOpTask(ctx context.Context, in *SubmitOpTaskReq, opts ...grpc.CallOption) (*SubmitOpTaskResp, error) {
	out := new(SubmitOpTaskResp)
	err := c.cc.Invoke(ctx, Core_SubmitOpTask_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) DeployOpTask(ctx context.Context, in *DeployOpTaskReq, opts ...grpc.CallOption) (*DeployOpTaskResp, error) {
	out := new(DeployOpTaskResp)
	err := c.cc.Invoke(ctx, Core_DeployOpTask_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) ReportOpTask(ctx context.Context, in *ReportOpTaskReq, opts ...grpc.CallOption) (*ReportOpTaskResp, error) {
	out := new(ReportOpTaskResp)
	err := c.cc.Invoke(ctx, Core_ReportOpTask_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) FindOpTask(ctx context.Context, in *FindOpTaskReq, opts ...grpc.CallOption) (*FindOpTaskResp, error) {
	out := new(FindOpTaskResp)
	err := c.cc.Invoke(ctx, Core_FindOpTask_FullMethodName, in, out, opts...)
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
	EmitEvent(context.Context, *Event) (*EmitEventResp, error)
	// TODO raft
	// Send heart beat to coordinator or Ask if the coordinator is alive
	HeartBeat(context.Context, *HeartBeatReq) (*HeartBeatResp, error)
	// Submit tasks to the coordinator
	SubmitOpTask(context.Context, *SubmitOpTaskReq) (*SubmitOpTaskResp, error)
	// Deploy tasks to task manager
	DeployOpTask(context.Context, *DeployOpTaskReq) (*DeployOpTaskResp, error)
	// Manage tasks status
	// rpc ManageOpTask(ManageOpTaskReq) returns (ManageOpTaskResp) {}
	// Report task's infomation to coordinator
	ReportOpTask(context.Context, *ReportOpTaskReq) (*ReportOpTaskResp, error)
	// Download executable file
	// rpc DownloadOpTaskEF(DownloadReq) returns (stream DownloadResp) {}
	// Find task info
	FindOpTask(context.Context, *FindOpTaskReq) (*FindOpTaskResp, error)
	mustEmbedUnimplementedCoreServer()
}

// UnimplementedCoreServer must be embedded to have forward compatible implementations.
type UnimplementedCoreServer struct {
}

func (UnimplementedCoreServer) EmitEvent(context.Context, *Event) (*EmitEventResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EmitEvent not implemented")
}
func (UnimplementedCoreServer) HeartBeat(context.Context, *HeartBeatReq) (*HeartBeatResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HeartBeat not implemented")
}
func (UnimplementedCoreServer) SubmitOpTask(context.Context, *SubmitOpTaskReq) (*SubmitOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitOpTask not implemented")
}
func (UnimplementedCoreServer) DeployOpTask(context.Context, *DeployOpTaskReq) (*DeployOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeployOpTask not implemented")
}
func (UnimplementedCoreServer) ReportOpTask(context.Context, *ReportOpTaskReq) (*ReportOpTaskResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportOpTask not implemented")
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

func _Core_EmitEvent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Event)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).EmitEvent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Core_EmitEvent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).EmitEvent(ctx, req.(*Event))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_HeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartBeatReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).HeartBeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Core_HeartBeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).HeartBeat(ctx, req.(*HeartBeatReq))
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
		FullMethod: Core_SubmitOpTask_FullMethodName,
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
		FullMethod: Core_DeployOpTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).DeployOpTask(ctx, req.(*DeployOpTaskReq))
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
		FullMethod: Core_ReportOpTask_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).ReportOpTask(ctx, req.(*ReportOpTaskReq))
	}
	return interceptor(ctx, in, info, handler)
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
		FullMethod: Core_FindOpTask_FullMethodName,
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
			MethodName: "EmitEvent",
			Handler:    _Core_EmitEvent_Handler,
		},
		{
			MethodName: "HeartBeat",
			Handler:    _Core_HeartBeat_Handler,
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
			MethodName: "ReportOpTask",
			Handler:    _Core_ReportOpTask_Handler,
		},
		{
			MethodName: "FindOpTask",
			Handler:    _Core_FindOpTask_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/core.proto",
}
