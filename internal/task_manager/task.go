package task_manager

import (
	"context"
	"go-liteflow/internal/pkg"
	pb "go-liteflow/pb"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeployOpTask 部署任务
func (tm *taskManager) DeployOpTask(ctx context.Context, req *pb.DeployOpTaskReq) (resp *pb.DeployOpTaskResp, err error) {
	resp = new(pb.DeployOpTaskResp)

	tm.digraphMux.Lock()
	defer tm.digraphMux.Unlock()

	opTaskIds := make([]string, 0)
	for i := range req.Digraph.Adj {
		// validate opTask
		opTask := req.Digraph.Adj[i]

		if e := pkg.ValidateOpTask(opTask); e != nil {
			slog.Warn("validate optask.", slog.Any("err", e))
			return resp, status.Error(codes.InvalidArgument, "")
		}

		// TODO download plugin.so

		slog.Debug("Deploy Optask:%s to TaskManager:%s", opTask.Id, tm.ID())
		tm.tasks[opTask.Id] = req.Digraph.Adj[i]
		opTaskIds = append(opTaskIds, opTask.Id)
	}


	return resp, nil
}