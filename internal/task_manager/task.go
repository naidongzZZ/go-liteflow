package task_manager

import (
	"context"
	pb "go-liteflow/pb"
	"log/slog"
	"os/exec"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeployOpTask 部署任务
func (tm *taskManager) DeployOpTask(ctx context.Context, req *pb.DeployOpTaskReq) (resp *pb.DeployOpTaskResp, err error) {
	resp = new(pb.DeployOpTaskResp)

	err = tm.storager.Write(ctx, req.Ef, req.EfHash)
	if err != nil {
		slog.Error("write ef failed", slog.Any("err", err))
		return resp, status.Error(codes.Internal, "")
	}

	if len(req.Digraph.Adj) != 1 {
		slog.Error("invalid digraph", slog.Any("digraph", req.Digraph))
		return resp, status.Error(codes.InvalidArgument, "")
	}

	task := req.Digraph.Adj[0]
	slog.Debug("Deploy Optask:%s to TaskManager:%s", task.Id, tm.ID())
	slog.Debug("Task: %+v", task)
	
	tm.tasks[task.Id] = req.Digraph
	task.State = pb.TaskStatus_Running

	// 执行任务
	RunnerPool.Run(ctx, func(c context.Context) {

		filepath := tm.storager.GetExecFilePath(req.EfHash)
		if filepath == "" {
			slog.Error("exec file not found", slog.Any("hash", req.EfHash))
			return
		}

		slog.Info("tm exec file", slog.Any("filepath", filepath), slog.Any("task", task.Id), slog.Any("optask", task.OpType))
		cmd := exec.Command(filepath)
		err = cmd.Run()
		if err != nil {
			slog.Error("exec file failed", slog.Any("err", err))
			return
		}

		// TODO 设置输入输出
	})

	return resp, nil
}