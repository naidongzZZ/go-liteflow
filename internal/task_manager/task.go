package task_manager

import (
	"context"
	"go-liteflow/internal/pkg/log"
	"go-liteflow/internal/pkg/stream"
	pb "go-liteflow/pb"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeployOpTask 部署任务
func (tm *taskManager) DeployOpTask(ctx context.Context, req *pb.DeployOpTaskReq) (resp *pb.DeployOpTaskResp, err error) {
	resp = new(pb.DeployOpTaskResp)

	task := req.Digraph.Adj[0]

	program := task.OpType.String() + "_" + task.Id + "_" + req.EfHash

	err = tm.storager.Write(ctx, req.Ef, program)
	if err != nil {
		log.Errorf("write ef failed: %v", err)
		return resp, status.Error(codes.Internal, "")
	}

	if len(req.Digraph.Adj) != 1 {
		log.Errorf("invalid digraph: %+v", req.Digraph)
		return resp, status.Error(codes.InvalidArgument, "")
	}

	log.Debugf("Deploy Optask:%s to TaskManager:%s", task.Id, tm.ID())
	log.Debugf("Task: %+v", task)

	tm.tasks[task.Id] = req.Digraph
	task.State = pb.TaskStatus_Running

	fpath := tm.storager.GetExecFilePath(program)
	if fpath == "" {
		log.Errorf("exec file not found: %s", req.EfHash)
		return
	}
	log.Infof("tm exec file: %s, task: %s, optask: %s", fpath, task.Id, task.OpType)

	tmids := strings.Join(
		stream.Map(task.Downstream, func(t *pb.OperatorTask) string { return t.TaskManagerId }),
		",",
	)
	tids := strings.Join(
		stream.Map(task.Downstream, func(t *pb.OperatorTask) string { return t.Id }),
		",",
	)

	// 执行任务
	GracefulRun(func(c context.Context) (error) {
		cmd := exec.CommandContext(c, fpath,
			"-id", task.Id,
			"-op", task.OpType.String(),
			"-otmid", tmids,
			"-otid", tids,
			"-tmid", tm.ID())
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		cmd.Cancel = func() error {
			return cmd.Process.Signal(syscall.SIGTERM)
		}
		return cmd.Run()
	})

	return resp, nil
}
