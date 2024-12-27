package task_manager

import (
	"context"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"log/slog"
	"time"
)

// 发送心跳到coordinator, 获取其他task manager的信息
func (tm *taskManager) heartBeat(ctx context.Context) (err error) {
	go func() {
		for {
			time.Sleep(5 * time.Second)

			req := &pb.HeartBeatReq{ServiceInfo: &tm.SelfServiceInfo().ServiceInfo}

			resp, err := tm.Coordinator().SendHeartBeat(ctx, req)
			if err != nil {
				slog.Error("send heart beat to coordinator.", slog.Any("err", err))
				continue
			}
			if resp == nil || len(resp.ServiceInfos) == 0 {
				slog.Info("no service info from coordinator.")
				continue
			}

			for tmid, servInfo := range resp.ServiceInfos {
				if tmid == tm.Id {
					continue
				}

				tm.RegisterServiceInfo(servInfo)
			}

			slog.Info("task manager send heart beat.", slog.Int("tm_size", len(resp.ServiceInfos)))
		}
	}()

	return
}

func (tm *taskManager) RegisterServiceInfo(servInfo *pb.ServiceInfo) {
	tm.mux.Lock()
	defer tm.mux.Unlock()

	info, ok := tm.serviceInfos[servInfo.Id]
	if !ok {
		coreCli, err := core.NewCoreClient(servInfo.ServiceAddr)
		if err != nil {
			slog.Error("new core client.", slog.Any("err", err))
			return
		}

		tm.serviceInfos[servInfo.Id] = &core.Service{
			ServiceInfo: pb.ServiceInfo{
				Id:            servInfo.Id,
				ServiceAddr:   servInfo.ServiceAddr,
				ServiceType:   servInfo.ServiceType,
				ServiceStatus: servInfo.ServiceStatus,
				Timestamp:     servInfo.Timestamp,
			},
			ClientConn: coreCli,
		}
	}
	info.ServiceStatus = servInfo.ServiceStatus
	info.Timestamp = servInfo.Timestamp
}
