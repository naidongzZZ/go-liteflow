package task_manager


import (
	pb "go-liteflow/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	pb.UnimplementedCoreServer
}

func (s *GrpcServer) EventChannel(srv pb.Core_EventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method EventChannel not implemented")
}