package task_manager

import (
	"context"
	pb "go-liteflow/pb"
	"log/slog"
	"time"

	"google.golang.org/grpc"
)

func (tm *taskManager) heartBeat(ctx context.Context) (err error) {
	go func() {
		for {
			time.Sleep(5 * time.Second)

			req := &pb.HeartBeatReq{ServiceInfo: tm.taskManagerInfo}
			resp, err := tm.Coordinator().SendHeartBeat(ctx, req)
			if err != nil {
				slog.Error("send heart beat to coordinator.", slog.Any("err", err))
				continue
			}
			if resp == nil || len(resp.ServiceInfos) == 0 {
				continue
			}

			for tmid, servInfo := range resp.ServiceInfos {
				if tmid == tm.taskManagerInfo.Id {
					continue
				}
				tm.mux.Lock()
				if _, ok := tm.serviceInfos[tmid]; ok {
					// TODO update service info ?
					tm.mux.Unlock()
					continue
				}

				addr := servInfo.ServiceAddr
				conn, err := grpc.Dial(addr, grpc.WithInsecure())
				if err != nil {
					slog.Error("dial task manager", slog.String("addr", addr), slog.Any("err", err))
					tm.mux.Unlock()
					continue
				}

				client := pb.NewCoreClient(conn)
				ch, err := NewChannelWithClient(context.Background(), tmid, client)
				if err != nil {
					slog.Error("new channel.", slog.Any("err", err))
					tm.mux.Unlock()
					continue
				}

				tm.clientConns[addr] = client
				tm.TaskManangerEventChans[tmid] = ch
				tm.serviceInfos[tmid] = servInfo
				tm.mux.Unlock()
			}

			slog.Info("task manager send heart beat.", slog.Int("tm_size", len(resp.ServiceInfos)))
		}
	}()

	return
}
