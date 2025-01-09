package task_manager

import (
	"context"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type taskManager struct {
	core.Comm

	// 服务id
	Id string
	// coordinator的id
	CoId string

	srv *grpc.Server

	mux sync.Mutex
	// key: 服务id
	serviceInfos map[string]*core.Service


	digraphMux sync.Mutex
	// key: client_id, val: pb.disgraph
	// taskDigraph map[string]*pb.Digraph
	// key: optask_id, val: pb.OperatorTask
	tasks map[string]*pb.OperatorTask

	chMux sync.Mutex
	// key: optask_id, val: Channel
	//taskChannels map[string]*Channel
}

var tm *taskManager

func GetTaskManager() *taskManager {
	return tm
}

func NewTaskManager(addr, coordAddr string) *taskManager {
	if tm != nil {
		return tm
	}
	if len(coordAddr) == 0 {
		panic("coordinator's address is nil")
	}

	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	tm = &taskManager{
		Id:          			  ranUid.String(),
		serviceInfos:             make(map[string]*core.Service),
		tasks:                    make(map[string]*pb.OperatorTask),
	}

	tm.serviceInfos[tm.Id] = &core.Service{	
		ServiceInfo: pb.ServiceInfo{
			Id:            tm.Id,
			ServiceAddr:   addr,
			ServiceType:   pb.ServiceType_TaskManager,
			ServiceStatus: pb.ServiceStatus_SsRunning,
			Timestamp:     time.Now().Unix()},
	}
	
	coCli, err := core.NewCoreClient(coordAddr)
	if err != nil {
		slog.Error("new coordinator client.", slog.Any("err", err))
		panic(err)
	}	
	tm.serviceInfos[tm.CoId] = &core.Service{
		ServiceInfo: pb.ServiceInfo{
			Id:            tm.CoId,
			ServiceAddr:   coordAddr,
			ServiceType:   pb.ServiceType_Coordinator,
			ServiceStatus: pb.ServiceStatus_SsRunning,
			Timestamp:     time.Now().Unix()},
		ClientConn: coCli,	
	}

	tm.srv = grpc.NewServer()
	pb.RegisterCoreServer(tm.srv, tm)
	return tm
}

func (tm *taskManager) Start(ctx context.Context) {

	err := tm.heartBeat(ctx)
	if err != nil {
		slog.Error("task_manager heart beat.", slog.Any("err", err))
		return
	}

	tm.schedule(ctx)

	listener, err := net.Listen("tcp", tm.SelfServiceInfo().ServiceAddr)
	if err != nil {
		slog.Error("new listener.", slog.Any("err", err))
		panic(err)
	}

	slog.Info("task_manager grpc server start success")
	if err = tm.srv.Serve(listener); err != nil {
		slog.Error("grpc serve fail.", slog.Any("err", err))
		panic(err)
	}
}

func (tm *taskManager) Stop(ctx context.Context) {
	tm.srv.Stop()
}

func (tm *taskManager) Coordinator() pb.CoreClient {
	return tm.serviceInfos[tm.CoId].ClientConn
}

func (tm *taskManager) ID() string {
	return tm.Id
}

func (tm *taskManager) SelfServiceInfo() *core.Service {
	return tm.serviceInfos[tm.Id]
}


func (tm *taskManager) schedule(ctx context.Context) {
	// TODO Currently, task scheduling is not supported. start task by ManageOpTask method

	// schedule ready optask
	// new channel
	// TODO use goroutine pool
	// run goroutine
	// TODO notify optask status
}

