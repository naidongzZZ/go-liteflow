package coordinator

import (
	"context"
	"errors"
	pb "go-liteflow/pb"
	"log/slog"

	"google.golang.org/grpc/codes"
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
	resp.ServiceInfos = co.GetServiceInfo()
	return resp, nil
}

func (co *coordinator) ReportOpTask(ctx context.Context, req *pb.ReportOpTaskReq) (resp *pb.ReportOpTaskResp, err error) {
	resp = new(pb.ReportOpTaskResp)

	co.digraphMux.Lock()
	defer co.digraphMux.Unlock()

	digraph, ok := co.taskDigraph[req.ClientId]
	if !ok {
		return resp, status.Errorf(codes.InvalidArgument, "")
	}

	for _, task := range digraph.Adj {
		if task.Id != req.OpTaskId {
			continue
		}
		task.State = req.OpTaskStatus
		resp.TaskManagerId = task.TaskManagerId
		break
	}

	return
}

func (co *coordinator) FindOpTask(ctx context.Context, req *pb.FindOpTaskReq) (resp *pb.FindOpTaskResp, err error) {

	resp = new(pb.FindOpTaskResp)

	co.digraphMux.Lock()
	defer co.digraphMux.Unlock()

	digraph, ok := co.taskDigraph[req.ClientId]
	if !ok {
		return resp, status.Errorf(codes.InvalidArgument, "")
	}

	for _, t := range digraph.Adj {
		if t.Id == req.OpTaskId {
			resp.TaskManagerId = t.TaskManagerId
			break
		}
	}
	return
}
