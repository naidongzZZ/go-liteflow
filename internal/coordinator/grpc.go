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

func (c *grpcServer) SubmitOpTask(ctx context.Context, req *pb.SubmitOpTaskReq) (resp *pb.SubmitOpTaskResp, err error) {
	resp = new(pb.SubmitOpTaskResp)

	slog.Debug("Recv submit op task.", slog.Any("req", req))

	err = c.coord.SubmitOpTask(req.Digraph)
	if err != nil {
		slog.Error("Recv op task.", slog.Any("err", err))
	}

	return resp, nil
}
