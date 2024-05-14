package task_manager

import (
	"context"
	pb "go-liteflow/pb"
	"net"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type taskManager struct {
	taskManagerId string
	addr          string

	coordinatorId   string
	coordinatorAddr string
	coordinatorConn *grpc.ClientConn

	srv  *grpc.Server
	gSrv *grpcServer

	taskManagerAddrs map[string]*grpc.ClientConn
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
	}

	srv := grpc.NewServer()

	gSrv := &grpcServer{tm: tm}
	pb.RegisterCoreServer(srv, gSrv)

	tm.gSrv = gSrv
	tm.srv = srv

	return tm
}

func (tm *taskManager) Start(ctx context.Context) {

	listener, err := net.Listen("tcp", tm.addr)
	if err != nil {
		panic(err)
	}
	if err = tm.srv.Serve(listener); err != nil {
		panic(err)
	}
}

func (tm *taskManager) InitGrpcClients(ctx context.Context) (err error) {
	conn, err := grpc.Dial(tm.coordinatorAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	tm.coordinatorConn = pb.NewCoreClient(conn)
	heartBeatResp, err := tm.coordinatorConn.SendHeartBeat(context.Background(), 
		&pb.HeartBeatReq{
			TaskManagerId: tm.taskManagerId,
			ServiceStatus: pb.ServiceStatus_SsReady,
	})
	if err != nil {
		return err
	}
	if heartBeatResp == nil || len(heartBeatResp.TaskManagerAddrs) == 0 {
		return nil
	}

	tm.taskManagerAddrs = make(map[string]*grpc.ClientConn)
	for _, tmAddr := range heartBeatResp.TaskManagerAddrs {
		obj, ok := tm.taskManagerAddrs[tmAddr]
		if ok || obj != nil {
			continue
		}
		conn, err := grpc.Dial(tm.coordinatorAddr, grpc.WithInsecure())
		if err != nil {
			return err
		}
		tm.taskManagerAddr[tmAddr] = pb.NewCoreClient(conn)
	}

	return
}