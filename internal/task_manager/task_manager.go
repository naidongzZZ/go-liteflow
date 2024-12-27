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


	// key：task_manager_id, val: RemoteEventChannel
	TaskManangerEventChans map[string]*Channel

	TaskManagerBufferMonitor

	digraphMux sync.Mutex
	// key: client_id, val: pb.disgraph
	// taskDigraph map[string]*pb.Digraph
	// key: optask_id, val: pb.OperatorTask
	tasks map[string]*pb.OperatorTask

	chMux sync.Mutex
	// key: optask_id, val: Channel
	taskChannels map[string]*Channel
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
		TaskManangerEventChans:   make(map[string]*Channel),
		taskChannels:             make(map[string]*Channel),
		TaskManagerBufferMonitor: *NewTaskManagerBufferMonitor(),
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

func (tm *taskManager) GetBuffer(opId string) *Buffer {
	return tm.bufferPool[opId]
}

func (tm *taskManager) schedule(ctx context.Context) {
	// TODO Currently, task scheduling is not supported. start task by ManageOpTask method

	// schedule ready optask
	// new channel
	// TODO use goroutine pool
	// run goroutine
	// TODO notify optask status
}

func (tm *taskManager) RegisterChannel(channels ...*Channel) {
	if len(channels) == 0 {
		return
	}
	for _, ch := range channels {
		if ch.opTaskId != "" {
			tm.chMux.Lock()
			tm.taskChannels[ch.opTaskId] = ch
			tm.chMux.Unlock()
			continue
		}
		if ch.taskManagerId != "" {
			tm.mux.Lock()
			tm.TaskManangerEventChans[ch.taskManagerId] = ch
			tm.mux.Unlock()
			continue
		}
	}
}

func (tm *taskManager) GetChannel(clientId, optaskId string) (ch *Channel) {
	tm.chMux.Lock()
	ch = tm.taskChannels[optaskId]
	tm.chMux.Unlock()
	if ch != nil {
		return
	}

	req := &pb.FindOpTaskReq{ClientId: clientId, OpTaskId: optaskId}
	resp, err := tm.Coordinator().FindOpTask(context.TODO(), req)
	if err != nil {
		slog.Error("find opTask.", slog.Any("err", err))
		return nil
	}
	if resp == nil || resp.TaskManagerId == "" {
		return nil
	}
	tm.mux.Lock()
	defer tm.mux.Unlock()
	ch = tm.TaskManangerEventChans[resp.TaskManagerId]
	return
}
