package coordinator

import (
	"context"
	"go-liteflow/internal/pkg"
	pb "go-liteflow/pb"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Coordinator 接收提交任务请求, 生成任务图
func (co *coordinator) SubmitOpTask(ctx context.Context, req *pb.SubmitOpTaskReq) (resp *pb.SubmitOpTaskResp, err error) {
	resp = new(pb.SubmitOpTaskResp)

	slog.Debug("Recv submit op task.", slog.Any("req", req))

	

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
