package coordinator

import (
	"context"
	"go-liteflow/internal/pkg"
	"go-liteflow/internal/pkg/log"
	pb "go-liteflow/pb"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Coordinator 接收提交任务请求, 生成任务图
func (co *coordinator) SubmitOpTask(ctx context.Context, req *pb.SubmitOpTaskReq) (resp *pb.SubmitOpTaskResp, err error) {
	resp = new(pb.SubmitOpTaskResp)

	log.Debugf("Recv submit op task. req: %v", req.Digraph.EfHash)
	if uuid.Validate(req.ClientId) != nil {
		return resp, status.Errorf(codes.InvalidArgument, "client id is invalid")
	}

	// 校验任务拓扑
	if err := pkg.ValidateDigraph(req.Digraph); err != nil {
		slog.Warn("validate digraph failed", slog.Any("err", err))
		return resp, status.Error(codes.InvalidArgument, "")
	}

	// 将可执行文件写入本地
	err = co.storager.Write(ctx, req.Ef, req.Digraph.EfHash)
	if err != nil {
		slog.Error("Write executable file failed", slog.Any("err", err))
		return resp, status.Errorf(codes.Internal, "write executable file failed")
	}

	// TODO 先简单处理

	co.digraphMux.Lock()
	defer co.digraphMux.Unlock()

	// 将任务拓扑写入缓存
	co.taskDigraph[req.ClientId] = req.Digraph

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

// 获取任务拓扑
func (co *coordinator) GetDigraph() (digraphs []*pb.Digraph) {
	co.digraphMux.Lock()
	defer co.digraphMux.Unlock()

	for _, digraph := range co.taskDigraph {
		for _, task := range digraph.Adj {
			// 返回存在状态为Ready的任务
			if task.State == pb.TaskStatus_Ready {
				digraphs = append(digraphs, digraph)
				break
			}
		}
	}
	return digraphs
}
