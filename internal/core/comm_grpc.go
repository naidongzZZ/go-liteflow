package core

import (
	"context"
	pb "go-liteflow/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// common grpc implement
type Comm struct {
	pb.UnimplementedCoreServer
}

func (c *Comm) EventChannel(srv pb.Core_EventChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method EventChannel not implemented")
}

func (c *Comm) SendHeartBeat(ctx context.Context, req *pb.HeartBeatReq) (resp *pb.HeartBeatResp, err error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendHeartBeat not implemented")
}
