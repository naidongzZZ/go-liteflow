package coordinator

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	pb "go-liteflow/pb"
	"google.golang.org/grpc/status"
)

// Coordinator 接收心跳, 注册服务信息
func (co *coordinator) SendHeartBeat(ctx context.Context, req *pb.HeartBeatReq) (resp *pb.HeartBeatResp, err error) {
	resp = new(pb.HeartBeatResp)
	if req == nil || req.ServiceInfo == nil {
		return resp, errors.New("args is nil") 
	}
	slog.Debug("Recv heart beat.", slog.Any("req", req))

	// 注册服务信息
	err = co.RegistServiceInfo(req.ServiceInfo)
	if err != nil {
		slog.Error("register service info.", slog.Any("err", err))
		return resp, status.Errorf(codes.InvalidArgument, "")
	}

	// 返回coordinator上注册的服务信息
	resp.ServiceInfos = co.GetServiceInfo()
	return resp, nil
}