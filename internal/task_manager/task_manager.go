package task_manager

import (
	"context"
	"go-liteflow/internal/core"
	pb "go-liteflow/pb"
	"log/slog"
	"net"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type taskManager struct {
	core.Comm

	taskManagerInfo *pb.ServiceInfo

	coordinatorInfo   *pb.ServiceInfo
	coordinatorClient pb.CoreClient

	srv *grpc.Server

	mux sync.Mutex
	// key: task_manager_id, val: service_addr, info
	serviceInfos map[string]*pb.ServiceInfo
	// key: service_addr, val: grpc_client
	clientConns map[string]pb.CoreClient
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
	ranUid, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	tm = &taskManager{
		taskManagerInfo: &pb.ServiceInfo{
			Id:          ranUid.String(),
			ServiceAddr: addr,
			ServiceType: pb.ServiceType_TaskManager,
		},
		coordinatorInfo: &pb.ServiceInfo{
			ServiceAddr: coordAddr,
		},
		serviceInfos:           make(map[string]*pb.ServiceInfo),
		clientConns:            make(map[string]pb.CoreClient),
		tasks:                  make(map[string]*pb.OperatorTask),
		TaskManangerEventChans: make(map[string]*Channel),
		taskChannels:           make(map[string]*Channel),
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

func (tm *taskManager) Stop(ctx context.Context) {
	tm.srv.Stop()
}

func (tm *taskManager) Coordinator() pb.CoreClient {
	if tm.coordinatorClient != nil {
		return tm.coordinatorClient
	}

	co := tm.coordinatorInfo
	if co == nil || len(co.ServiceAddr) == 0 {
		panic("coordinator info is nil")
	}

	conn, err := grpc.Dial(co.ServiceAddr, grpc.WithInsecure())
	if err != nil {
		panic("connect coordinator fail: " + err.Error())
	}
	tm.coordinatorClient = pb.NewCoreClient(conn)
	return tm.coordinatorClient
}

func (tm *taskManager) ID() string {
	return tm.taskManagerInfo.Id
}

func (tm *taskManager) GetOperatorNodeClient(opId string) pb.Core_EventChannelClient {
	var client pb.Core_EventChannelClient
	var err error
	client = tm.eventChanClient[opId]
	if client == nil && tm.clientConns[opId] != nil {
		client, err = tm.clientConns[opId].EventChannel(context.Background())
		if err != nil {
			slog.Error("Failed to create event client : %v", err)
			return nil
		}
		tm.eventChanClient[opId] = client
	}
	return client
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
