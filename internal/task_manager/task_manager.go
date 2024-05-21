package task_manager

import (
	"context"
	"go-liteflow/internal/pkg/operator"
	pb "go-liteflow/pb"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type taskManager struct {
	taskManagerInfo *pb.ServiceInfo

	coordinatorInfo *pb.ServiceInfo

	srv  *grpc.Server
	gSrv *grpcServer

	mux          sync.Mutex
	serviceInfos map[string]*pb.ServiceInfo
	clientConns  map[string]pb.CoreClient
	TaskManagerBufferMonitor
}

func NewTaskManager(addr, coordAddr string) *taskManager {

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	tm := &taskManager{
		taskManagerInfo: &pb.ServiceInfo{
			Id:          ranUid.String(),
			ServiceAddr: addr,
		},
		coordinatorInfo: &pb.ServiceInfo{
			ServiceAddr: coordAddr,
		},
		serviceInfos: make(map[string]*pb.ServiceInfo),
		clientConns:  make(map[string]pb.CoreClient),
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

	listener, err := net.Listen("tcp", tm.taskManagerInfo.ServiceAddr)
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
	conn, err := grpc.Dial(tm.coordinatorInfo.ServiceAddr, grpc.WithInsecure())
	if err != nil {
		slog.Error("dial coordinator", slog.Any("err", err))
		return err
	}

	coordinatorConn := pb.NewCoreClient(conn)
	tm.mux.Lock()
	tm.clientConns[tm.taskManagerInfo.ServiceAddr] = coordinatorConn
	tm.mux.Unlock()

	go func() {
		for {
			time.Sleep(5 * time.Second)

			resp, err := coordinatorConn.SendHeartBeat(ctx,
				&pb.HeartBeatReq{
					ServiceInfo: &pb.ServiceInfo{
						Id:            tm.taskManagerInfo.Id,
						ServiceType:   pb.ServiceType_TaskManager,
						ServiceStatus: tm.taskManagerInfo.ServiceStatus,
						ServiceAddr:   tm.taskManagerInfo.ServiceAddr}})
			if err != nil {
				slog.Error("send heart beat to coordinator,",
					slog.Any("err", err))
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
					tm.mux.Unlock()
					continue
				}
				tm.serviceInfos[tmid] = servInfo
				tm.mux.Unlock()
			}

			slog.Info("task manager send heart beat.",
				slog.Int("tm_size", len(resp.ServiceInfos)))
		}
	}()

	return
}

func (tm *taskManager) ID() string {
	return tm.taskManagerInfo.Id
}

func (tm *taskManager) Invoke(ctx context.Context, opTask *pb.OperatorTask, in, out chan *pb.Event) (err error) {

	for {
		select{
		case ev := <- in:
			if ev.EventType == pb.EventType_EtUnknown {
				return
			}

			output := operator.OpFn(opTask.ClientId, opTask.OpId, opTask.OpType, ev.Data)

			if len(opTask.Downstream) != 0 && out != nil {
				out <- &pb.Event{
					Id: 			uuid.NewString(), 
					EventTime: 		ev.EventTime, 
					SourceOpTaskId: opTask.Id,
					TargetOpTaskId: opTask.Downstream[0].Id,
					Data:			output,
				}
			}
			
		case <- ctx.Done():
			return 
		}
	}
	return nil
}