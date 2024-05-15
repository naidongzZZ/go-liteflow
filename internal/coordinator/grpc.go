package coordinator

import (
	"context"
	"errors"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"log/slog"
)

type grpcServer struct {
	core.Comm
	coord *coordinator
}

func NewGrpcServer() *grpcServer {
	return &grpcServer{}
}

func (c *grpcServer) SendHeartBeat(ctx context.Context, req *pb.HeartBeatReq) (resp *pb.HeartBeatResp, err error) {
	resp = new(pb.HeartBeatResp)
	if req == nil || req.ServiceInfo == nil {
		return resp, errors.New("args is nil")
	}
	slog.Debug("Recv heart beat.", slog.Any("req", req))
	err = c.coord.RegistServiceInfo(req.ServiceInfo)
	if err != nil {
		slog.Error("register service info.", slog.Any("err", err))
		return
	}
	resp.ServiceInfos = c.coord.GetServiceInfo()
	return resp, nil
}
