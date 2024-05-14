package coordinator

import (
	"context"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	core.Comm
	coord *coordinator
}

func (s *grpcServer) EventChannel(srv pb.Core_EventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method EventChannel not implemented")
}

func (c *grpcServer) SendHeartBeat(ctx context.Context, req *pb.HeartBeatReq) (resp *pb.HeartBeatResp, err error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendHeartBeat not implemented")
}
