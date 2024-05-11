package coordinator

import (
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcServer struct {
	core.Comm
}

func (s *GrpcServer) EventChannel(srv pb.Core_EventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method EventChannel not implemented")
}
