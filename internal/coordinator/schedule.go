package coordinator

import (
	"context"
	pb "go-liteflow/pb"
	"log/slog"
	"time"
)

func (co *coordinator) schedule(ctx context.Context) {

	// TODO downstream info, task_maanager_id
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := co.deployPendingTasks(ctx); err != nil {
					slog.Error("部署任务失败", slog.Any("err", err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// 获取空闲可用的task manager
func (co *coordinator) getIdleTaskManager() (taskManangerId string, client pb.CoreClient) {
	co.mux.Lock()
	defer co.mux.Unlock()

	var info *pb.ServiceInfo
	for _, si := range co.serviceInfos {
		if si.ServiceStatus == pb.ServiceStatus_SsRunning {
			info = si
			break
		}
	}
	if info == nil {
		return
	}
	taskManangerId = info.Id

	obj, ok := co.clientConns[info.ServiceAddr]
	if !ok {
		return
	}
	return taskManangerId, obj
}

func (co *coordinator) deployPendingTasks(ctx context.Context) error {
	// 获取空闲的 task manager
	taskManagerId, _ := co.getIdleTaskManager()
	if taskManagerId == "" {
		return nil // 没有可用的 task manager，暂时跳过
	}

	// TODO: 实现具体的任务部署逻辑
	// 1. 获取待部署的任务
	// 2. 通过 client 调用 task manager 的部署接口

	return nil
}
