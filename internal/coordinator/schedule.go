package coordinator

import (
	"context"
	"errors"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"log/slog"
	"time"
)

func (co *coordinator) schedule(ctx context.Context) {

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := co.deployReadyTasks(ctx); err != nil {
					slog.Error("部署任务失败", slog.Any("err", err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (co *coordinator) deployReadyTasks(ctx context.Context) error {
	// 全局锁定, 防止并发调度
	co.scheduleMux.Lock()
	defer co.scheduleMux.Unlock()

	// 实现具体的任务部署逻辑
	// 1. 获取待部署的任务
	// 2. 通过 client 调用 task manager 的部署接口
	digraphs := co.GetDigraph()
	for _, digraph := range digraphs {

		// 获取空闲的 task manager
		mgs := co.GetIdle(pb.ServiceType_TaskManager, pb.ServiceStatus_SsRunning)
		if len(mgs) == 0 {
			return errors.New("no idle task manager")
		}
		// 轮询分配
		assign(mgs, digraph.Adj)
		// 发射任务
		co.emit(ctx, mgs, digraph)
		// 标记任务状态为Running
		mark(digraph.Adj, pb.TaskStatus_Running)
	}
	return nil
}

func assign(srvs []*core.Service, tasks []*pb.OperatorTask) {

	taskId2tmId := make(map[string]string)
	// 将srv的Id轮询设置到task的TaskManagerId
	for i, task := range tasks {
		task.TaskManagerId = srvs[i%len(srvs)].Id
		taskId2tmId[task.Id] = task.TaskManagerId
	}

	// 补充task的upstream和downstream的TaskManagerId
	for _, task := range tasks {
		for _, upstream := range task.Upstream {
			upstream.TaskManagerId = taskId2tmId[upstream.Id]
		}
		for _, downstream := range task.Downstream {
			downstream.TaskManagerId = taskId2tmId[downstream.Id]
		}
	}
}

func mark(tasks []*pb.OperatorTask, status pb.TaskStatus) {
	for _, task := range tasks {
		task.State = status
	}
}

func (co *coordinator) emit(ctx context.Context, srvs []*core.Service, digraph *pb.Digraph) {

	m := make(map[string]*core.Service)
	for _, srv := range srvs {
		m[srv.Id] = srv
	}

	ef, err := co.storager.Read(ctx, digraph.EfHash)
	if err != nil {
		slog.Error("read ef failed", slog.Any("err", err))
		return
	}

	for i := len(digraph.Adj) - 1; i >= 0; i-- {
		task := digraph.Adj[i]
		srv, ok := m[task.TaskManagerId]
		if !ok {
			slog.Error("not found", slog.Any("task", task))
			return
		}

		// 部署任务
		_, err := srv.ClientConn.DeployOpTask(ctx, &pb.DeployOpTaskReq{
			Ef:      ef,
			EfHash:  digraph.EfHash,
			Digraph: &pb.Digraph{
				GraphId: digraph.GraphId,
				EfHash:  digraph.EfHash,
				Adj:     []*pb.OperatorTask{task},
			}})
		if err != nil {
			slog.Error("deploy task failed", slog.Any("task", task), slog.Any("err", err))
			return
		}
		time.Sleep(1 * time.Second)
	}
}
