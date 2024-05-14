package task_manager

import (
	"context"
	pb "go-liteflow/pb"
	"log/slog"
	"net"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type taskManager struct {
	taskManagerId string
	addr          string

	coordinatorId   string
	coordinatorAddr string
	coordinatorConn pb.CoreClient

	srv  *grpc.Server
	gSrv *grpcServer

	taskManagerAddrs map[string]pb.CoreClient
}

func NewTaskManager(addr, coordAddr string) *taskManager {

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	tm := &taskManager{
		taskManagerId:   ranUid.String(),
		addr:            addr,
		coordinatorAddr: coordAddr,
		taskManagerAddrs: make(map[string]pb.CoreClient),
	}

	srv := grpc.NewServer()

	gSrv := &grpcServer{tm: tm}
	pb.RegisterCoreServer(srv, gSrv)

	tm.gSrv = gSrv
	tm.srv = srv

	return tm
}

func (tm *taskManager) Start(ctx context.Context) {

	err := tm.heartBeat(ctx)
	if err != nil {
		slog.Error("task_manager heart beat.", slog.Any("err", err))
		return
	}

	listener, err := net.Listen("tcp", tm.addr)
	if err != nil {
		slog.Error("new listener.", slog.Any("err", err))
		panic(err)
	}

	slog.Info("task_manager grpc server start success.")
	if err = tm.srv.Serve(listener); err != nil {
		slog.Error("grpc serve fail.", slog.Any("err", err))
		panic(err)
	}
}

func (tm *taskManager) heartBeat(ctx context.Context) (err error) {
	conn, err := grpc.Dial(tm.coordinatorAddr, grpc.WithInsecure())
	if err != nil {
		slog.Error("dial coordinator", slog.Any("err", err))
		return err
	}

	tm.coordinatorConn = pb.NewCoreClient(conn)
	
	go func ()  {
		for {
			time.Sleep(5 * time.Second)

			if conn.GetState() != connectivity.Ready {
				slog.Debug("wait for connect coordinator.", slog.Any("state", conn.GetState()))
				continue
			}
			resp, err := tm.coordinatorConn.SendHeartBeat(ctx,
				&pb.HeartBeatReq{TaskManagerId: tm.taskManagerId})
			if err != nil {
				slog.Error("send heart beat to coordinator,", slog.Any("err", err))
				continue
			}
			if resp == nil || len(resp.TaskManagerAddrs) == 0 {
				continue
			}
			for _, tmAddr := range resp.TaskManagerAddrs {
				if _, ok := tm.taskManagerAddrs[tmAddr]; ok {
					continue
				}
				tm.taskManagerAddrs[tmAddr] = nil
			}

			slog.Info("task manager send heart beat", slog.Int("task_manager_size", len(resp.TaskManagerAddrs)))
		}
	}()

	return
}

func (tm *taskManager) ID() string {
	return tm.taskManagerId
}